package test

import (
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

// All runs a generic suite of tests for Stow storage
// implementations.
// Passing the kind name and a configuration is enough,
// because implementations should have registered themselves
// via stow.Register.
// Locations should be empty.
func All(t *testing.T, kind string, config stow.Config) {
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

}
