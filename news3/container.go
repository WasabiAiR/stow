package s3

import (
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

type container struct {
	name   string
	region string
	client *s3.S3
	cfg    cfg
}

func (c *container) ID() string {
	return c.name
}

func (c *container) Name() string {
	return c.name
}

func (c *container) Item(id string) (stow.Item, error) {

	params := &s3.GetObjectInput{
		Bucket: aws.String(c.name),
		Key:    aws.String(id),
	}

	res, err := c.client.GetObject(params)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, stow.ErrNotFound
		}
		return nil, errors.Wrap(err, "getting the item")
	}
	defer res.Body.Close()

	i := &item{
		id:           id,
		container:    c,
		client:       c.client,
		eTag:         cleanETag(*res.ETag),
		lastModified: res.LastModified,
		size:         *res.ContentLength,
		metadata:     parseMd(res.Metadata),
	}

	return i, nil
}

func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {

	params := &s3.ListObjectsInput{
		Bucket:  aws.String(c.name),
		Marker:  &cursor,
		MaxKeys: &int64(count),
		Prefix:  &prefix,
	}

	res, err := c.client.ListObjects(params)
	if err != nil {
		return nil, "", errors.Wrap(err, "listing items")
	}

	items := make([]stow.Item, len(res.Contents))

	for i, v := range res.Contents {
		items[i] = &item{
			id:           &v.Key,
			container:    c,
			client:       c.client,
			eTag:         cleanETag(*v.ETag),
			lastModified: v.LastModified,
			size:         *v.Size,
		}
	}

	cursor = ""
	if *res.IsTruncated != nil {
		if *res.IsTruncated {
			cursor = items[len(items)-1].Name()
		}
	}

	return items, cursor, nil
}

func (c *container) RemoveItem(id string) error {

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(c.name),
		Key:    aws.String(id),
	}

	_, err := c.client.DeleteObject(params)
	if err != nil {
		return errors.Wrap(err, "deleting an item")
	}

	return nil
}

func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {

	uploader := s3manager.NewUploaderWithClient(c.client)

	params := &s3manager.UploadInput{
		Bucket: aws.String(c.name),
		Key:    aws.String(name),
		Body:   r,
	}

	_, err := uploader.Upload(params)
	if err != nil {
		return nil, errors.Wrap(err, "uploading an item")
	}

	i := &item{
		id:        name,
		container: c,
		client:    c.client,
		size:      size,
		metadata:  metadata,
	}

	return i, nil
}

// todo(piotr): trim it with strings.TrimFunc?
func cleanETag(etag string) string {
	etag = strings.TrimLeft(etag, `W/`)
	etag = strings.Trim(etag, `"`)
	etag = strings.Trim(etag, `\"`)
}

func parseMd(md map[string]*string) map[string]interface{} {
	m := make(map[string]interface{}, len(md))

	for k, v := range md {
		k = strings.ToLower(k)
		m[k] = *v
	}

	return m
}
