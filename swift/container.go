package swift

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/flyteorg/stow"
	"github.com/ncw/swift/v2"
)

type container struct {
	id     string
	client *swift.Connection
}

var (
	_ stow.Container        = (*container)(nil)
	_ stow.ContextContainer = (*container)(nil)
)

func (c *container) ID() string {
	return c.id
}

func (c *container) Name() string {
	return c.id
}

func (c *container) PreSignRequest(_ context.Context, _ stow.ClientMethod, _ string,
	_ stow.PresignRequestParams) (response stow.PresignResponse, err error) {
	return stow.PresignResponse{}, fmt.Errorf("unsupported")
}

func (c *container) Item(id string) (stow.Item, error) {
	return c.ItemContext(context.Background(), id)
}

// ItemContext is Item with a context.
func (c *container) ItemContext(ctx context.Context, id string) (stow.Item, error) {
	return c.getItem(ctx, id)
}

func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return c.ItemsContext(context.Background(), prefix, cursor, count)
}

// ItemsContext is Items with a context.
func (c *container) ItemsContext(ctx context.Context, prefix, cursor string, count int) ([]stow.Item, string, error) {
	params := &swift.ObjectsOpts{
		Limit:  count,
		Marker: cursor,
		Prefix: prefix,
	}
	objects, err := c.client.Objects(ctx, c.id, params)
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

func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]any) (stow.Item, error) {
	return c.PutContext(context.Background(), name, r, size, metadata)
}

// PutContext is Put with a context.
func (c *container) PutContext(ctx context.Context, name string, r io.Reader, size int64, metadata map[string]any) (stow.Item, error) {
	mdPrepped, err := prepMetadata(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to create or update Item, preparing metadata: %w", err)
	}

	headers, err := c.client.ObjectPut(ctx, c.id, name, r, false, "", "", mdPrepped)
	if err != nil {
		return nil, fmt.Errorf("unable to create or update Item: %w", err)
	}

	mdParsed, err := parseMetadata(headers)
	if err != nil {
		return nil, fmt.Errorf("unable to create or update Item, parsing metadata: %w", err)
	}

	item := &item{
		id:        name,
		container: c,
		client:    c.client,
		size:      size,
		metadata:  mdParsed,
	}
	return item, nil
}

func (c *container) RemoveItem(id string) error {
	return c.RemoveItemContext(context.Background(), id)
}

// RemoveItemContext is RemoveItem with a context.
func (c *container) RemoveItemContext(ctx context.Context, id string) error {
	return c.client.ObjectDelete(ctx, c.id, id)
}

func (c *container) getItem(ctx context.Context, id string) (*item, error) {
	info, headers, err := c.client.Object(ctx, c.id, id)
	if err != nil {
		if strings.Contains(err.Error(), "Object Not Found") {
			return nil, stow.ErrNotFound
		}
		return nil, fmt.Errorf("error retrieving item: %w", err)
	}

	md, err := parseMetadata(headers)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Item information, parsing metadata: %w", err)
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

// Keys are returned as all lowercase, dashes are allowed
func parseMetadata(md swift.Headers) (map[string]any, error) {
	m := make(map[string]any, len(md))
	for key, value := range md.ObjectMetadata() {
		m[key] = value
	}
	return m, nil
}

// TODO determine invalid keys.
func prepMetadata(md map[string]any) (map[string]string, error) {
	m := make(map[string]string, len(md))
	for key, value := range md {
		str, ok := value.(string)
		if !ok {
			return nil, errors.New("could not convert key value") // add a msg mentioning strings only?
		}
		m["X-Object-Meta-"+key] = str
	}
	return m, nil
}
