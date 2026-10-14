package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/flyteorg/stow"
)

var (
	// copyPollInterval is the first delay between two checks of a pending copy.
	copyPollInterval = 100 * time.Millisecond
	// copyPollMaxInterval caps the delay between two checks of a pending copy.
	copyPollMaxInterval = 5 * time.Second
)

var _ stow.Copier = (*container)(nil)

// Copy copies src into the container with the Copy Blob operation, so the
// content stays on the Azure side. The properties and the metadata of the
// source blob are kept.
//
// The credentials of the request authorize reading a source in the storage
// account of the destination. An item from another storage account, or from
// another kind of location, is streamed instead.
func (c *container) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	srcItem, ok := src.(*item)
	if !ok || accountURL(srcItem.container) != accountURL(c) {
		return stow.StreamCopy(ctx, c, src, name)
	}

	name = strings.Replace(name, " ", "+", -1)
	client := c.client.NewBlobClient(name)
	resp, err := client.StartCopyFromURL(ctx, srcItem.client.URL(), nil)
	if err != nil {
		return nil, fmt.Errorf("start copy: %w", err)
	}

	var (
		status       blob.CopyStatusType
		etag         azcore.ETag
		lastModified time.Time
	)
	if resp.CopyStatus != nil {
		status = *resp.CopyStatus
	}
	if resp.ETag != nil {
		etag = *resp.ETag
	}
	if resp.LastModified != nil {
		lastModified = *resp.LastModified
	}

	// Copy Blob may finish asynchronously.
	interval := copyPollInterval
	for status == blob.CopyStatusTypePending {
		select {
		case <-ctx.Done():
			if resp.CopyID != nil {
				abortCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				_, _ = client.AbortCopyFromURL(abortCtx, *resp.CopyID, nil)
				cancel()
			}
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		if interval *= 2; interval > copyPollMaxInterval {
			interval = copyPollMaxInterval
		}

		props, err := client.GetProperties(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("copy status: %w", err)
		}
		if props.CopyID == nil || resp.CopyID == nil || *props.CopyID != *resp.CopyID {
			return nil, errors.New("copy status: the blob was overwritten by another operation")
		}
		if props.CopyStatus == nil {
			return nil, errors.New("copy status: missing")
		}
		status = *props.CopyStatus
		if status != blob.CopyStatusTypeSuccess {
			if props.CopyStatusDescription != nil && status != blob.CopyStatusTypePending {
				return nil, fmt.Errorf("copy %s: %s", status, *props.CopyStatusDescription)
			}
			continue
		}
		if props.ETag != nil {
			etag = *props.ETag
		}
		if props.LastModified != nil {
			lastModified = *props.LastModified
		}
	}
	if status != blob.CopyStatusTypeSuccess {
		return nil, fmt.Errorf("copy %s", status)
	}

	return &item{
		id:        name,
		container: c,
		client:    client,
		metadata:  srcItem.metadata,
		properties: &BlobProps{
			ETag:          etag,
			LastModified:  lastModified,
			ContentLength: srcItem.properties.ContentLength,
		},
	}, nil
}

// accountURL returns the URL of the storage account of the container.
func accountURL(c *container) string {
	if c == nil || c.client == nil {
		return ""
	}
	return strings.TrimSuffix(c.client.URL(), "/"+c.id)
}
