package azure

import (
	"net/url"

	"github.com/graymeta/stow"

	"errors"

	"strings"

	"time"

	az "github.com/Azure/azure-sdk-for-go/storage"
)

type location struct {
	config     stow.Config
	client     *az.BlobStorageClient
	containers []stow.Container
}

var _ stow.Location = (*location)(nil)

func new(cfg stow.Config) (*az.BlobStorageClient, error) {
	acc, ok := cfg.Config(ConfigAccount)
	if !ok {
		return nil, errors.New("missing account id")
	}
	key, ok := cfg.Config(ConfigKey)
	if !ok {
		return nil, errors.New("missing auth key")
	}
	basicClient, err := az.NewBasicClient(acc, key)
	if err != nil {
		return nil, errors.New("bad credentials")
	}
	client := basicClient.GetBlobService()
	return &client, err
}

func (l *location) CreateContainer(name string) (stow.Container, error) {
	if l.client == nil {
		client, err := new(l.config)
		if err != nil {
			return nil, err
		}
		l.client = client
	}

	err := l.client.CreateContainer(name, az.ContainerAccessTypeBlob)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 3)

	container, err := l.Container(name)
	if err != nil {
		return nil, errors.New("Couldn't get newly created container")
	}
	return container, nil
}

func (l *location) Containers(prefix string, page int) ([]stow.Container, bool, error) {
	// TODO implement paging
	if l.client == nil {
		client, err := new(l.config)
		if err != nil {
			return nil, false, err
		}
		l.client = client
	}
	var containers []stow.Container
	cts, err := l.client.ListContainers(az.ListContainersParameters{
		Prefix: prefix,
	})
	if err != nil {
		return nil, false, err
	}

	if len(cts.Containers) == 0 {
		return nil, false, errors.New("No containers with given prefix")
	}

	for _, c := range cts.Containers {
		var sc stow.Container
		sc = &container{
			id:         c.Name,
			properties: c.Properties,
			client:     l.client,
		}
		containers = append(containers, sc)
	}
	l.containers = containers
	return containers, false, nil
}

func (l *location) Container(id string) (stow.Container, error) {
	_, _, err := l.Containers(id[:3], 0)
	if err != nil {
		return nil, stow.ErrNotFound
	}
	for _, i := range l.containers {
		if i.ID() == id {
			return i, nil
		}
	}
	return nil, stow.ErrNotFound
}

func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	if url.Scheme != "azure" {
		return nil, errors.New("not valid azure URL")
	}
	location := strings.Split(url.Host, ".")[0]
	a, ok := l.config.Config(ConfigAccount)
	if !ok {
		// shouldn't really happen
		panic("wrong configuration")
	}
	if a != location {
		return nil, errors.New("wrong azure URL")
	}
	path := strings.TrimLeft(url.Path, "/")
	params := strings.Split(path, "/")
	if len(params) != 2 {
		return nil, errors.New("wrong path")
	}

	c, err := l.Container(params[0])
	if err != nil {
		return nil, err
	}
	items, _, err := c.Items(0)
	if err != nil {
		return nil, err
	}
	for _, x := range items {
		if x.ID() == params[1] {
			return x, nil
		}
	}

	// not found
	return nil, stow.ErrNotFound
}
