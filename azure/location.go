package azure

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	azcontainer "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/flyteorg/stow"
)

var _ stow.ContextLocation = (*location)(nil)

type location struct {
	accountName       string
	uploadConcurrency int
	client            *azblob.Client
	preSigner         RequestPreSigner
}

func (l *location) Close() error {
	return nil // nothing to close
}

var publicAccessTypeContainer = azcontainer.PublicAccessTypeContainer

// CreateContainer follows the contract from stow.Location, with one notable opinion.
// Attempts to create an already-existing container will not produce an error.
func (l *location) CreateContainer(name string) (stow.Container, error) {
	return l.CreateContainerContext(context.Background(), name)
}

// CreateContainerContext is CreateContainer with a context.
func (l *location) CreateContainerContext(ctx context.Context, name string) (stow.Container, error) {
	resp, err := l.client.CreateContainer(
		ctx,
		name,
		&azblob.CreateContainerOptions{Access: &publicAccessTypeContainer})

	if err != nil {
		var tErr *azcore.ResponseError
		ok := errors.As(err, &tErr)
		// Note: StatusConflict (409) is used for both "already exists"
		// and "deleting" failures.
		if ok &&
			tErr.StatusCode == http.StatusConflict &&
			tErr.ErrorCode == "ContainerAlreadyExists" {
			return l.ContainerContext(ctx, name)
		}
		return nil, err
	}

	container := &container{
		id: name,
		properties: &BlobProps{
			ETag:         *resp.ETag,
			LastModified: *resp.LastModified,
		},
		client:            l.client.ServiceClient().NewContainerClient(name),
		preSigner:         l.preSigner,
		uploadConcurrency: l.uploadConcurrency,
	}
	// TK: What is this here for? Presumably to wait for the container to
	// really be available. If that's the case, a validation mechanism is
	// a much better path if you want this to always work.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Second * 3):
	}
	return container, nil
}

func (l *location) Containers(prefix, cursor string, count int) ([]stow.Container, string, error) {
	return l.ContainersContext(context.Background(), prefix, cursor, count)
}

// ContainersContext is Containers with a context.
func (l *location) ContainersContext(ctx context.Context, prefix, cursor string, count int) ([]stow.Container, string, error) {
	params := azblob.ListContainersOptions{
		MaxResults: new(int32(count)),
		Prefix:     &prefix,
	}
	if cursor != stow.CursorStart {
		params.Marker = &cursor
	}

	pager := l.client.NewListContainersPager(&params)
	resp, err := pager.NextPage(ctx)
	if err != nil {
		return nil, cursor, err
	}

	stowContainers := make([]stow.Container, len(resp.ContainerItems))
	for i, azContainer := range resp.ContainerItems {
		stowContainers[i] = &container{
			id: *azContainer.Name,
			properties: &BlobProps{
				ETag:         *azContainer.Properties.ETag,
				LastModified: *azContainer.Properties.LastModified,
			},
			client:            l.client.ServiceClient().NewContainerClient(*azContainer.Name),
			preSigner:         l.preSigner,
			uploadConcurrency: l.uploadConcurrency,
		}
	}

	return stowContainers, *resp.NextMarker, nil
}

func (l *location) Container(id string) (stow.Container, error) {
	return l.ContainerContext(context.Background(), id)
}

// ContainerContext is Container with a context.
func (l *location) ContainerContext(ctx context.Context, id string) (stow.Container, error) {
	cursor := stow.CursorStart
	for {
		containers, crsr, err := l.ContainersContext(ctx, id[:3], cursor, 100)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			return nil, stow.ErrNotFound
		}
		for _, i := range containers {
			if i.ID() == id {
				return i, nil
			}
		}

		cursor = crsr
		if cursor == "" {
			break
		}
	}

	return nil, stow.ErrNotFound
}

func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	return l.ItemByURLContext(context.Background(), url)
}

// ItemByURLContext is ItemByURL with a context.
func (l *location) ItemByURLContext(ctx context.Context, url *url.URL) (stow.Item, error) {
	if url.Scheme != "azure" {
		return nil, errors.New("not valid azure URL")
	}

	locationAccountPart, _, _ := strings.Cut(url.Host, ".")
	if locationAccountPart != l.accountName {
		return nil, errors.New("wrong azure URL")
	}

	path := strings.TrimLeft(url.Path, "/")
	params := strings.SplitN(path, "/", 2)
	if len(params) != 2 {
		return nil, errors.New("wrong path")
	}
	c, err := l.ContainerContext(ctx, params[0])
	if err != nil {
		return nil, err
	}
	return stow.ItemContext(ctx, c, params[1])
}

func (l *location) RemoveContainer(id string) error {
	return l.RemoveContainerContext(context.Background(), id)
}

// RemoveContainerContext is RemoveContainer with a context.
func (l *location) RemoveContainerContext(ctx context.Context, id string) error {
	_, err := l.client.DeleteContainer(ctx, id, &azblob.DeleteContainerOptions{})
	return err
}
