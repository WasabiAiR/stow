package local

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
)

type location struct {
	config stow.Config
}

func (l *location) ItemByURL(u *url.URL) (stow.Item, error) {
	i := &item{}
	i.path = u.Path
	i.name = filepath.Base(i.path)
	return i, nil
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
	abspath, err := filepath.Abs(fullpath)
	if err != nil {
		return nil, err
	}
	return &container{
		name: name,
		path: abspath,
	}, nil
}

func (l *location) Containers(prefix string, page int) ([]stow.Container, bool, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, false, errors.New("missing " + ConfigKeyPath + " configuration")
	}
	files, err := filepath.Glob(filepath.Join(path, prefix+"*"))
	if err != nil {
		return nil, false, err
	}
	cs, err := filesToContainers(path, files...)
	return cs, false, err
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
	if len(containers) == 0 {
		return nil, nil // TODO: should this be stow.ErrNotFound?
	}
	return containers[0], nil
}
