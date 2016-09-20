package swift

import (
	"io"
	"strings"

	"github.com/graymeta/stow"
	"github.com/ncw/swift"
)

type container struct {
	id     string
	client *swift.Connection
}

var _ stow.Container = (*container)(nil)

func (c *container) ID() string {
	return c.id
}

func (c *container) Name() string {
	return c.id
}

func (c *container) Item(id string) (stow.Item, error) {
	return c.getItem(id)
}

func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	params := &swift.ObjectsOpts{
		Limit:  count,
		Marker: cursor,
		Prefix: prefix,
	}
	objects, err := c.client.Objects(c.id, params)
	if err != nil {
		return nil, "", err
	}
	items := make([]stow.Item, len(objects))
	for i, obj := range objects {

		items[i] = &item{
			id:           obj.Name,
			container:    c,
			client:       c.client,
			hash:         obj.Hash,
			size:         obj.Bytes,
			lastModified: obj.LastModified,
		}
	}
	marker := ""
	if len(objects) == count {
		marker = objects[len(objects)-1].Name
	}
	return items, marker, nil
}

func (c *container) Put(name string, r io.Reader, size int64) (stow.Item, error) {
	_, err := c.client.ObjectPut(c.id, name, r, false, "", "", swift.Headers{})
	if err != nil {
		return nil, err
	}
	item := &item{
		id:        name,
		container: c,
		client:    c.client,
		size:      size,
	}
	return item, nil
}

func (c *container) RemoveItem(id string) error {
	return c.client.ObjectDelete(c.id, id)
}

func (c *container) getItem(id string) (*item, error) {
	info, _, err := c.client.Object(c.id, id)
	if err != nil {
		if strings.Contains(err.Error(), "Object Not Found") {
			return nil, stow.ErrNotFound
		}
		return nil, err
	}
	item := &item{
		id:           id,
		container:    c,
		client:       c.client,
		hash:         info.Hash,
		size:         info.Bytes,
		lastModified: info.LastModified,
	}
	return item, nil
}
