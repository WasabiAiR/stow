package swift

import (
	"context"
	"io"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/flyteorg/stow"
	"github.com/ncw/swift/v2"
)

type item struct {
	id        string
	container *container
	client    *swift.Connection
	//properties az.BlobProperties
	hash         string
	size         int64
	url          url.URL
	lastModified time.Time
	metadata     map[string]any
	infoMu       sync.Mutex
	infoLoaded   bool
}

var (
	_ stow.Item        = (*item)(nil)
	_ stow.ContextItem = (*item)(nil)
)

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
	url.Path = path.Join(url.Path, i.container.id, i.id)
	return url
}

func (i *item) Size() (int64, error) {
	return i.size, nil
}

func (i *item) Open() (io.ReadCloser, error) {
	return i.OpenContext(context.Background())
}

// OpenContext is Open with a context.
func (i *item) OpenContext(ctx context.Context) (io.ReadCloser, error) {
	r, _, err := i.client.ObjectOpen(ctx, i.container.id, i.id, false, nil)
	return r, err
}

func (i *item) ETag() (string, error) {
	return i.ETagContext(context.Background())
}

// ETagContext is ETag with a context.
func (i *item) ETagContext(ctx context.Context) (string, error) {
	err := i.ensureInfo(ctx)
	if err != nil {
		return "", err
	}
	return i.hash, nil
}

func (i *item) LastMod() (time.Time, error) {
	return i.LastModContext(context.Background())
}

// LastModContext is LastMod with a context.
func (i *item) LastModContext(ctx context.Context) (time.Time, error) {
	err := i.ensureInfo(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return i.lastModified, nil
}

// Metadata returns a map of key value pairs representing an Item's metadata
func (i *item) Metadata() (map[string]any, error) {
	return i.MetadataContext(context.Background())
}

// MetadataContext is Metadata with a context.
func (i *item) MetadataContext(ctx context.Context) (map[string]any, error) {
	err := i.ensureInfo(ctx)
	if err != nil {
		return nil, err
	}
	return i.metadata, nil
}

// ensureInfo checks the fields that may be empty when an item is PUT.
// Verify if the fields are empty, get information on the item, fill in
// the missing fields.
func (i *item) ensureInfo(ctx context.Context) error {
	i.infoMu.Lock()
	defer i.infoMu.Unlock()

	// If lastModified is empty, so is hash. get info on the Item and
	// update the necessary fields at the same time.
	if i.infoLoaded || (!i.lastModified.IsZero() && i.hash != "" && i.metadata != nil) {
		return nil
	}

	itemInfo, err := i.container.getItem(ctx, i.ID())
	if err != nil {
		return err
	}
	i.hash = itemInfo.hash
	i.lastModified = itemInfo.lastModified
	i.metadata = itemInfo.metadata
	i.infoLoaded = true
	return nil
}
