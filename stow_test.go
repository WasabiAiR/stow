package stow_test

import (
	"net/url"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

func TestKindByURL(t *testing.T) {
	is := is.New(t)
	u, err := url.Parse("test://container/item")
	is.NoErr(err)
	kind, err := stow.KindByURL(u)
	is.NoErr(err)
	is.Equal(kind, testKind)
}

func TestKinds(t *testing.T) {
	is := is.New(t)
	stow.Register("example", nil, nil)
	is.Equal(stow.Kinds(), []string{"test", "example"})
}

func TestIsCursorEnd(t *testing.T) {
	is := is.New(t)
	is.True(stow.IsCursorEnd(""))
	is.False(stow.IsCursorEnd("anything"))
}
