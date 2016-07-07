package azure

import (
	"io"
	"net/url"
	"sync"

	"encoding/base64"

	"encoding/hex"

	az "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/graymeta/stow"
)

type item struct {
	id         string
	container  *container
	client     *az.BlobStorageClient
	properties az.BlobProperties
	page       int
	url        url.URL

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
	u := i.client.GetBlobURL(i.container.id, i.id)
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
	decoded, err := base64.StdEncoding.DecodeString(i.properties.ContentMD5)
	if err != nil {
		return "", err
	}
	str := hex.EncodeToString(decoded)
	return str, nil
}
