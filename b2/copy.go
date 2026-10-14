package b2

import (
	"context"
	"fmt"

	"github.com/flyteorg/stow"
	"gopkg.in/kothar/go-backblaze.v0"
)

// copyFileMaxSize is the largest file B2 copies with one b2_copy_file request.
// Bigger files are streamed.
const copyFileMaxSize int64 = 5 << 30

var _ stow.Copier = (*container)(nil)

// Copy copies src into the container on the B2 side, keeping the metadata of
// the source file. An item from another kind of location, or a file too big
// for a single copy request, is streamed instead.
func (c *container) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	srcItem, ok := src.(*item)
	if !ok || srcItem.size > copyFileMaxSize {
		return stow.StreamCopy(ctx, c, src, name)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := srcItem.bucket.CopyFile(srcItem.id, name, c.bucket.ID, backblaze.FileMetaDirectiveCopy)
	if err != nil {
		return nil, fmt.Errorf("copy, copying the file: %w", err)
	}

	return &item{
		id:     file.ID,
		name:   file.Name,
		size:   file.ContentLength,
		bucket: c.bucket,
	}, nil
}
