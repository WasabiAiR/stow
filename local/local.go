package stow

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
)

// ConfigKeys are the supported configuration items for
// local storage.
const (
	ConfigKeyPath = "path"
)

// Kind is the kind of Location this package provides.
const Kind = "local"

func init() {
	// register local storage
	stow.Locations[Kind] = func(config stow.Config) stow.Location {
		return &location{
			config: config,
		}
	}
}

type location struct {
	config stow.Config
}

func (l *location) Containers(prefix string) (stow.ContainerList, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, errors.New("missing " + ConfigKeyPath + " configuration")
	}
	files, err := filepath.Glob(filepath.Join(path, prefix+"*"))
	if err != nil {
		return nil, err
	}
	return filesToContainers(path, files...)
}

func (l *location) Container(id string) (stow.Container, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, errors.New("missing " + ConfigKeyPath + " configuration")
	}
	containers, err := filesToContainers(path, id)
	if err != nil {
		return nil, err
	}
	cs := containers.Items()
	if len(cs) == 0 {
		return nil, nil
	}
	return cs[0], nil
}

type containerList struct {
	items []stow.Container
}

func (c *containerList) Items() []stow.Container {
	return c.items
}

// More returns false as the first page of
// containers will always be exhaustive.
func (c *containerList) More() bool {
	return false
}

type container struct {
	name string
	path string
}

func (c *container) ID() string {
	return c.path
}

func (c *container) Name() string {
	return c.name
}

func (c *container) Items() (stow.ItemList, error) {
	files, err := ioutil.ReadDir(c.path)
	if err != nil {
		return nil, err
	}
	var items []stow.Item
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path, err := filepath.Abs(filepath.Join(c.path, f.Name()))
		if err != nil {
			return nil, err
		}
		items = append(items, &item{
			name: f.Name(),
			path: path,
		})
	}
	return &itemlist{items: items}, nil
}

type itemlist struct {
	items []stow.Item
}

func (i *itemlist) Items() []stow.Item {
	return i.items
}

// More returns false as the first page of
// items will always be exhaustive.
func (i *itemlist) More() bool {
	return false
}

type item struct {
	name string
	path string
}

func (i *item) ID() string {
	return i.path
}

func (i *item) Name() string {
	return i.name
}

// filesToContainers takes a list of files and turns it into a
// stow.ContainerList.
func filesToContainers(root string, files ...string) (stow.ContainerList, error) {
	cs := &containerList{}
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			continue
		}
		absroot, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		path, err := filepath.Abs(f)
		if err != nil {
			return nil, err
		}
		name, err := filepath.Rel(absroot, path)
		if err != nil {
			return nil, err
		}
		cs.items = append(cs.items, &container{name: name, path: path})
	}
	return cs, nil
}
