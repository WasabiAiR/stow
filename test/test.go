package test

import (
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

// Test runs a generic suite of tests for Stow storage
// implementations.
// Passing the kind name and a configuration is enough,
// because implementations should have registered themselves
// via stow.Register.
// Locations should be empty.
func Test(t *testing.T, kind string, config stow.Config) {
	is := is.New(t)

	location, err := stow.New(kind, config)
	is.NoErr(err)
	is.OK(location)

	// make a container
	container1, err := location.CreateContainer("container1")
	is.NoErr(err)
	is.OK(container1)
	is.OK(container1.ID())
	is.Equal(container1.Name(), "container1")
	is.OK(container1.URL())

	containerURL := container1.URL()

	// get container by URL
	containerB, err := location.ContainerByURL(containerURL)
	is.NoErr(err)
	is.Equal(containerB.ID(), container1.ID())

}
