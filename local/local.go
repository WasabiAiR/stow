package local

import (
	"errors"
	"io"
	"io/ioutil"
	"net/url"
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

const (
	paramTypeValue = "item"
)

func init() {
	makefn := func(config stow.Config) stow.Location {
		return &location{
			config: config,
		}
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == "file"
	}
	stow.Register(Kind, makefn, kindfn)
}

type location struct {
	config stow.Config
}

func (l *location) CreateContainer(name string) (stow.Container, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, errors.New("missing " + ConfigKeyPath + " configuration")
	}
	fullpath := filepath.Join(path, name)
	if err := os.Mkdir(fullpath, 0777); err != nil {
		return nil, err
	}
	return &container{
		name: name,
		path: fullpath,
	}, nil
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

func (l *location) ItemByURL(u *url.URL) (stow.Item, error) {
	i := &item{}
	i.path = u.Path
	i.name = filepath.Base(i.path)
	return i, nil
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

func (c *container) CreateItem(name string) (stow.Item, io.WriteCloser, error) {
	path := filepath.Join(c.path, name)
	item := &item{
		name: name,
		path: path,
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return item, f, nil
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

func (i *item) URL() *url.URL {
	return &url.URL{
		Scheme: "file",
		Path:   filepath.Clean(i.path),
	}
}

// Open opens the file for reading.
func (i *item) Open() (io.ReadCloser, error) {
	return os.Open(i.path)
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
