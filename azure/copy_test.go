package azure

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	azcontainer "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The account and the key every Azurite instance accepts.
const (
	azuriteAccount = "devstoreaccount1"
	azuriteKey     = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
)

// TestCopyAzurite runs against the Azurite emulator, e.g.
// AZURITEENDPOINT=http://127.0.0.1:10000
func TestCopyAzurite(t *testing.T) {
	endpoint := os.Getenv("AZURITEENDPOINT")
	if endpoint == "" {
		t.Skip("skipping test because missing AZURITEENDPOINT")
	}
	ctx := t.Context()

	cred, err := azblob.NewSharedKeyCredential(azuriteAccount, azuriteKey)
	require.NoError(t, err)
	newContainer := func(prefix string) *container {
		name := fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
		client, err := azcontainer.NewClientWithSharedKeyCredential(
			fmt.Sprintf("%s/%s/%s", endpoint, azuriteAccount, name), cred, nil)
		require.NoError(t, err)
		_, err = client.Create(ctx, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _, _ = client.Delete(context.WithoutCancel(ctx), nil) })
		return &container{id: name, client: client, uploadConcurrency: defaultUploadConcurrency}
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

	t.Run("missing source", func(t *testing.T) {
		require.NoError(t, srcContainer.RemoveItem(src.ID()))
		_, err := dstContainer.Copy(ctx, src, "dir/missing.pb")
		assert.Error(t, err)
	})
}

// copyServer fakes the Copy Blob operation. The copy stays pending for the
// given number of status checks and then ends with finalStatus.
func copyServer(t *testing.T, pendingChecks int32, finalStatus string) (*container, *item, *int32) {
	var checks int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/dst/copied.pb") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("ETag", `"0x1"`)
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		w.Header().Set("x-ms-copy-id", "copy-1")
		switch r.Method {
		case http.MethodPut:
			if r.ContentLength > 0 {
				t.Errorf("copy request carries a %d bytes body", r.ContentLength)
			}
			if source := r.Header.Get("x-ms-copy-source"); !strings.HasSuffix(source, "/src/item.pb") {
				t.Errorf("unexpected copy source %s", source)
			}
			w.Header().Set("x-ms-copy-status", "pending")
			w.WriteHeader(http.StatusAccepted)
		case http.MethodHead:
			status := "pending"
			if atomic.AddInt32(&checks, 1) > pendingChecks {
				status = finalStatus
				w.Header().Set("x-ms-copy-status-description", "the reason")
			}
			w.Header().Set("x-ms-copy-status", status)
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	}))
	t.Cleanup(server.Close)

	oldInterval := copyPollInterval
	copyPollInterval = time.Millisecond
	t.Cleanup(func() { copyPollInterval = oldInterval })

	newContainer := func(name string) *container {
		client, err := azcontainer.NewClientWithNoCredential(server.URL+"/"+name, nil)
		require.NoError(t, err)
		return &container{id: name, client: client}
	}
	src := newContainer("src")
	return newContainer("dst"), &item{
		id:         "item.pb",
		container:  src,
		client:     src.client.NewBlobClient("item.pb"),
		properties: &BlobProps{ContentLength: 42},
	}, &checks
}

func TestCopyWaitsForPendingCopy(t *testing.T) {
	dst, src, checks := copyServer(t, 2, "success")

	copied, err := dst.Copy(t.Context(), src, "copied.pb")
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(checks))
	assert.Equal(t, "copied.pb", copied.ID())
	size, err := copied.Size()
	require.NoError(t, err)
	assert.Equal(t, int64(42), size)
}

func TestCopyFailedCopy(t *testing.T) {
	dst, src, _ := copyServer(t, 1, "failed")

	_, err := dst.Copy(t.Context(), src, "copied.pb")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "the reason")
}

func TestAccountURL(t *testing.T) {
	newContainer := func(account, name string) *container {
		client, err := azcontainer.NewClientWithNoCredential("https://"+account+".blob.core.windows.net/"+name, nil)
		require.NoError(t, err)
		return &container{id: name, client: client}
	}

	assert.Equal(t, "https://one.blob.core.windows.net", accountURL(newContainer("one", "data")))
	assert.Equal(t, accountURL(newContainer("one", "src")), accountURL(newContainer("one", "dst")))
	assert.NotEqual(t, accountURL(newContainer("one", "data")), accountURL(newContainer("two", "data")))
}

func TestCopyNilItem(t *testing.T) {
	_, err := (&container{}).Copy(t.Context(), nil, "name")
	assert.Error(t, err)
}
