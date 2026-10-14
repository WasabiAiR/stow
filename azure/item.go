package azure

import (
	"context"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/flyteorg/stow"
)

type item struct {
	id         string
	container  *container
	client     *blob.Client
	properties *BlobProps
	url        url.URL
	metadata   map[string]any
	infoOnce   sync.Once
	infoErr    error
}

var (
	_ stow.Item              = (*item)(nil)
	_ stow.ContextItem       = (*item)(nil)
	_ stow.ItemRanger        = (*item)(nil)
	_ stow.ContextItemRanger = (*item)(nil)
)

func (i *item) ID() string {
	return i.id
}

func (i *item) Name() string {
	return i.id
}

func (i *item) URL() *url.URL {
	url, _ := url.Parse(i.client.URL())
	url.Scheme = "azure"
	return url
}

func (i *item) Size() (int64, error) {
	return i.properties.ContentLength, nil
}

func (i *item) Open() (io.ReadCloser, error) {
	return i.OpenContext(context.Background())
}

// OpenContext is Open with a context.
func (i *item) OpenContext(ctx context.Context) (io.ReadCloser, error) {
	dlResp, err := i.client.DownloadStream(ctx, nil)
	if err != nil {
		return nil, err
	}
	return dlResp.Body, nil
}

func (i *item) ETag() (string, error) {
	return i.ETagContext(context.Background())
}

// ETagContext is ETag with a context.
func (i *item) ETagContext(_ context.Context) (string, error) {
	return cleanEtag(string(i.properties.ETag)), nil
}

func (i *item) LastMod() (time.Time, error) {
	return i.LastModContext(context.Background())
}

// LastModContext is LastMod with a context.
func (i *item) LastModContext(_ context.Context) (time.Time, error) {
	return i.properties.LastModified, nil
}

func (i *item) Metadata() (map[string]any, error) {
	return i.MetadataContext(context.Background())
}

// MetadataContext is Metadata with a context.
func (i *item) MetadataContext(_ context.Context) (map[string]any, error) {
	return i.metadata, nil
}

// OpenRange opens the item for reading starting at byte start and ending
// at byte end.
func (i *item) OpenRange(start, end uint64) (io.ReadCloser, error) {
	return i.OpenRangeContext(context.Background(), start, end)
}

// OpenRangeContext is OpenRange with a context.
func (i *item) OpenRangeContext(ctx context.Context, start, end uint64) (io.ReadCloser, error) {
	resp, err := i.client.DownloadStream(ctx, &blob.DownloadStreamOptions{
		Range: blob.HTTPRange{
			Offset: int64(start),
			Count:  int64(end) - int64(start) + 1,
		},
	})

	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
