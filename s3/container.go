package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/flyteorg/stow"
)

// Amazon S3 bucket contains a creation date and a name.
type container struct {
	// name is needed to retrieve items.
	name string
	// client is responsible for performing the requests.
	client *s3.Client
	// region describes the AWS Availability Zone of the S3 Bucket.
	region         string
	customEndpoint string
}

var (
	_ stow.Container        = (*container)(nil)
	_ stow.ContextContainer = (*container)(nil)
)

func (c *container) PreSignRequest(ctx context.Context, clientMethod stow.ClientMethod, id string,
	params stow.PresignRequestParams) (response stow.PresignResponse, err error) {

	presignClient := s3.NewPresignClient(c.client, func(o *s3.PresignOptions) {
		o.Expires = params.ExpiresIn
		if signer, ok := c.client.Options().HTTPSignerV4.(v2Signer); ok {
			o.Presigner = signer
		}
	})

	var req *v4.PresignedHTTPRequest
	var requestHeaders map[string]string
	switch clientMethod {
	case stow.ClientMethodGet:
		req, err = presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: new(c.name),
			Key:    new(id),
		})
	case stow.ClientMethodPut:
		var contentMD5 *string
		if len(params.ContentMD5) > 0 {
			contentMD5 = new(params.ContentMD5)
		}

		metadata := make(map[string]string)
		requestHeaders = map[string]string{"Content-Length": strconv.Itoa(len(params.ContentMD5)), "Content-MD5": params.ContentMD5}
		if params.AddContentMD5Metadata {
			metadata[stow.FlyteContentMD5] = params.ContentMD5
			requestHeaders[fmt.Sprintf("x-amz-meta-%s", stow.FlyteContentMD5)] = params.ContentMD5
		}

		req, err = presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket:     new(c.name),
			Key:        new(id),
			ContentMD5: contentMD5,
			Metadata:   metadata,
		})
	default:
		return stow.PresignResponse{}, fmt.Errorf("unsupported client method [%v]", clientMethod.String())
	}

	if err != nil {
		return stow.PresignResponse{}, err
	}

	return stow.PresignResponse{Url: req.URL, RequiredRequestHeaders: requestHeaders}, nil
}

// ID returns a string value which represents the name of the container.
func (c *container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *container) Name() string {
	return c.name
}

// Item returns a stow.Item instance of a container based on the name of the container and the key representing. The
// retrieved item only contains metadata about the object. This ensures that only the minimum amount of information is
// transferred. Calling item.Open() will actually do a get request and open a stream to read from.
func (c *container) Item(id string) (stow.Item, error) {
	return c.ItemContext(context.Background(), id)
}

// ItemContext is Item with a context.
func (c *container) ItemContext(ctx context.Context, id string) (stow.Item, error) {
	return c.getItem(ctx, id)
}

// Items sends a request to retrieve a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return c.ItemsContext(context.Background(), prefix, cursor, count)
}

// ItemsContext is Items with a context.
func (c *container) ItemsContext(ctx context.Context, prefix, cursor string, count int) ([]stow.Item, string, error) {
	itemLimit := int32(count)

	params := &s3.ListObjectsV2Input{
		Bucket:     new(c.Name()),
		StartAfter: &cursor,
		MaxKeys:    &itemLimit,
		Prefix:     &prefix,
	}

	response, err := c.client.ListObjectsV2(ctx, params)
	if err != nil {
		return nil, "", fmt.Errorf("Items, listing objects: %w", err)
	}

	var containerItems []stow.Item

	for _, object := range response.Contents {
		if object.StorageClass == types.ObjectStorageClassGlacier {
			continue
		}
		etag := cleanEtag(*object.ETag) // Copy etag value and remove the strings.

		newItem := &item{
			container: c,
			client:    c.client,
			properties: properties{
				ETag:         &etag,
				Key:          object.Key,
				LastModified: object.LastModified,
				Owner:        object.Owner,
				Size:         object.Size,
				StorageClass: new(string(object.StorageClass)),
			},
		}
		containerItems = append(containerItems, newItem)
	}

	// Create a marker and determine if the list of items to retrieve is complete.
	// If not, the last file is the input to the value of after which item to start
	startAfter := ""
	if aws.ToBool(response.IsTruncated) {
		startAfter = containerItems[len(containerItems)-1].Name()
	}

	return containerItems, startAfter, nil
}

func (c *container) RemoveItem(id string) error {
	return c.RemoveItemContext(context.Background(), id)
}

// RemoveItemContext is RemoveItem with a context.
func (c *container) RemoveItemContext(ctx context.Context, id string) error {
	params := &s3.DeleteObjectInput{
		Bucket: new(c.Name()),
		Key:    new(id),
	}

	_, err := c.client.DeleteObject(ctx, params)
	if err != nil {
		return fmt.Errorf("RemoveItem, deleting object %+v: %w", params, err)
	}
	return nil
}

