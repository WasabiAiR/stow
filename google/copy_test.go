package google

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/flyteorg/stow"
)

// TestCopyEmulator runs against a Google Cloud Storage emulator, e.g.
// STORAGE_EMULATOR_HOST=127.0.0.1:4443 with fake-gcs-server.
func TestCopyEmulator(t *testing.T) {
	if os.Getenv("STORAGE_EMULATOR_HOST") == "" {
		t.Skip("skipping test because missing STORAGE_EMULATOR_HOST")
	}
	ctx := t.Context()
	// The context of a test is cancelled before its cleanups run.
	cleanupCtx := context.WithoutCancel(ctx)

	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	newContainer := func(prefix string) *Container {
		name := fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
		require.NoError(t, client.Bucket(name).Create(ctx, "stow", nil))
		c := &Container{name: name, client: client, ctx: ctx}
		t.Cleanup(func() {
			items, _, _ := c.ItemsContext(cleanupCtx, "", stow.CursorStart, 100)
			for _, i := range items {
				_ = c.RemoveItemContext(cleanupCtx, i.ID())
			}
			_ = client.Bucket(name).Delete(cleanupCtx)
		})
		return c
	}
	srcContainer, dstContainer := newContainer("stow-copy-src"), newContainer("stow-copy-dst")

	content := make([]byte, 9<<20)
	_, err = rand.Read(content)
	require.NoError(t, err)
	metadata := map[string]any{"stow": "copy"}
	_, err = srcContainer.Put("dir/src item.pb", bytes.NewReader(content), int64(len(content)), metadata)
	require.NoError(t, err)

	src, err := srcContainer.Item("dir/src item.pb")
	require.NoError(t, err)

	copied, err := dstContainer.Copy(ctx, src, "dir/dst.pb")
	require.NoError(t, err)
	assert.Equal(t, "dir/dst.pb", copied.ID())
	size, err := copied.Size()
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)

	dst, err := dstContainer.Item("dir/dst.pb")
	require.NoError(t, err)
	r, err := dst.Open()
	require.NoError(t, err)
	defer r.Close()
	got, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(content, got), "copied content differs")
	md, err := dst.Metadata()
	require.NoError(t, err)
	assert.Equal(t, metadata, md)

	t.Run("item of another client", func(t *testing.T) {
		other, err := storage.NewClient(ctx, option.WithoutAuthentication())
		require.NoError(t, err)
		t.Cleanup(func() { _ = other.Close() })
		otherDst := &Container{name: dstContainer.name, client: other, ctx: ctx}

		copied, err := otherDst.Copy(ctx, src, "dir/other-client.pb")
		require.NoError(t, err)
		size, err := copied.Size()
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)), size)
	})

	t.Run("missing source", func(t *testing.T) {
		require.NoError(t, srcContainer.RemoveItem(src.ID()))
		_, err := dstContainer.Copy(ctx, src, "dir/missing.pb")
		assert.Error(t, err)
	})
}

func TestSourceUnreadable(t *testing.T) {
	assert.True(t, sourceUnreadable(&googleapi.Error{Code: http.StatusForbidden}))
	assert.True(t, sourceUnreadable(&googleapi.Error{Code: http.StatusNotFound}))
	assert.True(t, sourceUnreadable(fmt.Errorf("copy: %w", storage.ErrObjectNotExist)))
	assert.False(t, sourceUnreadable(&googleapi.Error{Code: http.StatusInternalServerError}))
	assert.False(t, sourceUnreadable(context.Canceled))
}

func TestCopyNilItem(t *testing.T) {
	_, err := (&Container{}).Copy(t.Context(), nil, "name")
	assert.Error(t, err)
}
