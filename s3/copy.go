package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/flyteorg/stow"
)

var (
	// copyObjectMaxSize is the largest object S3 copies with one CopyObject
	// request. Bigger objects are copied in parts.
	copyObjectMaxSize int64 = 5 << 30
	// copyPartSize is the size of the parts of a multipart copy.
	copyPartSize int64 = 1 << 30
)

var _ stow.Copier = (*container)(nil)

// Copy copies src into the container on the S3 side. An item from another
// kind of location, or from another S3 service, is streamed instead.
func (c *container) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	srcItem, ok := src.(*item)
	if !ok || !sameService(srcItem.container, c) {
		return stow.StreamCopy(ctx, c, src, name)
	}
	size, err := srcItem.Size()
	if err != nil {
		return nil, fmt.Errorf("copy, getting the source size: %w", err)
	}
	source := copySource(srcItem.container.name, srcItem.ID())

	var etag string
	if size <= copyObjectMaxSize {
		res, err := c.client.CopyObject(ctx, &s3.CopyObjectInput{
			Bucket:     new(c.name),
			Key:        new(name),
			CopySource: new(source),
		})
		if err != nil {
			return nil, fmt.Errorf("copy, copying the object: %w", err)
		}
		if res.CopyObjectResult != nil && res.CopyObjectResult.ETag != nil {
			etag = cleanEtag(*res.CopyObjectResult.ETag)
		}
	} else {
		etag, err = c.multipartCopy(ctx, srcItem, source, name, size)
		if err != nil {
			return nil, err
		}
	}

	return &item{
		container: c,
		client:    c.client,
		properties: properties{
			ETag: &etag,
			Key:  &name,
			Size: &size,
		},
	}, nil
}

// multipartCopy copies an object that is too big for CopyObject. A multipart
// upload does not inherit anything from the source, so the metadata and the
// content headers are carried over explicitly.
func (c *container) multipartCopy(ctx context.Context, srcItem *item, source, name string, size int64) (string, error) {
	head, err := srcItem.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: new(srcItem.container.name),
		Key:    new(srcItem.ID()),
	})
	if err != nil {
		return "", fmt.Errorf("copy, getting the source object: %w", err)
	}

	upload, err := c.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:             new(c.name),
		Key:                new(name),
		Metadata:           head.Metadata,
		ContentType:        head.ContentType,
		ContentEncoding:    head.ContentEncoding,
		ContentDisposition: head.ContentDisposition,
		ContentLanguage:    head.ContentLanguage,
		CacheControl:       head.CacheControl,
	})
	if err != nil {
		return "", fmt.Errorf("copy, creating the multipart upload: %w", err)
	}

	abort := func() {
		// The caller's context may be the reason for the failure, and the
		// parts are billed until the upload is aborted.
		_, _ = c.client.AbortMultipartUpload(context.Background(), &s3.AbortMultipartUploadInput{
			Bucket:   new(c.name),
			Key:      new(name),
			UploadId: upload.UploadId,
		})
	}

	var parts []types.CompletedPart
	for start, number := int64(0), int32(1); start < size; start, number = start+copyPartSize, number+1 {
		end := start + copyPartSize - 1
		if end >= size {
			end = size - 1
		}
		res, err := c.client.UploadPartCopy(ctx, &s3.UploadPartCopyInput{
			Bucket:          new(c.name),
			Key:             new(name),
			UploadId:        upload.UploadId,
			PartNumber:      new(number),
			CopySource:      new(source),
			CopySourceRange: new(fmt.Sprintf("bytes=%d-%d", start, end)),
		})
		if err != nil {
			abort()
			return "", fmt.Errorf("copy, copying part %d: %w", number, err)
		}
		if res.CopyPartResult == nil {
			abort()
			return "", fmt.Errorf("copy, copying part %d: empty result", number)
		}
		parts = append(parts, types.CompletedPart{
			ETag:       res.CopyPartResult.ETag,
			PartNumber: new(number),
		})
	}

	res, err := c.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          new(c.name),
		Key:             new(name),
		UploadId:        upload.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: parts},
	})
	if err != nil {
		abort()
		return "", fmt.Errorf("copy, completing the multipart upload: %w", err)
	}
	if res.ETag == nil {
		return "", nil
	}
	return cleanEtag(*res.ETag), nil
}

// sameService reports whether both containers live in the same S3 service, so
// a bucket and a key name the same object for both of them. Buckets of two
// S3-compatible endpoints, or of two AWS partitions, do not share a namespace.
func sameService(a, b *container) bool {
	return a.customEndpoint == b.customEndpoint && partition(a) == partition(b)
}

func partition(c *container) string {
	if c.client == nil {
		return ""
	}
	region := c.client.Options().Region
	for _, p := range partitions {
		if strings.HasPrefix(region, p.regionPrefix) {
			return p.id
		}
	}
	return "aws"
}

// partitions lists the AWS partitions other than "aws" with the prefix their
// regions start with.
var partitions = []struct{ regionPrefix, id string }{
	{"cn-", "aws-cn"},
	{"us-gov-", "aws-us-gov"},
	{"us-iso-", "aws-iso"},
	{"us-isob-", "aws-iso-b"},
	{"eu-isoe-", "aws-iso-e"},
	{"us-isof-", "aws-iso-f"},
	{"eusc-", "aws-eusc"},
}

// copySource builds the URL-encoded "bucket/key" S3 expects as a copy source.
// S3 reads a literal "+" in it as a space, so it is escaped as well.
func copySource(bucket, key string) string {
	segments := strings.Split(key, "/")
	for i, segment := range segments {
		segments[i] = strings.ReplaceAll(url.PathEscape(segment), "+", "%2B")
	}
	return url.PathEscape(bucket) + "/" + strings.Join(segments, "/")
}
