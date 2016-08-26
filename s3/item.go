package s3

import (
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
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
	client *s3.S3

	// Properties represent the characteristics of the file. Name, Etag, etc.
	properties *s3.Object
}

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
	genericURL := []string{"https://s3-", i.container.Region(), ".amazonaws.com/",
		i.container.Name(), "/", i.Name()}

	return &url.URL{
		Scheme: "s3",
		Path:   strings.Join(genericURL, ""),
	}
}

// Open retrieves specic information about an item based on the container name
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

// LastMod returns the last modified date of the item. The response of an item that is PUT
// does not contain this field. Solution? Detect when the LastModified field (a *time.Time)
// is nil, then do a manual request for it via the Item() method of the container which
// does return the specified field. This more detailed information is kept so that we
// won't have to do it again.
func (i *item) LastMod() (time.Time, error) {
	if i.properties.LastModified == nil {
		it, err := i.container.getItem(i.ID())
		if err != nil {
			return time.Time{}, err
		}

		// Went through the work of sending a request to get this information.
		// Let's keep it since the response contains more specific information
		// about itself.
		i.properties = it.properties
	}

	return *i.properties.LastModified, nil
}

// Metadata returns a nil map and no error.
// TODO: Implement this.
func (i *item) Metadata() (map[string]interface{}, error) {
	return nil, nil
}

// ETag returns the ETag value from the properies field of an item.
func (i *item) ETag() (string, error) {
	return *(i.properties.ETag), nil
}
