package s3

import (
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// The item struct contains an id (also the name of the file/S3 Object/Item),
// a container which it belongs to (s3 Bucket), a client, and a URL. The last
// field, properties, contains information about the item, including the ETag,
// file name/id, size, owner, last modified date, and storage class.
// see Object type at http://docs.aws.amazon.com/sdk-for-go/api/service/s3/
// for more info.
type item struct {
	container *container
	client    *s3.S3

	properties *s3.Object
}

// ID returns a string value that represents the name of a file.
func (i *item) ID() string {
	return *(i.properties.Key)
}

// Name returns a string value that represents the name of the file.
func (i *item) Name() string {
	return *(i.properties.Key)
}

// URL returns a formatted string which follows the predefined format
// that every S3 asset is given.
func (i *item) URL() *url.URL {
	//genericURL := []string{"https://s3-", i.container.Region(), ".amazonaws.com/", i.container.Name(), "/", i.Name()}
	genericURL := []string{"https://s3.amazonaws.com", i.container.Name(), i.Name()}

	return &url.URL{
		Scheme: "s3",
		Path:   strings.Join(genericURL, "/"),
	}
}

// Open retrieves specic information about an item baseed on the container name
// and path of the file within the container. This response includes the body of
// resource which is returned along with an error.
func (i *item) Open() (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(i.container.Name()),
		Key:    aws.String(i.ID()),
	}

	response, err := i.client.GetObject(params)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

// ETag returns the ETag value from the properies field of an item.
func (i *item) ETag() (string, error) {
	return *(i.properties.ETag), nil
}

// MD5 doesn't seem to be implemented in S3. There is no metadata field, and the
// Etag field is not always guaranteed to be an MD5 hash. This seems to be true
// for files uploaded in multiple parts.
func (i *item) MD5() (string, error) {
	return "", nil
}
