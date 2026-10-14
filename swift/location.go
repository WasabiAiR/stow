package swift

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/flyteorg/stow"
	"github.com/ncw/swift/v2"
)

var _ stow.ContextLocation = (*location)(nil)

type location struct {
	config stow.Config
	client *swift.Connection
}

func (l *location) Close() error {
	return nil // nothing to close
}

func (l *location) CreateContainer(name string) (stow.Container, error) {
	return l.CreateContainerContext(context.Background(), name)
}

// CreateContainerContext is CreateContainer with a context.
func (l *location) CreateContainerContext(ctx context.Context, name string) (stow.Container, error) {
	err := l.client.ContainerCreate(ctx, name, nil)
	if err != nil {
		return nil, err
	}
	container := &container{
		id:     name,
		client: l.client,
	}
	return container, nil
}

func (l *location) Containers(prefix, cursor string, count int) ([]stow.Container, string, error) {
	return l.ContainersContext(context.Background(), prefix, cursor, count)
}

// ContainersContext is Containers with a context.
func (l *location) ContainersContext(ctx context.Context, prefix, cursor string, count int) ([]stow.Container, string, error) {
	params := &swift.ContainersOpts{
		Limit:  count,
		Prefix: prefix,
		Marker: cursor,
	}
	response, err := l.client.Containers(ctx, params)
	if err != nil {
		return nil, "", err
	}
	containers := make([]stow.Container, len(response))
	for i, cont := range response {
		containers[i] = &container{
			id:     cont.Name,
			client: l.client,
			// count: cont.Count,
			// bytes: cont.Bytes,
		}
	}
	marker := ""
	if len(response) == count {
		marker = response[len(response)-1].Name
	}
	return containers, marker, nil
}

func (l *location) Container(id string) (stow.Container, error) {
	return l.ContainerContext(context.Background(), id)
}

// ContainerContext is Container with a context.
func (l *location) ContainerContext(ctx context.Context, id string) (stow.Container, error) {
	_, _, err := l.client.Container(ctx, id)
	// TODO: grab info + headers
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, stow.ErrNotFound
	}

	c := &container{
		id:     id,
		client: l.client,
	}

	return c, nil
}

func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	return l.ItemByURLContext(context.Background(), url)
}

// ItemByURLContext is ItemByURL with a context.
func (l *location) ItemByURLContext(ctx context.Context, url *url.URL) (stow.Item, error) {

	if url.Scheme != Kind {
		return nil, errors.New("not valid swift URL")
	}

	path := strings.TrimLeft(url.Path, "/")
	pieces := strings.SplitN(path, "/", 4)

	// swift://lax-proxy-03.storagesvc.sohonet.com/v1/AUTH_b04239c7467548678b4822e9dad96030/<container_name>/<path_to_object>

	c, err := l.ContainerContext(ctx, pieces[2])
	if err != nil {
		return nil, err
	}

	return stow.ItemContext(ctx, c, pieces[3])
}

func (l *location) RemoveContainer(id string) error {
	return l.RemoveContainerContext(context.Background(), id)
}

// RemoveContainerContext is RemoveContainer with a context.
func (l *location) RemoveContainerContext(ctx context.Context, id string) error {
	return l.client.ContainerDelete(ctx, id)
}
