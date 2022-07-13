package azure

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	az "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

type item struct {
	id         string
	container  *container
	client     *az.ServiceClient
	properties *stowAzureProperties
	url        url.URL
	metadata   map[string]interface{}
	infoOnce   sync.Once
	infoErr    error
}

//stowAzureProperties Azure gives a different set of properties depending on the call.  We're encapsulating the data
// we're interested in into a single struct
type stowAzureProperties struct {
	ETag          *string
	ContentLength *int64
	LastModified  *time.Time
}

func newStowProperties(etag string, size int64, lastModified time.Time) *stowAzureProperties {
	return &stowAzureProperties{
		ETag:          &etag,
		ContentLength: &size,
		LastModified:  &lastModified,
	}

}

var (
	_ stow.Item       = (*item)(nil)
	_ stow.ItemRanger = (*item)(nil)
)

func (i *item) ID() string {
	return i.id
}

func (i *item) Name() string {
	return i.id
}

func (i *item) URL() *url.URL {
	blobClient, err := getBlobClient(i.client, i.container.id, i.Name())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not find item: %s", i.id)
		return nil
	}

	u := blobClient.URL()
	url, err := url.Parse(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to parse URL: %v", u)
		return nil
	}
	url.Scheme = "azure"
	return url
}

func (i *item) Size() (int64, error) {
	return *i.properties.ContentLength, nil
	//if i.properties != nil {
	//	return *i.properties.ContentLength, nil
	//} else if i.containerProperties.ContentLength != nil {
	//	return *i.containerProperties.ContentLength, nil
	//}
	//return 0, fmt.Errorf("unable to get item properties")
}

func (i *item) Open() (io.ReadCloser, error) {
	blobClient, err := getBlobClient(i.client, i.container.id, i.Name())
	if err != nil {
		return nil, err
	}
	resp, err := blobClient.Download(context.Background(), nil)
	return resp.Body(nil), nil
}

func (i *item) ETag() (string, error) {
	return *i.properties.ETag, nil
}

func (i *item) LastMod() (time.Time, error) {
	return *i.properties.LastModified, nil
}

func (i *item) Metadata() (map[string]interface{}, error) {
	err := i.ensureInfo()
	if err != nil {
		return nil, errors.Wrap(err, "retrieving metadata")
	}

	return i.metadata, nil
}

func (i *item) ensureInfo() error {
	if i.metadata == nil {
		i.infoOnce.Do(func() {

			blobClient, err := getBlobClient(i.client, i.container.id, i.Name())
			v := new(az.BlobGetPropertiesOptions)
			properties, err := blobClient.GetProperties(context.Background(), v)
			if err != nil {
				return
			}
			mdParsed, infoErr := parseMetadata(properties.Metadata)
			mdParsed = fixAzureMetadataBug(mdParsed)

			if infoErr != nil {
				i.infoErr = infoErr
				return
			}
			i.metadata = mdParsed
		})
	}

	return i.infoErr
}

//fixAzureMetadataBug: Fix Capitalization: https://github.com/Azure/azure-sdk-for-go/issues/17850
func fixAzureMetadataBug(mdParsed map[string]interface{}) map[string]interface{} {
	fixedParsed := make(map[string]interface{}, len(mdParsed))
	for key, value := range mdParsed {
		if unicode.IsUpper(rune(key[0])) {
			fixedParsed[strings.ToLower(key)] = value
		} else {
			fixedParsed[key] = value
		}
	}
	return fixedParsed
}

func (i *item) getInfo() (stow.Item, error) {
	itemInfo, err := i.container.Item(i.ID())
	if err != nil {
		return nil, err
	}
	return itemInfo, nil
}

// OpenRange opens the item for reading starting at byte start and ending
// at byte end.
func (i *item) OpenRange(start, end uint64) (io.ReadCloser, error) {
	blobClient, err := getBlobClient(i.client, i.container.id, i.Name())
	if err != nil {
		return nil, err
	}
	startParam := int64(start)
	var count int64 = (int64(end) - startParam) + 1
	options := az.BlobDownloadOptions{
		Offset: &startParam,
		Count:  &count,
	}
	resp, err := blobClient.Download(context.Background(), &options)
	if err != nil {
		return nil, err
	}
	return resp.Body(nil), nil
}

//Helper Utilities
func getBlobClient(client *az.ServiceClient, containerId, blobName string) (*az.BlockBlobClient, error) {
	containerObj, err := client.NewContainerClient(containerId)
	if err != nil {
		return nil, err
	}
	return containerObj.NewBlockBlobClient(blobName)

}
