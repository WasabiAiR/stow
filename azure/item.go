package azure

import (
	"io"
	"net/url"
	"sync"

	az "github.com/Azure/azure-sdk-for-go/storage"
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

func (i *item) ID() string {
	panic("not implemented")
}

func (i *item) Name() string {
	panic("not implemented")
}

func (i *item) URL() *url.URL {
	panic("not implemented")
}

func (i *item) Open() (io.ReadCloser, error) {
	panic("not implemented")
}

func (i *item) ETag() (string, error) {
	panic("not implemented")
}

func (i *item) MD5() (string, error) {
	panic("not implemented")
}
