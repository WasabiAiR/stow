package local

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
)

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

func (c *container) URL() *url.URL {
	return &url.URL{
		Scheme: "file",
		Path:   filepath.Clean(c.path),
	}
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

func (c *container) Items(page int) ([]stow.Item, bool, error) {
	files, err := ioutil.ReadDir(c.path)
	if err != nil {
		return nil, false, err
	}
	var items []stow.Item
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path, err := filepath.Abs(filepath.Join(c.path, f.Name()))
		if err != nil {
			return nil, false, err
		}
		items = append(items, &item{
			name: f.Name(),
			path: path,
		})
	}
	return items, false, nil
}
