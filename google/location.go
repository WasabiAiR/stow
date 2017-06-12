package google

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/graymeta/stow"
	"google.golang.org/api/iterator"
)

// A Location contains a client + the configurations used to create the client.
type Location struct {
	config stow.Config
	client *storage.Client
}

func (l *Location) Client() *storage.Client {
	return l.client
}

// Close simply satisfies the Location interface. There's nothing that
// needs to be done in order to satisfy the interface.
func (l *Location) Close() error {
	return l.Close()
}

// CreateContainer creates a new container, in this case a bucket.
func (l *Location) CreateContainer(containerName string) (stow.Container, error) {

	projId, _ := l.config.Config(ConfigProjectId)
	// Create a bucket.
	bucket := l.client.Bucket(containerName)

	if err := bucket.Create(context.Background(), projId, nil); err != nil {
		return nil, err
	}

	newContainer := &Container{
		name:   containerName,
		client: l.client,
	}

	return newContainer, nil
}

func (l *Location) bucketAttrToContainer(attr *storage.BucketAttrs) *Container {
	return &Container{
		name:   attr.Name,
		client: l.client,
	}
}

// Containers returns a slice of the Container interface, a cursor, and an error.
func (l *Location) Containers(prefix string, cursor string, count int) ([]stow.Container, string, error) {

	projId, _ := l.config.Config(ConfigProjectId)

	// List all objects in a bucket using pagination

	pager := iterator.NewPager(l.client.Buckets(context.Background(), projId), count, cursor)

	var attrs []*storage.BucketAttrs
	nextPageToken, err := pager.NextPage(&attrs)
	if err != nil {
		return nil, "", err
	}
	containerItems := make([]stow.Container, len(attrs))

	for i, o := range attrs {
		containerItems[i] = l.bucketAttrToContainer(o)
	}

	return containerItems, nextPageToken, nil
}

// Container retrieves a stow.Container based on its name which must be
// exact.
func (l *Location) Container(id string) (stow.Container, error) {
	_, err := l.client.Bucket(id).Attrs(context.Background())
	if err != nil {
		return nil, stow.ErrNotFound
	}

	c := &Container{
		name:   id,
		client: l.client,
	}

	return c, nil
}

// RemoveContainer removes a container simply by name.
func (l *Location) RemoveContainer(id string) error {

	return l.client.Bucket(id).Delete(context.Background())
}

// ItemByURL retrieves a stow.Item by parsing the URL, in this
// case an item is an object.
func (l *Location) ItemByURL(url *url.URL) (stow.Item, error) {

	if url.Scheme != Kind {
		return nil, errors.New("not valid google storage URL")
	}

	// /download/storage/v1/b/stowtesttoudhratik/o/a_first%2Fthe%20item
	pieces := strings.SplitN(url.Path, "/", 8)

	c, err := l.Container(pieces[5])
	if err != nil {
		return nil, stow.ErrNotFound
	}

	i, err := c.Item(pieces[7])
	if err != nil {
		return nil, stow.ErrNotFound
	}

	return i, nil
}
