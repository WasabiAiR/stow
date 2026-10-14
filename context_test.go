package stow_test

import (
	"context"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/stow"
)

type ctxKey struct{}

// plainLocation, plainContainer and plainItem have no methods with a context.
type plainLocation struct{ stow.Location }

func (plainLocation) CreateContainer(string) (stow.Container, error) { return plainContainer{}, nil }
func (plainLocation) Container(string) (stow.Container, error)       { return plainContainer{}, nil }
func (plainLocation) RemoveContainer(string) error                   { return nil }
func (plainLocation) ItemByURL(*url.URL) (stow.Item, error)          { return plainItem{}, nil }
func (plainLocation) Containers(string, string, int) ([]stow.Container, string, error) {
	return []stow.Container{plainContainer{}}, "next", nil
}

type plainContainer struct{ stow.Container }

func (plainContainer) Item(string) (stow.Item, error) { return plainItem{}, nil }
func (plainContainer) RemoveItem(string) error        { return nil }
func (plainContainer) Items(string, string, int) ([]stow.Item, string, error) {
	return []stow.Item{plainItem{}}, "next", nil
}
func (plainContainer) Put(string, io.Reader, int64, map[string]any) (stow.Item, error) {
	return plainItem{}, nil
}

type plainItem struct{ stow.Item }

func (plainItem) Open() (io.ReadCloser, error)      { return io.NopCloser(strings.NewReader("plain")), nil }
func (plainItem) ETag() (string, error)             { return "plain", nil }
func (plainItem) LastMod() (time.Time, error)       { return time.Unix(1, 0), nil }
func (plainItem) Metadata() (map[string]any, error) { return map[string]any{"k": "plain"}, nil }
func (plainItem) Tags() (map[string]any, error)     { return map[string]any{"t": "plain"}, nil }
func (plainItem) OpenRange(uint64, uint64) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("plain")), nil
}

// ctxLocation, ctxContainer and ctxItem record the context they are called
// with. Their methods without a context are those of the plain types.
type ctxLocation struct {
	plainLocation
	got context.Context
}

func (l *ctxLocation) CreateContainerContext(ctx context.Context, _ string) (stow.Container, error) {
	l.got = ctx
	return &ctxContainer{}, nil
}
func (l *ctxLocation) ContainersContext(ctx context.Context, _, _ string, _ int) ([]stow.Container, string, error) {
	l.got = ctx
	return nil, "ctx", nil
}
func (l *ctxLocation) ContainerContext(ctx context.Context, _ string) (stow.Container, error) {
	l.got = ctx
	return &ctxContainer{}, nil
}
func (l *ctxLocation) RemoveContainerContext(ctx context.Context, _ string) error {
	l.got = ctx
	return nil
}
func (l *ctxLocation) ItemByURLContext(ctx context.Context, _ *url.URL) (stow.Item, error) {
	l.got = ctx
	return &ctxItem{}, nil
}

type ctxContainer struct {
	plainContainer
	got context.Context
}

func (c *ctxContainer) ItemContext(ctx context.Context, _ string) (stow.Item, error) {
	c.got = ctx
	return &ctxItem{}, nil
}
func (c *ctxContainer) ItemsContext(ctx context.Context, _, _ string, _ int) ([]stow.Item, string, error) {
	c.got = ctx
	return nil, "ctx", nil
}
func (c *ctxContainer) RemoveItemContext(ctx context.Context, _ string) error {
	c.got = ctx
	return nil
}
func (c *ctxContainer) PutContext(ctx context.Context, _ string, _ io.Reader, _ int64, _ map[string]any) (stow.Item, error) {
	c.got = ctx
	return &ctxItem{}, nil
}

type ctxItem struct {
	plainItem
	got context.Context
}

func (i *ctxItem) OpenContext(ctx context.Context) (io.ReadCloser, error) {
	i.got = ctx
	return io.NopCloser(strings.NewReader("ctx")), nil
}
func (i *ctxItem) ETagContext(ctx context.Context) (string, error) {
	i.got = ctx
	return "ctx", nil
}
func (i *ctxItem) LastModContext(ctx context.Context) (time.Time, error) {
	i.got = ctx
	return time.Unix(2, 0), nil
}
func (i *ctxItem) MetadataContext(ctx context.Context) (map[string]any, error) {
	i.got = ctx
	return map[string]any{"k": "ctx"}, nil
}
func (i *ctxItem) OpenRangeContext(ctx context.Context, _, _ uint64) (io.ReadCloser, error) {
	i.got = ctx
	return io.NopCloser(strings.NewReader("ctx")), nil
}
func (i *ctxItem) TagsContext(ctx context.Context) (map[string]any, error) {
	i.got = ctx
	return map[string]any{"t": "ctx"}, nil
}

