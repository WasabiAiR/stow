package oracle

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

// ID returns a string value representing the Item, in this case it's the
// name of the object.
func (i *item) ID() string {
	return i.id
}

// Name returns a string value representing the Item, in this case it's the
// name of the object.
func (i *item) Name() string {
	return i.id
}

// URL returns a URL that for the given CloudStorage object.
func (i *item) URL() *url.URL {
	url, _ := url.Parse(i.client.StorageUrl)
	url.Scheme = Kind
	url.Path = path.Join(url.Path, i.container.id, i.id)
	return url
}

// Size returns the size in bytes of the CloudStorage object.
func (i *item) Size() (int64, error) {
	return i.size, nil
}

// Open is a method that returns an io.ReadCloser which represents the content
// of the CloudStorage object.
func (i *item) Open() (io.ReadCloser, error) {
	return i.OpenContext(context.Background())
}

// OpenContext is Open with a context.
func (i *item) OpenContext(ctx context.Context) (io.ReadCloser, error) {
	r, _, err := i.client.ObjectOpen(ctx, i.container.id, i.id, false, nil)
	var res io.ReadCloser = r
	// FIXME: this is a workaround to issue https://github.com/graymeta/stow/issues/120
	if s, ok := res.(readSeekCloser); ok {
		res = &fixReadSeekCloser{readSeekCloser: s, item: i}
	}
	return res, err
}

type readSeekCloser interface {
	io.ReadSeeker
	io.Closer
}

type fixReadSeekCloser struct {
	readSeekCloser
	item *item
	read bool
}

func (f *fixReadSeekCloser) Read(p []byte) (int, error) {
	f.read = true
	return f.readSeekCloser.Read(p)
}

func (f *fixReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekEnd && !f.read {
		return f.item.size, nil
	}
	return f.readSeekCloser.Seek(offset, whence)
}

// ETag returns a string value representing the CloudStorage Object
func (i *item) ETag() (string, error) {
	return i.ETagContext(context.Background())
}

// ETagContext is ETag with a context.
func (i *item) ETagContext(_ context.Context) (string, error) {
	return i.hash, nil
}

// LastMod returns a time.Time object representing information on the date
// of the last time the CloudStorage object was modified.
func (i *item) LastMod() (time.Time, error) {
	return i.LastModContext(context.Background())
}

// LastModContext is LastMod with a context.
func (i *item) LastModContext(ctx context.Context) (time.Time, error) {
	// If an object is PUT, certain information is missing. Detect
	// if the lastModified field is missing, send a request to retrieve
	// it, and save both this and other missing information so that a
	// request doesn't have to be sent again.
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
