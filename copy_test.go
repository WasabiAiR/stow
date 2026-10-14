package stow_test

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/stow"
)

// streamItem is an item of no particular kind of location.
type streamItem struct {
	content  string
	metadata map[string]any
	openErr  error
	closed   bool
}

func (i *streamItem) ID() string                        { return "src" }
func (i *streamItem) Name() string                      { return "src" }
func (i *streamItem) URL() *url.URL                     { return nil }
func (i *streamItem) Size() (int64, error)              { return int64(len(i.content)), nil }
func (i *streamItem) ETag() (string, error)             { return "", nil }
func (i *streamItem) LastMod() (time.Time, error)       { return time.Time{}, nil }
func (i *streamItem) Metadata() (map[string]any, error) { return i.metadata, nil }
func (i *streamItem) Open() (io.ReadCloser, error) {
	if i.openErr != nil {
		return nil, i.openErr
	}
	return &closeRecorder{Reader: strings.NewReader(i.content), item: i}, nil
}

type closeRecorder struct {
	io.Reader
	item *streamItem
}

func (c *closeRecorder) Close() error {
	c.item.closed = true
	return nil
}

// streamContainer records what is put into it.
type streamContainer struct {
	stow.Container
	name     string
	content  string
	size     int64
	metadata map[string]any
	putErr   error
}

func (c *streamContainer) Put(name string, r io.Reader, size int64, metadata map[string]any) (stow.Item, error) {
	if c.putErr != nil {
		return nil, c.putErr
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	c.name, c.content, c.size, c.metadata = name, string(content), size, metadata
	return &streamItem{content: c.content}, nil
}

func TestStreamCopy(t *testing.T) {
	src := &streamItem{content: "the content", metadata: map[string]any{"k": "v"}}
	dst := &streamContainer{}

	item, err := stow.StreamCopy(t.Context(), dst, src, "dir/dst")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "dir/dst", dst.name)
	assert.Equal(t, "the content", dst.content)
	assert.Equal(t, int64(len("the content")), dst.size)
	assert.Equal(t, map[string]any{"k": "v"}, dst.metadata)
	assert.True(t, src.closed, "source was not closed")
}

func TestStreamCopyErrors(t *testing.T) {
	t.Run("nil source", func(t *testing.T) {
		_, err := stow.StreamCopy(t.Context(), &streamContainer{}, nil, "dst")
		assert.Error(t, err)
	})

	t.Run("canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := stow.StreamCopy(ctx, &streamContainer{}, &streamItem{}, "dst")
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("open fails", func(t *testing.T) {
		openErr := errors.New("open failed")
		_, err := stow.StreamCopy(t.Context(), &streamContainer{}, &streamItem{openErr: openErr}, "dst")
		assert.ErrorIs(t, err, openErr)
	})

	t.Run("put fails", func(t *testing.T) {
		putErr := errors.New("put failed")
		src := &streamItem{content: "x"}
		_, err := stow.StreamCopy(t.Context(), &streamContainer{putErr: putErr}, src, "dst")
		assert.ErrorIs(t, err, putErr)
		assert.True(t, src.closed, "source was not closed")
	})
}