// Put sends a request to upload content to the container. The arguments
// received are the name of the item (S3 Object), a reader representing the
// content, and the size of the file. Many more attributes can be given to the
// file, including metadata. Keeping it simple for now.
func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]any) (stow.Item, error) {
	return c.PutContext(context.Background(), name, r, size, metadata)
}

// PutContext is Put with a context.
func (c *container) PutContext(ctx context.Context, name string, r io.Reader, size int64, metadata map[string]any) (stow.Item, error) {
	// Convert map[string]interface{} to map[string]string
	mdPrepped, err := prepMetadata(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to create or update item, preparing metadata: %w", err)
	}

	body, ok := r.(readerAtSeeker)
	var offset, length int64
	if ok {
		offset, length, err = bodyRange(body)
		if err != nil {
			// A pipe, FIFO or stdin is an *os.File too, but it can't seek.
			// Nothing has been read from it yet, so read it to the end
			// instead. The transfer manager would seek it too, so hide that.
			ok = false
			r = struct{ io.Reader }{r}
		}
	}
	if ok {
		err = c.putStreamed(ctx, name, body, offset, length, mdPrepped)
	} else {
		uploader := transfermanager.New(c.client, func(o *transfermanager.Options) {
			o.RequestChecksumCalculation = c.client.Options().RequestChecksumCalculation
		})
		_, err = uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
			Bucket:   new(c.name), // Required
			Key:      new(name),   // Required
			Body:     r,
			Metadata: mdPrepped, // map[string]string
		})
	}

	if err != nil {
		return nil, fmt.Errorf("PutObject, putting object: %w", err)
	}
	i, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Key:    new(name),
		Bucket: new(c.name),
	})

	var etag string
	if err == nil && i.ETag != nil {
		etag = cleanEtag(*i.ETag)
	}

	// Some fields are empty because this information isn't included in the response.
	// May have to involve sending a request if we want more specific information.
	// Keeping it simple for now.
	newItem := &item{
		container: c,
		client:    c.client,
		properties: properties{
			ETag: &etag,
			Key:  &name,
			Size: &size,
			//LastModified *time.Time
			//Owner        *types.Owner
			//StorageClass *string
		},
	}

	return newItem, nil
}

// Region returns a string representing the region/availability zone of the container.
func (c *container) Region() string {
	return c.region
}

// A request to retrieve a single item includes information that is more specific than
// a PUT. Instead of doing a request within the PUT, make this method available so that the
// request can be made by the field retrieval methods when necessary. This is the case for
// fields that are left out, such as the object's last modified date. This also needs to be
// done only once since the requested information is retained.
// May be simpler to just stick it in PUT and and do a request every time, please vouch
// for this if so.
func (c *container) getItem(ctx context.Context, id string) (*item, error) {
	params := &s3.HeadObjectInput{
		Bucket: new(c.name),
		Key:    new(id),
	}

	res, err := c.client.HeadObject(ctx, params)
	if err != nil {
		// stow needs ErrNotFound to pass the test but amazon returns an opaque error
		if aerr, ok := errors.AsType[smithy.APIError](err); ok && aerr.ErrorCode() == "NotFound" {
			return nil, stow.ErrNotFound
		}
		return nil, fmt.Errorf("getItem, getting the object: %w", err)
	}

	etag := cleanEtag(*res.ETag) // etag string value contains quotations. Remove them.
	md, err := parseMetadata(res.Metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Item information, parsing metadata: %w", err)
	}

	i := &item{
		container: c,
		client:    c.client,
		properties: properties{
			ETag:         &etag,
			Key:          &id,
			LastModified: res.LastModified,
			Owner:        nil, // not returned in the response.
			Size:         res.ContentLength,
			StorageClass: new(string(res.StorageClass)),
			Metadata:     md,
		},
	}

	return i, nil
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

// prepMetadata parses a raw map into the native type required by S3 to set metadata (map[string]string).
// TODO: validation for key values. This function also assumes that the value of a key value pair is a string.
func prepMetadata(md map[string]any) (map[string]string, error) {
	m := make(map[string]string, len(md))
	for key, value := range md {
		strValue, valid := value.(string)
		if !valid {
			return nil, fmt.Errorf(`value of key '%s' in metadata must be of type string`, key)
		}
		m[key] = strValue
	}
	return m, nil
}

// The first letter of a dash separated key value is capitalized, so perform a ToLower on it.
// This Key transformation of returning lowercase is consistent with other locations..
func parseMetadata(md map[string]string) (map[string]any, error) {
	m := make(map[string]any, len(md))
	for key, value := range md {
		k := strings.ToLower(key)
		m[k] = value
	}
	return m, nil
}
