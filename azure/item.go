package azure

import (
	"io"
	"net/url"
	"sync"

	az "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/graymeta/stow"
)

type item struct {
	id         string
	container  *container
	client     *az.BlobStorageClient
	blob       *az.Blob
	properties *az.BlobProperties
	page       int

	rc       io.ReadCloser
	readOnce sync.Once
}

var _ stow.Item = (*item)(nil)

func (i *item) ID() string {
	return i.id
}

func (i *item) Name() string {
	return i.id
}

func (i *item) URL() *url.URL {
	u := i.client.GetBlobURL(i.container.id, i.blob.Name)
	url, _ := url.Parse(u)
	url.Scheme = "azure"
	return url
}

func (i *item) Open() (io.ReadCloser, error) {
	return i.client.GetBlob(i.container.id, i.id)
}

func (i *item) ETag() (string, error) {
	return i.properties.Etag, nil
}

func (i *item) MD5() (string, error) {
	return i.properties.ContentMD5, nil
}
