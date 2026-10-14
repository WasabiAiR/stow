package local

import (
	"context"
	"errors"
	"fmt"

	"github.com/flyteorg/stow"
)

var _ stow.Copier = (*container)(nil)

// Copy copies src into the container by streaming its content into a new
// file. Local items carry no user metadata, so none is copied.
func (c *container) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	if src == nil {
		return nil, errors.New("copy: nil source item")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	size, err := src.Size()
	if err != nil {
		return nil, fmt.Errorf("copy, getting the source size: %w", err)
	}
	r, err := src.Open()
	if err != nil {
		return nil, fmt.Errorf("copy, opening the source: %w", err)
	}
	defer r.Close()

	return c.Put(name, r, size, nil)
}
