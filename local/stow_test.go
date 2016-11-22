package local_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	is := is.New(t)

	dir, err := ioutil.TempDir("testdata", "stow")
	is.NoErr(err)
	defer os.RemoveAll(dir)
	cfg := stow.ConfigMap{"path": dir}

	test.All(t, "local", cfg)
}

func TestIssue109(t *testing.T) {
	is := is.New(t)

	dir, err := ioutil.TempDir("testdata", "issue109")
	is.NoErr(err)
	defer os.RemoveAll(dir)

	err = os.MkdirAll(filepath.Join(dir, "example"), 0700)
	is.NoErr(err)

	cfg := stow.ConfigMap{"path": dir}

	l, err := stow.Dial("local", cfg)
	is.NoErr(err)
	is.OK(l)

	cs, _, _ := l.Containers("", stow.CursorStart, 10)
	is.OK(cs)

	var tmpID string

	for _, c := range cs {
		if c.Name() == "example" {
			tmpID = c.ID()
		}
	}

	c, err := l.Container(tmpID)
	is.NoErr(err)
	is.OK(c)

	contents := "This is a new files stored in the cloud"
	r := strings.NewReader(contents)
	size := int64(len(contents))

	i, err := c.Put("index.txt", r, size, nil)
	is.NoErr(err)
	is.OK(i)

	result, err := c.Item(i.ID())
	is.NoErr(err)

	rc, err := result.Open()
	is.NoErr(err)
	is.OK(rc)

	b, err := ioutil.ReadAll(rc)
	is.NoErr(err)
	is.OK(b)

	is.Equal(string(b), contents)
}
