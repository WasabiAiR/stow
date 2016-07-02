package local

import (
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
)

// filesToContainers takes a list of files and turns it into a
// stow.ContainerList.
func filesToContainers(root string, files ...string) ([]stow.Container, error) {
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
		cs = append(cs, &container{name: name, path: path})
	}
	return cs, nil
}
