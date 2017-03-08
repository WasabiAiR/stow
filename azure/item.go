package azure

import (
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	az "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

type item struct {
	id         string
	container  *container
	client     *az.BlobStorageClient
	properties az.BlobProperties
	url        url.URL
	metadata   map[string]interface{}
	infoOnce   sync.Once
	infoErr    error
}

var _ stow.Item = (*item)(nil)

func (i *item) ID() string {
	return i.id
}

func (i *item) Name() string {
	return i.id
}

func (i *item) URL() *url.URL {
	u := i.client.GetBlobURL(i.container.id, i.id)
	url, _ := url.Parse(u)
	url.Scheme = "azure"
	return url
}

func (i *item) Size() (int64, error) {
	return i.properties.ContentLength, nil
}

func (i *item) Open() (io.ReadCloser, error) {
	return i.client.GetBlob(i.container.id, i.id)
}

func (i *item) Partial(length, offset int64) (io.ReadCloser, error) {
	// https://docs.microsoft.com/en-us/rest/api/storageservices/fileservices/get-blob
	if offset < 0 {
		return nil, errors.New("offset is negative")
	}
	if length < 0 {
		return nil, fmt.Errorf("invalid length %d", length)
	}
	var r string
	if length > 0 {
		r = fmt.Sprintf("%d-%d", offset, offset+length)
	} else {
		r = fmt.Sprintf("%d-", offset)
	}
	return i.client.GetBlobRange(i.container.id, i.id, r, nil)
}

func (i *item) ETag() (string, error) {
	return i.properties.Etag, nil
}

func (i *item) LastMod() (time.Time, error) {
	return time.Parse(timeFormat, i.properties.LastModified)
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
			md, infoErr := i.client.GetBlobMetadata(i.container.Name(), i.Name())
			if infoErr != nil {
				i.infoErr = infoErr
				return
			}

			mdParsed, infoErr := parseMetadata(md)
			if infoErr != nil {
				i.infoErr = infoErr
				return
			}
			i.metadata = mdParsed
		})
	}

	return i.infoErr
}

func (i *item) getInfo() (stow.Item, error) {
	itemInfo, err := i.container.Item(i.ID())
	if err != nil {
		return nil, err
	}
	return itemInfo, nil
}
