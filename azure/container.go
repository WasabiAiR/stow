package azure

import (
	"io"

	"github.com/graymeta/stow"

	az "github.com/Azure/azure-sdk-for-go/storage"
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

func (c *container) Items(page int) ([]stow.Item, bool, error) {
	// TODO implement paging
	var (
		previousPage = 0
		pageCount    = 1
		next         = ""
		sis          []stow.Item
	)

	for pp := 0; next != "" || pp == 0; pp++ {
		listblobs, err := c.client.ListBlobs(c.id, az.ListBlobsParameters{
			Marker:     next,
			MaxResults: 100,
		})
		if err != nil {
			return nil, false, err
		}

		if pp != previousPage {
			pageCount++
		}

		next = listblobs.NextMarker
		for _, x := range listblobs.Blobs {

			ii := item{
				id:         x.Name,
				container:  c,
				client:     c.client,
				blob:       &x,
				properties: &x.Properties,
				page:       page,
			}
			sis = append(sis, &ii)
		}
	}

	return sis, false, nil
}

func (c *container) CreateItem(name string) (stow.Item, io.WriteCloser, error) {
	panic("not implemented")
}
