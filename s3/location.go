package s3

import (
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/graymeta/stow"
)

// A location contains a client + the configurations used to create the client.
type location struct {
	config stow.Config
	client *s3.S3
}

// CreateContainer creates a new container, in this case an S3 bucket.
// The bare minimum needed is a container name, but there are many other
// options that can be provided.
func (l *location) CreateContainer(containerName string) (stow.Container, error) {
	createBucketParams := &s3.CreateBucketInput{
		Bucket: aws.String(containerName), // required
	}

	_, err := l.client.CreateBucket(createBucketParams)
	if err != nil {
		return nil, err
	}

	newContainer := &container{
		name:   containerName,
		client: l.client,
	}

	return newContainer, nil
}

// Containers returns a slice of the Container interface, a cursor, and an error.
// This doesn't seem to exist yet in the API without doing a ton of manual work.
// Get the list of buckets, query every single one to retrieve region info, and finally
// return the list of containers that have a matching region against the client. It's not
// possible to manipulate a container within a region that doesn't match the clients'.
// This is because AWS user credentials can be tied to regions. One solution would be
// to start a new client for every single container where the region matches, this would
// also check the credentials on every new instance... Tabled for later.
func (l *location) Containers(prefix string, cursor string) ([]stow.Container, string, error) {
	var params *s3.ListBucketsInput

	var containers []stow.Container

	// Response returns exported Owner(*s3.Owner) and Bucket(*s3.[]Bucket)
	response, err := l.client.ListBuckets(params)
	if err != nil {
		return nil, "", err
	}

	// Iterate through the slice of pointers to buckets
	for _, bucket := range response.Buckets {
		// Retrieve region information.
		bliParams := &s3.GetBucketLocationInput{
			Bucket: aws.String(*bucket.Name),
		}

		bliResponse, err := l.client.GetBucketLocation(bliParams)
		if err != nil {
			return nil, "", err
		}

		clientRegion, _ := l.config.Config("region")

		if bliResponse.String() == clientRegion {
			newContainer := &container{
				name:   *(bucket.Name),
				client: l.client,
			}

			containers = append(containers, newContainer)
		}
	}

	return containers, "", nil
}

// Close simply satisfies the Location interface. There's nothing that
// needs to be done in order to satisfy the interface.
func (l *location) Close() error {
	return nil // nothing to close
}

// Container retrieves a stow.Container based on its name which must be
// exact.
func (l *location) Container(id string) (stow.Container, error) {
	params := &s3.GetBucketLocationInput{
		Bucket: aws.String(id), // Required
	}

	_, err := l.client.GetBucketLocation(params)
	if err != nil {
		return nil, stow.ErrNotFound
	}

	c := &container{
		name:   id,
		client: l.client,
	}

	return c, nil
}

// RemoveContainer removes a container simply by name.
func (l *location) RemoveContainer(id string) error {
	params := &s3.DeleteBucketInput{
		Bucket: aws.String(id),
	}

	_, err := l.client.DeleteBucket(params)
	if err != nil {
		return err
	}

	return nil
}

// ItemByURL retrieves a stow.Item by parsing the URL, in this
// case an item is an object.
func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	genericString := "https://s3.amazonaws.com/"

	// Cut generic string.
	firstCut := strings.Replace(url.Path, genericString, "", 1)

	// firstCut is in the format <container name>/<item path>. Grab container name.
	firstSlash := strings.Index(firstCut, "/")
	containerName := firstCut[0:firstSlash]

	// item path is everything after the first slash.
	itemPath := firstCut[firstSlash+1:]

	// Get the container by name.
	cont, err := l.Container(containerName)
	if err != nil {
		return nil, stow.ErrNotFound
	}

	// Get the item by its path.
	it, err := cont.Item(itemPath)
	if err != nil {
		return nil, stow.ErrNotFound
	}

	return it, err
}
