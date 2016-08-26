package azure

import (
	"io"
	"net/url"
	"time"

	az "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/graymeta/stow"
)

type item struct {
	id         string
	container  *container
	client     *az.BlobStorageClient
	properties az.BlobProperties
	url        url.URL
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

func (i *item) ETag() (string, error) {
	return i.properties.Etag, nil
}

func (i *item) LastMod() (time.Time, error) {
	return time.Parse(timeFormat, i.properties.LastModified)
}

// Metadata returns a nil map and no error.
// TODO: Implement this.
func (i *item) Metadata() (map[string]interface{}, error) {
	return nil, nil
}
