package local_test

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
)

func setup() (string, func() error, error) {
	done := func() error { return nil } // noop
	dir, err := ioutil.TempDir("testdata", "stow")
	if err != nil {
		return dir, done, err
	}
	done = func() error {
		return os.RemoveAll(dir)
	}
	// add some "containers"
	err = os.Mkdir(filepath.Join(dir, "one"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = os.Mkdir(filepath.Join(dir, "two"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = os.Mkdir(filepath.Join(dir, "three"), 0777)
	if err != nil {
		return dir, done, err
	}

	// add three items
	err = ioutil.WriteFile(filepath.Join(dir, "three", "item1"), []byte("3.1"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "three", "item2"), []byte("3.2"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "three", "item3"), []byte("3.3"), 0777)
	if err != nil {
		return dir, done, err
	}

	return dir, done, nil
}

func TestContainers(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.Containers("")
	is.NoErr(err)
	is.OK(c)
	is.False(c.More())

	items := c.Items()
	is.Equal(len(items), 3)
	isDir(is, items[0].ID())
	is.Equal(items[0].Name(), "one")
	isDir(is, items[1].ID())
	is.Equal(items[1].Name(), "three")
	isDir(is, items[2].ID())
	is.Equal(items[2].Name(), "two")
}

func TestNewContainer(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.NewContainer("new_test_container")
	is.NoErr(err)
	is.OK(c)

	cc, err := l.Containers("new")
	is.NoErr(err)
	is.OK(cc)

	items := cc.Items()
	is.Equal(len(items), 1)
	isDir(is, items[0].ID())
	is.Equal(items[0].Name(), "new_test_container")
}

func TestDeleteContainer(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.NewContainer("delete_me_plox")
	is.NoErr(err)
	is.OK(c)

	cc, err := l.Containers("del")
	is.NoErr(err)
	is.OK(cc)

	items := cc.Items()
	is.Equal(len(items), 1)
	isDir(is, items[0].ID())
	is.Equal(items[0].Name(), "delete_me_plox")

	err = l.DeleteContainer("delete_me_plox")
	is.NoErr(err)
	check, err := l.Containers("del")
	is.NoErr(err)
	is.OK(check)

	items = check.Items()
	is.Equal(len(items), 0)
}

func TestContainersPrefix(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.Containers("t")
	is.NoErr(err)
	is.OK(c)

	items := c.Items()
	is.Equal(len(items), 2)
	isDir(is, items[0].ID())
	is.Equal(items[0].Name(), "three")
	isDir(is, items[1].ID())
	is.Equal(items[1].Name(), "two")

	cthree, err := l.Container(items[0].ID())
	is.NoErr(err)
	is.Equal(cthree.Name(), "three")
}

func TestContainer(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.Containers("t")
	is.NoErr(err)
	is.OK(c)

	items := c.Items()
	is.Equal(len(items), 2)
	isDir(is, items[0].ID())

	cthree, err := l.Container(items[0].ID())
	is.NoErr(err)
	is.Equal(cthree.Name(), "three")
}

func TestNewContainer(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.CreateContainer("new_test_container")
	is.NoErr(err)
	is.OK(c)

	cc, err := l.Containers("new")
	is.NoErr(err)
	is.OK(cc)

	items := cc.Items()
	is.Equal(len(items), 1)
	isDir(is, items[0].ID())
	is.Equal(items[0].Name(), "new_test_container")
}

func TestCreateItem(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}
	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.Containers("t")
	is.NoErr(err)
	is.OK(c)
	c1 := c.Items()[0]
	items, err := c1.Items()
	is.NoErr(err)
	beforecount := len(items.Items())

	newitem, w, err := c1.CreateItem("new_item")
	is.NoErr(err)
	defer w.Close()
	_, err = io.WriteString(w, "new item contents")
	is.NoErr(err)
	is.OK(newitem)
	is.Equal(newitem.Name(), "new_item")

	// get the container again
	c, err = l.Containers("t")
	is.NoErr(err)
	is.OK(c)
	c1 = c.Items()[0]
	items, err = c1.Items()
	is.NoErr(err)
	aftercount := len(items.Items())

	is.Equal(aftercount, beforecount+1)

	// get new item
	item := items.Items()[len(items.Items())-1]
	r, err := item.Open()
	is.NoErr(err)
	defer r.Close()
	itemContents, err := ioutil.ReadAll(r)
	is.NoErr(err)
	is.Equal("new item contents", string(itemContents))
}

func TestItems(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.Containers("t")
	is.NoErr(err)
	is.OK(c)
	three, err := l.Container(c.Items()[0].ID())
	is.NoErr(err)
	threeItemsPage, err := three.Items()
	is.NoErr(err)
	is.OK(threeItemsPage)

	items := threeItemsPage.Items()
	is.Equal(len(items), 3)
	is.Equal(items[0].ID(), filepath.Join(c.Items()[0].ID(), "item1"))
	is.Equal(items[0].Name(), "item1")
}

func TestItemByURL(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	c, err := l.Containers("t")
	is.NoErr(err)
	is.OK(c)
	three, err := l.Container(c.Items()[0].ID())
	is.NoErr(err)
	threeItemsPage, err := three.Items()
	is.NoErr(err)
	is.OK(threeItemsPage)

	items := threeItemsPage.Items()
	is.Equal(len(items), 3)

	item1 := items[0]

	// make sure we know the kind by URL
	kind, err := stow.KindByURL(item1.URL())
	is.NoErr(err)
	is.Equal(kind, local.Kind)

	i, err := l.ItemByURL(item1.URL())
	is.NoErr(err)
	is.OK(i)
	is.Equal(i.ID(), item1.ID())
	is.Equal(i.Name(), item1.Name())
	is.Equal(i.URL().String(), item1.URL().String())

}

func TestItemReader(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}
	l, err := stow.New(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)
	c, err := l.Containers("t")
	is.NoErr(err)
	is.OK(c)
	three, err := l.Container(c.Items()[0].ID())

	threeItemsPage, err := three.Items()
	is.NoErr(err)
	items := threeItemsPage.Items()
	item1 := items[0]

	rc, err := item1.Open()
	defer rc.Close()
	is.NoErr(err)
	b, err := ioutil.ReadAll(rc)
	is.NoErr(err)
	is.Equal("3.1", string(b))

}

func isDir(is is.I, path string) {
	info, err := os.Stat(path)
	is.NoErr(err)
	is.True(info.IsDir())
}
