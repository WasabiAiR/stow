package azure

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	az "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	"github.com/graymeta/stow"
)

type location struct {
	config stow.Config
	client *az.ServiceClient
}

func (l *location) Close() error {
	return nil // nothing to close
}

func (l *location) CreateContainer(name string) (stow.Container, error) {
	public := az.PublicAccessTypeContainer
	options := az.ContainerCreateOptions{
		Access: &public,
	}
	response, err := l.client.CreateContainer(context.Background(), name, &options)
	if err != nil {
		if strings.Contains(err.Error(), "ErrorCode=ContainerAlreadyExists") {
			return l.Container(name)
		}
		return nil, err
	}
	containerRef := &container{
		id: name,
		properties: az.ContainerProperties{
			LastModified: response.LastModified,
		},
		client: l.client,
	}
	time.Sleep(time.Second * 3)
	return containerRef, nil
}

func (l *location) Containers(prefix, cursor string, count int) ([]stow.Container, string, error) {
	countParam := int32(count)
	options := &az.ListContainersOptions{
		MaxResults: &countParam,
		Prefix:     &prefix,
	}

	if cursor != stow.CursorStart {
		options.Marker = &cursor
	}
	pager := l.client.ListContainers(options)

	success := pager.NextPage(context.Background())
	if !success {
		return nil, "", errors.New("no data found")
	}
	resp := pager.PageResponse()
	err := pager.Err()
	if err != nil {
		return nil, "", err
	}

	containers := make([]stow.Container, len(resp.ContainerItems))
	for i, azContainer := range resp.ContainerItems {
		c := &container{
			id:         *azContainer.Name,
			properties: *azContainer.Properties,
			client:     l.client,
		}
		containers[i] = c

	}

	return containers, *resp.NextMarker, nil
}

func (l *location) Container(id string) (stow.Container, error) {
	cursor := stow.CursorStart
	for {
		containers, crsr, err := l.Containers(id[:3], cursor, 100)
		if err != nil {
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
	if url.Scheme != "azure" {
		return nil, errors.New("not valid azure URL")
	}
	locationURL := strings.Split(url.Host, ".")[0]
	a, ok := l.config.Config(ConfigAccount)
	if !ok {
		// shouldn't really happen
		return nil, errors.New("missing " + ConfigAccount + " config")
	}
	if a != locationURL {
		return nil, errors.New("wrong azure URL")
	}
	path := strings.TrimLeft(url.Path, "/")
	params := strings.SplitN(path, "/", 2)
	if len(params) != 2 {
		return nil, errors.New("wrong path")
	}
	c, err := l.Container(params[0])
	if err != nil {
		return nil, err
	}
	return c.Item(params[1])
}

func (l *location) RemoveContainer(id string) error {
	_, err := l.client.DeleteContainer(context.Background(), id, nil)
	return err
}
