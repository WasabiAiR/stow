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
func Test(t *testing.T, kind string, config stow.Config) {
	is := is.New(t)

	location, err := stow.New(kind, config)
	is.NoErr(err)
	is.OK(location)

}
