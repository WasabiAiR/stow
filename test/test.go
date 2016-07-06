package test

import (
	"errors"
	"io/ioutil"
	"strings"
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

	// create two containers
	c1 := createContainer(is, location, "c1")
	c2 := createContainer(is, location, "c2")
	is.NotEqual(c1.ID(), c2.ID())

	// add three items to c1
	item1 := createItem(is, c1, "item1", "item one")
	item2 := createItem(is, c1, "item2", "item two")
	item3 := createItem(is, c1, "item3", "item three")
	is.OK(item1, item2, item3)

	// make sure we get these three items from the container
	items, more, err := c1.Items(0)
	is.NoErr(err)
	is.False(more)
	is.Equal(len(items), 3)

	// make sure the items are identical
	is.Equal(items[0].ID(), item1.ID())
	is.Equal(items[0].Name(), item1.Name())
	is.Equal(readItemContents(is, item1), "item one")

	is.Equal(items[1].ID(), item2.ID())
	is.Equal(items[1].Name(), item2.Name())
	is.Equal(readItemContents(is, item2), "item two")

	is.Equal(items[2].ID(), item3.ID())
	is.Equal(items[2].Name(), item3.Name())
	is.Equal(readItemContents(is, item3), "item three")

	// check MD5s
	is.Equal(len(md5(is, item1)), 32)
	is.Equal(len(md5(is, item2)), 32)
	is.Equal(len(md5(is, item3)), 32)

	// check ETags
	is.OK(etag(is, item1))
	is.OK(etag(is, item2))
	is.OK(etag(is, item3))

	// get items by URL
	u1 := item1.URL()
	item1b, err := location.ItemByURL(u1)
	is.NoErr(err)
	is.OK(item1b)
	is.Equal(item1b.ID(), item1.ID())
	is.Equal(etag(is, item1b), etag(is, item1))

	// test walking
	var walkedItems []stow.Item
	err = stow.Walk(c1, func(item stow.Item, page int, err error) error {
		if err != nil {
			return err
		}
		walkedItems = append(walkedItems, item)
		return nil
	})
	is.NoErr(err)
	is.Equal(len(walkedItems), 3)

	is.Equal(readItemContents(is, item1), "item one")
	is.Equal(readItemContents(is, item2), "item two")
	is.Equal(readItemContents(is, item3), "item three")

	// test walking error
	testErr := errors.New("test error")
	err = stow.Walk(c1, func(item stow.Item, page int, err error) error {
		return testErr
	})
	is.Equal(testErr, err)

}

func createContainer(is is.I, location stow.Location, name string) stow.Container {
	container, err := location.CreateContainer(name)
	is.NoErr(err)
	is.OK(container)
	is.OK(container.ID())
	is.Equal(container.Name(), name)
	return container
}

func createItem(is is.I, container stow.Container, name, content string) stow.Item {
	item, err := container.Put(name, strings.NewReader(content), int64(len(content)))
	is.NoErr(err)
	is.OK(item)
	return item
}

func readItemContents(is is.I, item stow.Item) string {
	r, err := item.Open()
	is.NoErr(err)
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	is.NoErr(err)
	return string(b)
}

func md5(is is.I, item stow.Item) string {
	md5, err := item.MD5()
	is.NoErr(err)
	return md5
}

func etag(is is.I, item stow.Item) string {
	etag, err := item.ETag()
	is.NoErr(err)
	return etag
}
