package local_test

import (
	"io/ioutil"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
)

func TestItemReader(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}
	l, err := stow.Dial(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)
	containers, cursor, err := l.Containers("t", stow.CursorStart, 10)
	is.NoErr(err)
	is.OK(containers)
	is.Equal(cursor, "")
	three, err := l.Container(containers[0].ID())

	items, cursor, err := three.Items("", stow.CursorStart, 10)
	is.NoErr(err)
	is.Equal(cursor, "")
	item1 := items[0]

	rc, err := item1.Open()
	defer rc.Close()
	is.NoErr(err)
	b, err := ioutil.ReadAll(rc)
	is.NoErr(err)
	is.Equal("3.1", string(b))

}

func TestHardlink(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}
	l, err := stow.Dial(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	containers, cursor, err := l.Containers("z", stow.CursorStart, 10)
	is.NoErr(err)
	is.OK(containers)
	is.Equal(cursor, "")

	links, err := l.Container(containers[0].ID())
	is.NoErr(err)

	items, cursor, err := links.Items("", stow.CursorStart, 10)
	is.NoErr(err)
	is.Equal(cursor, "")

	for _, item := range items {
		if item.Name() == "hardlink" {
			meta, err := item.Metadata()
			is.NoErr(err)
			is.OK(meta)

			is.Equal(meta["is_dir"], false)
			is.True(meta["is_hardlink"])
			is.False(meta["is_symlink"])

			// fmt.Println(meta["link"]) // we have to fix it, because right now it doesn't return the link
		}
	}
}
