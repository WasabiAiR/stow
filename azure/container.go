package azure

import (
	"context"
	az "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
	"io"
	"strings"
)

// The maximum size of an object that can be Put in a single request
const maxPutSize = 256 * 1024 * 1024

// timeFormat is the time format for azure.
var timeFormat = "Mon, 2 Jan 2006 15:04:05 MST"

type container struct {
	id         string
	properties az.ContainerProperties
	client     *az.ServiceClient
}

var _ stow.Container = (*container)(nil)

func (c *container) ID() string {
	return c.id
}

func (c *container) Name() string {
	return c.id
}

func (c *container) Item(id string) (stow.Item, error) {
	ctx := context.Background()

	id = strings.Replace(id, " ", "+", -1)
	blob, err := getBlobClient(c.client, c.id, id)
	if err != nil {
		return nil, err
	}
	properties, err := blob.GetProperties(ctx, nil)
	if err != nil {
		if strings.Contains(err.Error(), "BlobNotFound") {
			return nil, stow.ErrNotFound
		}
		return nil, err
	}
	item := &item{
		id:         id,
		container:  c,
		client:     c.client,
		properties: newStowProperties(cleanEtag(*properties.ETag), *properties.ContentLength, *properties.LastModified),
	}
	return item, nil
}

func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	countParam := int32(count)
	params := az.ContainerListBlobsFlatOptions{
		Prefix:     &prefix,
		MaxResults: &countParam,
	}

	containerObj, err := c.client.NewContainerClient(c.id)
	if err != nil {
		return nil, "", err
	}

	if cursor != "" {
		params.Marker = &cursor
	}
	pager := containerObj.ListBlobsFlat(&params)
	//Retrieve next page
	success := pager.NextPage(context.Background())
	if !success {
		return nil, "", errors.New("eof, no more data")
	}
	resp := pager.PageResponse()
	items := make([]stow.Item, len(resp.Segment.BlobItems))

	for i, blob := range resp.Segment.BlobItems {

		// Clean Etag just in case.
		cleanTag := cleanEtag(*blob.Properties.Etag)
		blob.Properties.Etag = &cleanTag

		items[i] = &item{
			id:         *blob.Name,
			container:  c,
			client:     c.client,
			properties: newStowProperties(*blob.Properties.Etag, *blob.Properties.ContentLength, *blob.Properties.LastModified),
		}
	}
	return items, *resp.NextMarker, nil
}

func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	name = strings.Replace(name, " ", "+", -1)
	resp, err := c.multipartUpload(name, r, metadata)

	var (
		etag string = ""
	)

	if err != nil {
		return nil, err

	}
	item := &item{
		id:         name,
		container:  c,
		client:     c.client,
		properties: newStowProperties(etag, size, *resp.LastModified),
	}
	return item, err
}

func (c *container) SetItemMetadata(itemName string, md map[string]string) error {
	containerObj, err := c.client.NewContainerClient(c.id)
	if err != nil {
		return err
	}
	blobClient, err := containerObj.NewBlobClient(itemName)
	if err != nil {
		return err
	}
	_, err = blobClient.SetMetadata(context.Background(), md, nil)
	return err
}

func parseMetadata(md map[string]string) (map[string]interface{}, error) {
	rtnMap := make(map[string]interface{}, len(md))
	for key, value := range md {
		rtnMap[key] = value
	}
	return rtnMap, nil
}

func prepMetadata(md map[string]interface{}) (map[string]string, error) {
	rtnMap := make(map[string]string, len(md))
	for key, value := range md {
		str, ok := value.(string)
		if !ok {
			return nil, errors.Errorf(`value of key '%s' in metadata must be of type string`, key)
		}
		rtnMap[key] = str
	}
	return rtnMap, nil
}

func (c *container) RemoveItem(id string) error {
	containerObj, err := c.client.NewContainerClient(c.id)
	if err != nil {
		return err
	}
	blobClient, err := containerObj.NewBlobClient(id)
	if err != nil {
		return err
	}
	_, err = blobClient.Delete(context.Background(), nil)
	return err
}

// Remove quotation marks from beginning and end. This includes quotations that
// are escaped. Also removes leading `W/` from prefix for weak Etags.
//
// Based on the Etag spec, the full etag value (<FULL ETAG VALUE>) can include:
// - W/"<ETAG VALUE>"
// - "<ETAG VALUE>"
// - ""
// Source: https://tools.ietf.org/html/rfc7232#section-2.3
//
// Based on HTTP spec, forward slash is a separator and must be enclosed in
// quotes to be used as a valid value. Hence, the returned value may include:
// - "<FULL ETAG VALUE>"
// - \"<FULL ETAG VALUE>\"
// Source: https://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2
//
// This function contains a loop to check for the presence of the three possible
// filler characters and strips them, resulting in only the Etag value.
func cleanEtag(etag string) string {
	for {
		// Check if the filler characters are present
		if strings.HasPrefix(etag, `\"`) {
			etag = strings.Trim(etag, `\"`)

		} else if strings.HasPrefix(etag, `"`) {
			etag = strings.Trim(etag, `"`)

		} else if strings.HasPrefix(etag, `W/`) {
			etag = strings.Replace(etag, `W/`, "", 1)

		} else {

			break
		}
	}

	return etag
}
