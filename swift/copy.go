package swift

import (
	"context"
	"fmt"

	"github.com/flyteorg/stow"
)

var _ stow.Copier = (*container)(nil)

// Copy copies src into the container on the server side, keeping the metadata
// of the source object. An item from another kind of location is streamed
// instead.
func (c *container) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	srcItem, ok := src.(*item)
	if !ok {
		return stow.StreamCopy(ctx, c, src, name)
	}

	if _, err := c.client.ObjectCopy(ctx, srcItem.container.id, srcItem.id, c.id, name, nil); err != nil {
		return nil, fmt.Errorf("copy, copying the object: %w", err)
	}

	return c.getItem(ctx, name)
}
