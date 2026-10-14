package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/flyteorg/stow"
)

// The item struct contains an id (also the name of the file/S3 Object/Item),
// a container which it belongs to (s3 Bucket), a client, and a URL. The last
// field, properties, contains information about the item, including the ETag,
// file name/id, size, owner, last modified date, and storage class.
// see Object type at http://docs.aws.amazon.com/sdk-for-go/api/service/s3/
// for more info.
// All fields are unexported because methods exist to facilitate retrieval.
type item struct {
	// Container information is required by a few methods.
	container *container
	// A client is needed to make requests.
	client *s3.Client
	// properties represent the characteristics of the file. Name, Etag, etc.
	properties properties
	infoMu     sync.Mutex
	tags       map[string]any
	tagsMu     sync.Mutex
}

type properties struct {
	ETag         *string      `type:"string"`
	Key          *string      `min:"1" type:"string"`
	LastModified *time.Time   `type:"timestamp" timestampFormat:"iso8601"`
	Owner        *types.Owner `type:"structure"`
	Size         *int64       `type:"integer"`
	StorageClass *string      `type:"string" enum:"ObjectStorageClass"`
	Metadata     map[string]any
}

var (
	_ stow.Item              = (*item)(nil)
	_ stow.ContextItem       = (*item)(nil)
	_ stow.ItemRanger        = (*item)(nil)
	_ stow.ContextItemRanger = (*item)(nil)
	_ stow.Taggable          = (*item)(nil)
	_ stow.ContextTaggable   = (*item)(nil)
)

// ID returns a string value that represents the name of a file.
func (i *item) ID() string {
	return *i.properties.Key
}

// Name returns a string value that represents the name of the file.
func (i *item) Name() string {
	return *i.properties.Key
}

// Size returns the size of an item in bytes.
func (i *item) Size() (int64, error) {
	return *i.properties.Size, nil
}

// URL returns a formatted string which follows the predefined format
// that every S3 asset is given.
func (i *item) URL() *url.URL {
	if i.container.customEndpoint == "" {
		genericURL := fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", i.container.Region(), i.container.Name(), i.Name())

		return &url.URL{
			Scheme: "s3",
			Path:   genericURL,
		}
	}

	genericURL := fmt.Sprintf("%s/%s", i.container.Name(), i.Name())
	return &url.URL{
		Scheme: "s3",
		Path:   genericURL,
	}
}

// Open retrieves specic information about an item based on the container name
// and path of the file within the container. This response includes the body of
// resource which is returned along with an error.
func (i *item) Open() (io.ReadCloser, error) {
	return i.OpenContext(context.Background())
}

// OpenContext is Open with a context.
func (i *item) OpenContext(ctx context.Context) (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: new(i.container.Name()),
		Key:    new(i.ID()),
	}

	response, err := i.client.GetObject(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Open, getting the object: %w", err)
	}
	return response.Body, nil
}

// LastMod returns the last modified date of the item. The response of an item that is PUT
// does not contain this field. Solution? Detect when the LastModified field (a *time.Time)
// is nil, then do a manual request for it via the Item() method of the container which
// does return the specified field. This more detailed information is kept so that we
// won't have to do it again.
func (i *item) LastMod() (time.Time, error) {
	return i.LastModContext(context.Background())
}

// LastModContext is LastMod with a context.
func (i *item) LastModContext(ctx context.Context) (time.Time, error) {
	err := i.ensureInfo(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("retrieving Last Modified information of Item: %w", err)
	}
	return *i.properties.LastModified, nil
}

// ETag returns the ETag value from the properies field of an item.
func (i *item) ETag() (string, error) {
	return i.ETagContext(context.Background())
}

// ETagContext is ETag with a context.
func (i *item) ETagContext(_ context.Context) (string, error) {
	return *(i.properties.ETag), nil
}

func (i *item) Metadata() (map[string]any, error) {
	return i.MetadataContext(context.Background())
}

// MetadataContext is Metadata with a context.
func (i *item) MetadataContext(ctx context.Context) (map[string]any, error) {
	err := i.ensureInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving metadata: %w", err)
	}
	return i.properties.Metadata, nil
}

func (i *item) ensureInfo(ctx context.Context) error {
	i.infoMu.Lock()
	defer i.infoMu.Unlock()

	if i.properties.Metadata != nil && i.properties.LastModified != nil {
		return nil
	}

	// Retrieve Item information
	itemInfo, err := i.container.getItem(ctx, i.ID())
	if err != nil {
		return err
	}
	i.properties.Metadata = itemInfo.properties.Metadata
	i.properties.LastModified = itemInfo.properties.LastModified
	return nil
}

// Tags returns a map of tags on an Item
func (i *item) Tags() (map[string]any, error) {
	return i.TagsContext(context.Background())
}

// TagsContext is Tags with a context.
func (i *item) TagsContext(ctx context.Context) (map[string]any, error) {
	i.tagsMu.Lock()
	defer i.tagsMu.Unlock()

	if i.tags != nil {
		return i.tags, nil
	}

	params := &s3.GetObjectTaggingInput{
		Bucket: new(i.container.name),
		Key:    new(i.ID()),
	}

	res, err := i.client.GetObjectTagging(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, stow.ErrNotFound
		}
		return nil, fmt.Errorf("getObjectTagging: %w", err)
	}

	tags := make(map[string]any)
	for _, t := range res.TagSet {
		tags[*t.Key] = *t.Value
	}
	i.tags = tags
	return i.tags, nil
}

// OpenRange opens the item for reading starting at byte start and ending
// at byte end.
func (i *item) OpenRange(start, end uint64) (io.ReadCloser, error) {
	return i.OpenRangeContext(context.Background(), start, end)
}

// OpenRangeContext is OpenRange with a context.
func (i *item) OpenRangeContext(ctx context.Context, start, end uint64) (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: new(i.container.Name()),
		Key:    new(i.ID()),
		Range:  new(fmt.Sprintf("bytes=%d-%d", start, end)),
	}

	response, err := i.client.GetObject(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Open, getting the object: %w", err)
	}
	return response.Body, nil
}
