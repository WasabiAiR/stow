package google

import (
	"io"
	"time"

	"github.com/graymeta/stow"
	storage "google.golang.org/api/storage/v1"
)

type container struct {
	// Name is needed to retrieve items.
	name string

	// Client is responsible for performing the requests.
	client *storage.Service
}

// ID returns a string value which represents the name of the container.
func (c *container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *container) Name() string {
	return c.name
}

// Item returns a stow.Item instance of a container based on the
// name of the container
func (c *container) Item(id string) (stow.Item, error) {

	res, err := c.client.Objects.Get(c.name, id).Do()
	if err != nil {
		return nil, stow.ErrNotFound
	}

	t, err := time.Parse(time.RFC3339, res.Updated)
	if err != nil {
		return nil, err
	}

	u, err := prepUrl(res.MediaLink)
	if err != nil {
		return nil, err
	}

	i := &item{
		name:         id,
		container:    c,
		client:       c.client,
		size:         int64(res.Size),
		etag:         res.Etag,
		hash:         res.Md5Hash,
		lastModified: t,
		url:          u,
	}

	return i, nil
}

// Items retrieves a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *container) Items(prefix string, cursor string) ([]stow.Item, string, error) {
	// List all objects in a bucket using pagination
	call := c.client.Objects.List(c.name).MaxResults(10)

	if prefix != "" {
		call.Prefix(prefix)
	}

	if cursor != "" {
		call = call.PageToken(cursor)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	containerItems := make([]stow.Item, len(res.Items))

	for i, o := range res.Items {
		t, err := time.Parse(time.RFC3339, o.Updated)
		if err != nil {
			return nil, "", err
		}

		u, err := prepUrl(o.MediaLink)
		if err != nil {
			return nil, "", err
		}

		containerItems[i] = &item{
			name:         o.Name,
			container:    c,
			client:       c.client,
			size:         int64(o.Size),
			etag:         o.Etag,
			hash:         o.Md5Hash,
			lastModified: t,
			url:          u,
		}
	}

	return containerItems, res.NextPageToken, nil
}

func (c *container) RemoveItem(id string) error {
	return c.client.Objects.Delete(c.name, id).Do()
}

// Put sends a request to upload content to the container. The arguments
// received are the name of the item, a reader representing the
// content, and the size of the file.
func (c *container) Put(name string, r io.Reader, size int64) (stow.Item, error) {

	object := &storage.Object{Name: name}
	res, err := c.client.Objects.Insert(c.name, object).Media(r).Do()

	if err != nil {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, res.Updated)
	if err != nil {
		return nil, err
	}

	u, err := prepUrl(res.MediaLink)
	if err != nil {
		return nil, err
	}

	newItem := &item{
		name:         name,
		container:    c,
		client:       c.client,
		size:         size,
		etag:         res.Etag,
		hash:         res.Md5Hash,
		lastModified: t,
		url:          u,
	}

	return newItem, nil
}
