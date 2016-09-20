package google

import (
	"errors"
	"net/url"
	"strings"

	"github.com/graymeta/stow"
	storage "google.golang.org/api/storage/v1"
)

// A location contains a client + the configurations used to create the client.
type location struct {
	config stow.Config
	client *storage.Service
}

// Close simply satisfies the Location interface. There's nothing that
// needs to be done in order to satisfy the interface.
func (l *location) Close() error {
	return nil // nothing to close
}

// CreateContainer creates a new container, in this case a bucket.
func (l *location) CreateContainer(containerName string) (stow.Container, error) {

	projId, _ := l.config.Config(ConfigProjectId)
	// Create a bucket.
	_, err := l.client.Buckets.Insert(projId, &storage.Bucket{Name: containerName}).Do()
	//res, err := l.client.Buckets.Insert(projId, &storage.Bucket{Name: containerName}).Do()
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
func (l *location) Containers(prefix string, cursor string, count int) ([]stow.Container, string, error) {

	projId, _ := l.config.Config(ConfigProjectId)

	// List all objects in a bucket using pagination
	call := l.client.Buckets.List(projId).MaxResults(int64(count))

	if prefix != "" {
		call.Prefix(prefix)
	}

	if cursor != "" {
		call = call.PageToken(cursor)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	containers := make([]stow.Container, len(res.Items))

	for i, o := range res.Items {
		containers[i] = &container{
			name:   o.Name,
			client: l.client,
		}
	}

	return containers, res.NextPageToken, nil
}

// Container retrieves a stow.Container based on its name which must be
// exact.
func (l *location) Container(id string) (stow.Container, error) {

	_, err := l.client.Buckets.Get(id).Do()
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

	if err := l.client.Buckets.Delete(id).Do(); err != nil {
		return err
	}

	return nil
}

// ItemByURL retrieves a stow.Item by parsing the URL, in this
// case an item is an object.
func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {

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
