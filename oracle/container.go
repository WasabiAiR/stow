package swift

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/graymeta/stow"
	"github.com/ncw/swift"
	"github.com/pkg/errors"
)

type container struct {
	id     string
	client *swift.Connection
}

var _ stow.Container = (*container)(nil)

// ID returns a string value representing a unique container, in this case it's
// the Container's name.
func (c *container) ID() string {
	return c.id
}

// Name returns a string value representing a unique container, in this case
// it's the Container's name.
func (c *container) Name() string {
	return c.id
}

func (c *container) Item(id string) (stow.Item, error) {
	return c.getItem(id)
}

func (c *container) Browse(prefix, delimiter, cursor string, count int) (*stow.ItemPage, error) {
	params := &swift.ObjectsOpts{
		Limit:  count,
		Marker: cursor,
		Prefix: prefix,
	}
	r, sz := utf8.DecodeRuneInString(delimiter)
	if r == utf8.RuneError {
		if sz > 0 {
			return nil, fmt.Errorf("Bad delimiter %v", delimiter)
		}
	} else {
		params.Delimiter = r
	}
	objects, err := c.client.Objects(c.id, params)
	if err != nil {
		return nil, err
	}

	var prefixes []string
	for _, obj := range objects {
		if obj.PseudoDirectory {
			prefixes = append(prefixes, obj.Name)
		}
	}

	var items []stow.Item
	for _, obj := range objects {
		if obj.PseudoDirectory {
			continue
		}
		items = append(items, &item{
			id:           obj.Name,
			container:    c,
			client:       c.client,
			hash:         obj.Hash,
			size:         obj.Bytes,
			lastModified: obj.LastModified,
		})
	}
	marker := ""
	if len(objects) == count {
		marker = objects[len(objects)-1].Name
	}
	return &stow.ItemPage{Prefixes: prefixes, Items: items, Cursor: marker}, nil
}

// Items returns a collection of CloudStorage objects based on a matching
// prefix string and cursor information.
func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	page, err := c.Browse(prefix, "", cursor, count)
	if err != nil {
		return nil, "", err
	}
	return page.Items, cursor, err
}

// Put creates or updates a CloudStorage object within the given container.
func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	mdPrepped, err := prepMetadata(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create or update Item, preparing metadata")
	}

	_, err = c.client.ObjectPut(c.id, name, r, false, "", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create or update Item")
	}

	err = c.client.ObjectUpdate(c.id, name, mdPrepped)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update Item metadata")
	}

	item := &item{
		id:        name,
		container: c,
		client:    c.client,
		size:      size,
		// not setting metadata here, the refined version isn't available
		// unless an explicit getItem() is done. Possible to write a func to facilitate
		// this.
	}
	return item, nil
}

// RemoveItem removes a CloudStorage object located within the given
// container.
func (c *container) RemoveItem(id string) error {
	return c.client.ObjectDelete(c.id, id)
}

func (c *container) getItem(id string) (*item, error) {
	info, headers, err := c.client.Object(c.id, id)
	if err != nil {
		if strings.Contains(err.Error(), "Object Not Found") {
			return nil, stow.ErrNotFound
		}
		return nil, err
	}

	md, err := parseMetadata(headers)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve Item information, parsing metadata")
	}

	item := &item{
		id:           id,
		container:    c,
		client:       c.client,
		hash:         info.Hash,
		size:         info.Bytes,
		lastModified: info.LastModified,
		metadata:     md,
	}

	return item, nil
}

// Keys are returned as all lowercase
func parseMetadata(md swift.Headers) (map[string]interface{}, error) {
	m := make(map[string]interface{}, len(md))
	for key, value := range md.ObjectMetadata() {
		m[key] = value
	}
	return m, nil
}

func prepMetadata(md map[string]interface{}) (map[string]string, error) {
	m := make(map[string]string, len(md))
	for key, value := range md {
		str, ok := value.(string)
		if !ok {
			return nil, errors.Errorf(`value of key '%s' in metadata must be of type string`, key)
		}
		m["X-Object-Meta-"+key] = str
	}
	return m, nil
}
