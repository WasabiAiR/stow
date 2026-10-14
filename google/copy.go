package google

import (
	"context"
	"errors"
	"net/http"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"

	"github.com/flyteorg/stow"
)

var _ stow.Copier = (*Container)(nil)

// Copy copies src into the container on the Google Cloud Storage side. The
// metadata of the source object is kept. An item from another kind of location
// is streamed instead.
//
// The copy is one request made with the credentials of this container. An
// item of another location may be readable only with its own credentials, so
// when Google Cloud Storage refuses the copy or does not find the source, the
// item is streamed through its own client instead.
func (c *Container) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	srcItem, ok := src.(*Item)
	if !ok {
		return stow.StreamCopy(ctx, c, src, name)
	}

	source := c.client.Bucket(srcItem.container.name).Object(srcItem.name)
	attrs, err := c.Bucket().Object(name).CopierFrom(source).Run(ctx)
	if err != nil {
		if srcItem.client != c.client && sourceUnreadable(err) {
			return stow.StreamCopy(ctx, c, src, name)
		}
		return nil, err
	}

	return c.convertToStowItem(attrs)
}

// sourceUnreadable reports whether a copy failed because the credentials of
// the destination cannot read the source object.
func sourceUnreadable(err error) bool {
	if errors.Is(err, storage.ErrObjectNotExist) || errors.Is(err, storage.ErrBucketNotExist) {
		return true
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == http.StatusForbidden || apiErr.Code == http.StatusNotFound
	}
	return false
}
