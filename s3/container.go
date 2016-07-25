package s3

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/graymeta/stow"
)

// Amazon S3 bucket contains a creationdate and a name.
type container struct {
	// Name is needed to retrieve items.
	name string

	// Client is responsible for performing the requests.
	client *s3.S3
	region string
}

// ID returns a string value which represents the name of the container.
func (c *container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *container) Name() string {
	return c.name
}

// Item returns a stow.Item instance of a container based on the
// name of the container and the key representing
func (c *container) Item(id string) (stow.Item, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(c.Name()),
		Key:    aws.String(id),
	}

	response, err := c.client.GetObject(params)
	if err != nil {
		return nil, stow.ErrNotFound
	}

	i := &item{
		container: c,
		client:    c.client,
		properties: &s3.Object{
			ETag:         response.ETag,
			Key:          &id,
			LastModified: response.LastModified,
			Owner:        nil, // Weird that it's not returned in the response.
			Size:         response.ContentLength,
			StorageClass: response.StorageClass,
		},
	}

	return i, nil
}

// Items sends a request to retrieve a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *container) Items(prefix string, cursor string) ([]stow.Item, string, error) {
	itemLimit := int64(10)

	params := &s3.ListObjectsInput{
		Bucket:  aws.String(c.Name()),
		Marker:  &cursor,
		MaxKeys: &itemLimit,
		Prefix:  &prefix,
	}

	response, err := c.client.ListObjects(params)
	if err != nil {
		return nil, "", err
	}

	// Allocate space for the Item slice.
	containerItems := make([]stow.Item, len(response.Contents))

	for i, object := range response.Contents {
		containerItems[i] = &item{
			container:  c,
			client:     c.client,
			properties: object,
		}
	}

	// If there is a marker that denotes the next Item to be read in the
	// next page, return it.
	marker := ""
	if response.NextMarker != nil {
		marker = *response.NextMarker
	}

	return containerItems, marker, nil
}

func (c *container) RemoveItem(id string) error {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(c.Name()),
		Key:    aws.String(id),
	}

	_, err := c.client.DeleteObject(params)
	if err != nil {
		return err
	}

	return nil
}

// Put sends a request to upload content to the container. The arguments
// received are the name of the item (S3 Object), a reader representing the
// content, and the size of the file. Many more attributes can be given to the
// file, including metadata. Keeping it simple for now.
func (c *container) Put(name string, r io.Reader, size int64) (stow.Item, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil
	}

	params := &s3.PutObjectInput{
		Bucket:        aws.String(c.name), // Required
		Key:           aws.String(name),   // Required
		ContentLength: aws.Int64(size),
		Body:          bytes.NewReader(content),
	}

	// Only Etag returned.
	response, err := c.client.PutObject(params)
	if err != nil {
		return nil, err
	}

	newItem := &item{
		container: c,
		client:    c.client,
		properties: &s3.Object{
			ETag: response.ETag,
			Key:  &name,
		},
	}

	return newItem, nil
}

// Region returns a string representing the region/availability zone
// of the container.
func (c *container) Region() string {
	return c.region
}
