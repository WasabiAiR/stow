package swift

import (
	"io"
	"net/url"
	"path"
	"time"

	"github.com/graymeta/stow"
	"github.com/ncw/swift"
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
}

var _ stow.Item = (*item)(nil)

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
	r, _, err := i.client.ObjectOpen(i.container.id, i.id, false, nil)
	return r, err
}

// ETag returns a string value representing the CloudStorage Object
func (i *item) ETag() (string, error) {
	return i.hash, nil
}

// LastMod returns a time.Time object representing information on the date
// of the last time the CloudStorage object was modified.
func (i *item) LastMod() (time.Time, error) {
	// If an object is PUT, certain information is missing. Detect
	// if the lastModified field is missing, send a request to retrieve
	// it, and save both this and other missing information so that a
	// request doesn't have to be sent again. Could be placed in PUT,
	// but right now it seems cleaner to have a request sent when this
	// field is needed for a maximimum of a single request, rather than
	// sending a request to get the missing info every time an object
	// is PUT.
	if i.lastModified.IsZero() {
		itemInfo, err := i.container.getItem(i.ID())
		if err != nil {
			return time.Time{}, err
		}
		// Save the missing information so that a request won't need to be
		// sent again.
		i.lastModified = itemInfo.lastModified
		i.hash = itemInfo.hash
	}

	return i.lastModified, nil
}

// Metadata returns a nil map and no error.
// TODO: Implement this.
func (i *item) Metadata() (map[string]interface{}, error) {
	return nil, nil
}
