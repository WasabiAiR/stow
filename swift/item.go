package swift

import (
	"fmt"
	"io"
	"net/url"

	"github.com/graymeta/stow"
	"github.com/ncw/swift"
)

type item struct {
	id        string
	container *container
	client    *swift.Connection
	//properties az.BlobProperties
	hash string
	size int64
	url  url.URL
}

var _ stow.Item = (*item)(nil)

func (i *item) ID() string {
	return i.id
}

func (i *item) Name() string {
	return i.id
}

func (i *item) URL() *url.URL {

	// StorageUrl looks like this:
	// https://lax-proxy-03.storagesvc.sohonet.com/v1/AUTH_b04239c7467548678b4822e9dad96030
	// We want something like this:
	// swift://lax-proxy-03.storagesvc.sohonet.com/v1/AUTH_b04239c7467548678b4822e9dad96030/<container_name>/<path_to_object>

	url, _ := url.Parse(i.client.StorageUrl)
	url.Scheme = Kind
	url.Path = fmt.Sprintf("%s/%s/%s", url.Path, i.container.id, i.id)

	return url
}

func (i *item) Size() (int64, error) {
	return i.size, nil
}

func (i *item) Open() (io.ReadCloser, error) {
	r, _, err := i.client.ObjectOpen(i.container.id, i.id, false, nil)
	return r, err
}

func (i *item) ETag() (string, error) {
	return i.hash, nil
}

func (i *item) MD5() (string, error) {
	return i.hash, nil
}
