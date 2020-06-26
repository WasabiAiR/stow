package readonly

import (
	"io"
	"net/url"
	"time"

	"github.com/graymeta/stow"
)

type item struct {
	wrapped stow.Item
}

// ID gets a unique string describing this Item.
func (i *item) ID() string {
	return i.wrapped.ID()
}

// Name gets a human-readable name describing this Item.
func (i *item) Name() string {
	return i.wrapped.Name()
}

// URL gets a URL for this item.
// For example:
// local: file:///path/to/something
// azure: azure://host:port/api/something
//    s3: s3://host:post/etc
func (i *item) URL() *url.URL {
	return i.wrapped.URL()
}

// Size gets the size of the Item's contents in bytes.
func (i *item) Size() (int64, error) {
	return i.wrapped.Size()
}

// Open opens the Item for reading.
// Calling code must close the io.ReadCloser.
func (i *item) Open() (io.ReadCloser, error) {
	return i.wrapped.Open()
}

// ETag is a string that is different when the Item is
// different, and the same when the item is the same.
// Usually this is the last modified datetime.
func (i *item) ETag() (string, error) {
	return i.wrapped.ETag()
}

// LastMod returns the last modified date of the file.
func (i *item) LastMod() (time.Time, error) {
	return i.wrapped.LastMod()
}

// Metadata gets a map of key/values that belong
// to this Item.
func (i *item) Metadata() (map[string]interface{}, error) {
	return i.wrapped.Metadata()
}
