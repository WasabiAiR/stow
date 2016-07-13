package local

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
)

type location struct {
	// config is the configuration for this location.
	config stow.Config
	// pagesize is the number of items to return
	// per page (for Containers and Items).
	pagesize int
}

func (l *location) Close() error {
	return nil // nothing to close
}

func (l *location) ItemByURL(u *url.URL) (stow.Item, error) {
	i := &item{}
	i.path = u.Path
	i.name = filepath.Base(i.path)
	return i, nil
}

func (l *location) RemoveContainer(id string) error {
	return os.RemoveAll(id)
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

func (l *location) Containers(prefix string, cursor string) ([]stow.Container, string, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, "", errors.New("missing " + ConfigKeyPath + " configuration")
	}
	files, err := filepath.Glob(filepath.Join(path, prefix+"*"))
	if err != nil {
		return nil, "", err
	}
	if cursor != stow.CursorStart {
		// seek to the cursor
		for i, file := range files {
			if file == cursor {
				files = files[i:]
				break
			}
		}
	}
	if len(files) > l.pagesize {
		cursor = files[l.pagesize]
		files = files[:l.pagesize] // limit files to pagesize
	} else if len(files) <= l.pagesize {
		cursor = ""
	}
	cs, err := filesToContainers(path, files...)
	return cs, cursor, err
}

func (l *location) Container(id string) (stow.Container, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, errors.New("missing " + ConfigKeyPath + " configuration")
	}
	containers, err := filesToContainers(path, id)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, stow.ErrNotFound
		}
		return nil, err
	}
	if len(containers) == 0 {
		return nil, stow.ErrNotFound
	}
	return containers[0], nil
}
