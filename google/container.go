package google

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

type Container struct {
	// Name is needed to retrieve items.
	name string

	//bucket *storage.BucketHandle

	// Client is responsible for performing the requests.
	client *storage.Client
}

// ID returns a string value which represents the name of the container.
func (c *Container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *Container) Name() string {
	return c.name
}

func (c *Container) Bucket() *storage.BucketHandle {
	return c.client.Bucket(c.name)
}

func (c *Container) objectAttrToItem(res *storage.ObjectAttrs) *Item {
	u, err := prepUrl(res.MediaLink)
	if err != nil {
		return nil
	}

	mdParsed, err := parseMetadata(res.Metadata)
	if err != nil {
		return nil
	}

	return &Item{
		name:         res.Name,
		container:    c,
		client:       c.client,
		size:         res.Size,
		etag:         "",
		hash:         string(res.MD5),
		lastModified: res.Updated,
		url:          u,
		metadata:     mdParsed,
	}
}

// Item returns a stow.Item instance of a container based on the
// name of the container
func (c *Container) Item(id string) (stow.Item, error) {
	obj := c.Bucket().Object(id)
	res, err := obj.Attrs(context.TODO())
	if err != nil {
		return nil, stow.ErrNotFound
	}

	itm := c.objectAttrToItem(res)
	itm.object = obj
	return itm, nil
}

// Items retrieves a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *Container) Items(prefix string, cursor string, count int) ([]stow.Item, string, error) {
	// List all objects in a bucket using pagination
	q := storage.Query{
		Prefix: prefix,
	}

	pager := iterator.NewPager(c.Bucket().Objects(context.Background(), &q), count, cursor)

	var attrs []*storage.ObjectAttrs
	nextPageToken, err := pager.NextPage(&attrs)
	if err != nil {
		return nil, "", err
	}
	containerItems := make([]stow.Item, len(attrs))

	for i, o := range attrs {
		containerItems[i] = c.objectAttrToItem(o)
	}

	return containerItems, nextPageToken, nil
}

func (c *Container) RemoveItem(id string) error {
	return c.Bucket().Object(id).Delete(context.Background())
}

// Put sends a request to upload content to the container. The arguments
// received are the name of the item, a reader representing the
// content, and the size of the file.
func (c *Container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	mdPrepped, err := prepMetadata(metadata)
	if err != nil {
		return nil, err
	}

	obj := c.Bucket().Object(name)
	writer := obj.NewWriter(context.Background())

	writer.Name = name
	writer.Metadata = mdPrepped

	_, err = io.Copy(writer, r)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return c.Item(name)
}

func parseMetadata(metadataParsed map[string]string) (map[string]interface{}, error) {
	metadataParsedMap := make(map[string]interface{}, len(metadataParsed))
	for key, value := range metadataParsed {
		metadataParsedMap[key] = value
	}
	return metadataParsedMap, nil
}

func prepMetadata(metadataParsed map[string]interface{}) (map[string]string, error) {
	returnMap := make(map[string]string, len(metadataParsed))
	for key, value := range metadataParsed {
		str, ok := value.(string)
		if !ok {
			return nil, errors.Errorf(`value of key '%s' in metadata must be of type string`, key)
		}
		returnMap[key] = str
	}
	return returnMap, nil
}
