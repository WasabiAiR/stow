package google

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"

	"github.com/flyteorg/stow"
)

var _ stow.ContextLocation = (*Location)(nil)

// A Location contains a client + the configurations used to create the client.
type Location struct {
	config stow.Config
	client *storage.Client
	ctx    context.Context
}

func (l *Location) Service() *storage.Client {
	return l.client
}

// Close simply satisfies the Location interface. There's nothing that
// needs to be done in order to satisfy the interface.
func (l *Location) Close() error {
	return nil // nothing to close
}

// CreateContainer creates a new container, in this case a bucket.
func (l *Location) CreateContainer(containerName string) (stow.Container, error) {
	return l.CreateContainerContext(l.ctx, containerName)
}

// CreateContainerContext is CreateContainer with a context.
func (l *Location) CreateContainerContext(ctx context.Context, containerName string) (stow.Container, error) {
	projId, _ := l.config.Config(ConfigProjectId)
	bucket := l.client.Bucket(containerName)
	if err := bucket.Create(ctx, projId, nil); err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == 409 {
			return &Container{
				name:   containerName,
				client: l.client,
				ctx:    l.ctx,
			}, nil
		}
		return nil, err
	}

	return &Container{
		name:   containerName,
		client: l.client,
		ctx:    l.ctx,
	}, nil
}

// Containers returns a slice of the Container interface, a cursor, and an error.
func (l *Location) Containers(prefix string, cursor string, count int) ([]stow.Container, string, error) {
	return l.ContainersContext(l.ctx, prefix, cursor, count)
}

// ContainersContext is Containers with a context.
func (l *Location) ContainersContext(ctx context.Context, prefix string, cursor string, count int) ([]stow.Container, string, error) {
	projId, _ := l.config.Config(ConfigProjectId)
	call := l.client.Buckets(ctx, projId)
	if prefix != "" {
		call.Prefix = prefix
	}

	p := iterator.NewPager(call, count, cursor)
	var results []*storage.BucketAttrs
	nextPageToken, err := p.NextPage(&results)
	if err != nil {
		return nil, "", err
	}

	var containers []stow.Container
	for _, container := range results {
		containers = append(containers, &Container{
			name:   container.Name,
			client: l.client,
			ctx:    l.ctx,
		})
	}

	return containers, nextPageToken, nil
}

// Container retrieves a stow.Container based on its name which must be
// exact.
func (l *Location) Container(id string) (stow.Container, error) {
	return l.ContainerContext(l.ctx, id)
}

// ContainerContext is Container with a context.
func (l *Location) ContainerContext(ctx context.Context, id string) (stow.Container, error) {
	attrs, err := l.client.Bucket(id).Attrs(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist {
			return nil, stow.ErrNotFound
		}
		return nil, err
	}

	c := &Container{
		name:   attrs.Name,
		client: l.client,
		ctx:    l.ctx,
	}

	return c, nil
}

// RemoveContainer removes a container simply by name.
func (l *Location) RemoveContainer(id string) error {
	return l.RemoveContainerContext(l.ctx, id)
}

// RemoveContainerContext is RemoveContainer with a context.
func (l *Location) RemoveContainerContext(ctx context.Context, id string) error {
	if err := l.client.Bucket(id).Delete(ctx); err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == 404 {
			return stow.ErrNotFound
		}
		return err
	}

	return nil
}

// ItemByURL retrieves a stow.Item by parsing the URL, in this
// case an item is an object.
func (l *Location) ItemByURL(url *url.URL) (stow.Item, error) {
	return l.ItemByURLContext(l.ctx, url)
}

// ItemByURLContext is ItemByURL with a context.
func (l *Location) ItemByURLContext(ctx context.Context, url *url.URL) (stow.Item, error) {
	if url.Scheme == Kind {
		// Url in the form: google://storage.googleapis.com/download/storage/v1/b/some-bucket
		pieces := strings.SplitN(url.Path, "/", 8)

		c, err := l.ContainerContext(ctx, pieces[5])
		if err != nil {
			return nil, notFound(ctx)
		}

		i, err := stow.ItemContext(ctx, c, pieces[7])
		if err != nil {
			return nil, notFound(ctx)
		}
		return i, nil

	} else if url.Scheme == Protocol {
		// Url in the form: gs://some-bucket
		c, err := l.ContainerContext(ctx, url.Host)
		if err != nil {
			return nil, notFound(ctx)
		}

		i, err := stow.ItemContext(ctx, c, url.Path)
		if err != nil {
			return nil, notFound(ctx)
		}
		return i, nil
	} else {
		return nil, errors.New("not valid google storage URL")
	}
}

// notFound returns the error of an item that could not be retrieved: the
// error of the context when that is what stopped the request.
func notFound(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return stow.ErrNotFound
}
