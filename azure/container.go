package azure

import (
	"io"
	"time"

	az "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/graymeta/stow"
)

type container struct {
	id         string
	properties az.ContainerProperties
	client     *az.BlobStorageClient
}

var _ stow.Container = (*container)(nil)

func (c *container) ID() string {
	return c.id
}

func (c *container) Name() string {
	return c.id
}

func (c *container) Item(id string) (stow.Item, error) {
	blobProperties, err := c.client.GetBlobProperties(c.id, id)
	if err != nil {
		return nil, err
	}
	item := &item{
		id:         id,
		container:  c,
		client:     c.client,
		properties: *blobProperties,
	}
	return item, nil
}

func (c *container) Items(prefix, cursor string) ([]stow.Item, string, error) {
	listblobs, err := c.client.ListBlobs(c.id, az.ListBlobsParameters{
		Marker:     cursor,
		Prefix:     prefix,
		MaxResults: 10,
	})
	if err != nil {
		return nil, "", err
	}
	items := make([]stow.Item, len(listblobs.Blobs))
	for i, blob := range listblobs.Blobs {
		items[i] = &item{
			id:         blob.Name,
			container:  c,
			client:     c.client,
			properties: blob.Properties,
		}
	}
	return items, listblobs.NextMarker, nil
}

func (c *container) Put(name string, r io.Reader, size int64) (stow.Item, error) {
	err := c.client.CreateBlockBlobFromReader(c.id, name, uint64(size), r, nil)
	if err != nil {
		return nil, err
	}
	item := &item{
		id:        name,
		container: c,
		client:    c.client,
		properties: az.BlobProperties{ // TODO: confirm sensible properties
			LastModified:  time.Now().String(), // TODO(piotrrojek): check time format
			Etag:          "",
			ContentLength: size,
		},
	}
	return item, nil
}
