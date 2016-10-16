package google

import (
	"io"
	"net/url"

	//	"strings"
	"time"

	storage "google.golang.org/api/storage/v1"
)

type item struct {
	container    *container       // Container information is required by a few methods.
	client       *storage.Service // A client is needed to make requests.
	name         string
	hash         string
	etag         string
	size         int64
	url          *url.URL
	lastModified time.Time
	metadata     map[string]interface{}
}

// ID returns a string value that represents the name of a file.
func (i *item) ID() string {
	return i.name
}

// Name returns a string value that represents the name of the file.
func (i *item) Name() string {
	return i.name
}

// Size returns the size of an item in bytes.
func (i *item) Size() (int64, error) {
	return i.size, nil
}

// URL returns a url which follows the predefined format
func (i *item) URL() *url.URL {
	return i.url
}

// Open returns an io.ReadCloser to the object. Useful for downloading/streaming the object.
func (i *item) Open() (io.ReadCloser, error) {
	res, err := i.client.Objects.Get(i.container.name, i.name).Download()
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

// LastMod returns the last modified date of the item.
func (i *item) LastMod() (time.Time, error) {
	return i.lastModified, nil
}

// Metadata returns a nil map and no error.
// TODO: Implement this.
func (i *item) Metadata() (map[string]interface{}, error) {
	return i.metadata, nil
}

// ETag returns the ETag value
func (i *item) ETag() (string, error) {
	return i.etag, nil
}

// prepUrl takes a MediaLink string and returns a url
func prepUrl(str string) (*url.URL, error) {
	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	u.Scheme = "google"

	// Discard the query string
	u.RawQuery = ""
	return u, nil
}
