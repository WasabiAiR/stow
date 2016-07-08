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

func (c *container) Put(name string, r io.Reader, size int64) (stow.Item, error) {
	path := filepath.Join(c.path, name)
	item := &item{
		name: name,
		path: path,
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, errors.New("bad size")
	}
	return item, nil
}

func (c *container) Items(cursor string) ([]stow.Item, string, error) {
	files, err := ioutil.ReadDir(c.path)
	if err != nil {
		return nil, "", err
	}
	var items []stow.Item
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path, err := filepath.Abs(filepath.Join(c.path, f.Name()))
		if err != nil {
			return nil, "", err
		}
		items = append(items, &item{
			name: f.Name(),
			path: path,
		})
	}
	return items, "", nil
}

func (c *container) Item(id string) (stow.Item, error) {
	path := id
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, stow.ErrNotFound
	}
	if info.IsDir() {
		return nil, errors.New("unexpected directory")
	}
	item := &item{
		name: filepath.Base(path),
		path: path,
	}
	return item, nil
}
