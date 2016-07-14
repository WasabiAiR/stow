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
	info, _, err := c.client.Object(c.id, id)

	if err != nil {
		if strings.Contains(err.Error(), "Object Not Found") {
			return nil, stow.ErrNotFound
		}
		return nil, err
	}

	item := &item{
		id:        id,
		container: c,
		client:    c.client,
		hash:      info.Hash,
	}
	return item, nil
}

func (c *container) Items(prefix, cursor string) ([]stow.Item, string, error) {

	numitems := 10

	params := &swift.ObjectsOpts{
		Limit:  numitems,
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
			id:        obj.Name,
			container: c,
			client:    c.client,
			hash:      obj.Hash,
		}
	}

	marker := ""

	if len(objects) == numitems {
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
	}
	return item, nil
}

func (c *container) RemoveItem(id string) error {
	return c.client.ObjectDelete(c.id, id)
}
