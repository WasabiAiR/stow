package stow_test

import (
	"net/url"

	"github.com/graymeta/stow"
)

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {
		return &testLocation{
			config: config,
		}, nil
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == testKind
	}
	stow.Register(testKind, makefn, kindfn)
}

const testKind = "test"

type testLocation struct {
	config stow.Config
}

func (l *testLocation) CreateContainer(name string) (stow.Container, error) {
	return nil, nil
}

func (l *testLocation) Container(id string) (stow.Container, error) {
	return nil, nil
}
func (l *testLocation) Containers(prefix string, page int) ([]stow.Container, bool, error) {
	return nil, false, nil
}

func (l *testLocation) ItemByURL(u *url.URL) (stow.Item, error) {
	return nil, nil
}

func (l *testLocation) ContainerByURL(u *url.URL) (stow.Container, error) {
	return nil, nil
}
