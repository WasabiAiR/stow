package sftp

import (
	"context"

	"github.com/flyteorg/stow"
)

var _ stow.Copier = (*container)(nil)

// Copy copies src into the container. SFTP has no server-side copy, so the
// content is streamed.
func (c *container) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	return stow.StreamCopy(ctx, c, src, name)
}
