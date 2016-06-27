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