// contextCalls calls every function that takes a context and returns the
// error of each one by name.
func contextCalls(ctx context.Context, l stow.Location, c stow.Container, i stow.Item) map[string]error {
	errs := map[string]error{}
	_, errs["CreateContainer"] = stow.CreateContainerContext(ctx, l, "c")
	_, _, errs["Containers"] = stow.ContainersContext(ctx, l, "", stow.CursorStart, 1)
	_, errs["Container"] = stow.ContainerContext(ctx, l, "c")
	errs["RemoveContainer"] = stow.RemoveContainerContext(ctx, l, "c")
	_, errs["ItemByURL"] = stow.ItemByURLContext(ctx, l, &url.URL{})
	_, errs["Item"] = stow.ItemContext(ctx, c, "i")
	_, _, errs["Items"] = stow.ItemsContext(ctx, c, "", stow.CursorStart, 1)
	errs["RemoveItem"] = stow.RemoveItemContext(ctx, c, "i")
	_, errs["Put"] = stow.PutContext(ctx, c, "i", strings.NewReader(""), 0, nil)
	_, errs["Open"] = stow.OpenContext(ctx, i)
	_, errs["ETag"] = stow.ETagContext(ctx, i)
	_, errs["LastMod"] = stow.LastModContext(ctx, i)
	_, errs["Metadata"] = stow.MetadataContext(ctx, i)
	_, errs["OpenRange"] = stow.OpenRangeContext(ctx, i.(stow.ItemRanger), 0, 1)
	_, errs["Tags"] = stow.TagsContext(ctx, i.(stow.Taggable))
	return errs
}

func TestContextFunctionsWithoutContextMethods(t *testing.T) {
	errs := contextCalls(t.Context(), plainLocation{}, plainContainer{}, plainItem{})
	require.Len(t, errs, 15)
	for name, err := range errs {
		assert.NoError(t, err, name)
	}

	// the values come from the methods without a context
	_, cursor, err := stow.ItemsContext(t.Context(), plainContainer{}, "", stow.CursorStart, 1)
	require.NoError(t, err)
	assert.Equal(t, "next", cursor)
	etag, err := stow.ETagContext(t.Context(), plainItem{})
	require.NoError(t, err)
	assert.Equal(t, "plain", etag)

	// a cancelled context stops the call
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for name, err := range contextCalls(ctx, plainLocation{}, plainContainer{}, plainItem{}) {
		assert.ErrorIs(t, err, context.Canceled, name)
	}
}

func TestContextFunctionsWithContextMethods(t *testing.T) {
	ctx := context.WithValue(t.Context(), ctxKey{}, "value")
	l, c, i := &ctxLocation{}, &ctxContainer{}, &ctxItem{}

	errs := contextCalls(ctx, l, c, i)
	require.Len(t, errs, 15)
	for name, err := range errs {
		assert.NoError(t, err, name)
	}
	assert.Equal(t, ctx, l.got)
	assert.Equal(t, ctx, c.got)
	assert.Equal(t, ctx, i.got)

	// the values come from the methods with a context
	_, cursor, err := stow.ItemsContext(ctx, c, "", stow.CursorStart, 1)
	require.NoError(t, err)
	assert.Equal(t, "ctx", cursor)
	etag, err := stow.ETagContext(ctx, i)
	require.NoError(t, err)
	assert.Equal(t, "ctx", etag)

	// the context is left to the implementation
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	for name, err := range contextCalls(cancelled, l, c, i) {
		assert.NoError(t, err, name)
	}
}

func TestStreamCopyPassesContext(t *testing.T) {
	ctx := context.WithValue(t.Context(), ctxKey{}, "value")
	src, dst := &ctxItem{}, &ctxContainer{}

	_, err := stow.StreamCopy(ctx, dst, sizedItem{src}, "dst")
	require.NoError(t, err)
	assert.Equal(t, ctx, src.got)
	assert.Equal(t, ctx, dst.got)
}

// sizedItem adds the size StreamCopy asks for to a ctxItem.
type sizedItem struct{ *ctxItem }

func (sizedItem) Size() (int64, error) { return 3, nil }
