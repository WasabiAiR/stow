package local

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

type location struct {
	// config is the configuration for this location.
	config stow.Config
}

func (l *location) Close() error {
	return nil // nothing to close
}

func (l *location) ItemByURL(u *url.URL) (stow.Item, error) {
	rootPath, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, errors.New("missing " + ConfigKeyPath + " configuration")
	}

	cleanRootPath := filepath.Clean(rootPath)
	rootPathLen := len(cleanRootPath)
	if len(u.Path) < rootPathLen {
		return nil, errors.New("Url is too short")
	}

	rootIndex := strings.Index(u.Path, cleanRootPath)
	path := u.Path[rootIndex + rootPathLen + 1:]

	urlParts := strings.Split(path, "/")
	if len(urlParts) < 2 {
		return nil, errors.New("parsing ItemByURL URL")
	}
	containerName := urlParts[0]
	itemName := strings.Join(urlParts[1:], "/")

	c, err := l.Container(containerName)
	if err != nil {
		return nil, errors.Wrapf(err, "ItemByURL, getting container by the name %s", containerName)
	}

	i, err := c.Item(itemName)
	if err != nil {
		return nil, errors.Wrapf(err, "ItemByURL, getting item by object name %s", itemName)
	}
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

func (l *location) Containers(prefix string, cursor string, count int) ([]stow.Container, string, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, "", errors.New("missing " + ConfigKeyPath + " configuration")
	}
	files, err := filepath.Glob(filepath.Join(path, prefix+"*"))
	if err != nil {
		return nil, "", err
	}

	var cs []stow.Container

	if prefix == stow.NoPrefix && cursor == stow.CursorStart {
		allContainer := container{
			name: "All",
			path: path,
		}

		cs = append(cs, &allContainer)
	}

	cc, err := l.filesToContainers(path, files...)
	if err != nil {
		return nil, "", err
	}

	cs = append(cs, cc...)

	if cursor != stow.CursorStart {
		// seek to the cursor
		ok := false
		for i, c := range cs {
			if c.ID() == cursor {
				ok = true
				cs = cs[i:]
				break
			}
		}
		if !ok {
			return nil, "", stow.ErrBadCursor
		}
	}
	if len(cs) > count {
		cursor = cs[count].ID()
		cs = cs[:count] // limit cs to count
	} else if len(cs) <= count {
		cursor = ""
	}

	return cs, cursor, err
}

func (l *location) Container(id string) (stow.Container, error) {
	path, ok := l.config.Config(ConfigKeyPath)
	if !ok {
		return nil, errors.New("missing " + ConfigKeyPath + " configuration")
	}
	containers, err := l.filesToContainers(path, filepath.Join(path, id))
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

// filesToContainers takes a list of files and turns it into a
// stow.ContainerList.
func (l *location) filesToContainers(root string, files ...string) ([]stow.Container, error) {
	cs := make([]stow.Container, 0, len(files))
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
		cs = append(cs, &container{
			name: name,
			path: path,
		})
	}
	return cs, nil
}
