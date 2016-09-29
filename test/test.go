package test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

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

	location, err := stow.Dial(kind, config)
	is.NoErr(err)
	is.OK(location)

	startFDs, fdsStart := totalNetFDs(t)
	defer func() {
		err := location.Close()
		is.NoErr(err)

		// Close Idle HTTP connectors before counting TCP File Descriptors
		if tr, ok := http.DefaultTransport.(interface {
			CloseIdleConnections()
		}); ok {
			tr.CloseIdleConnections()
		}
		endFDs, fds := totalNetFDs(t)
		diffFDs := endFDs - startFDs
		if diffFDs > 0 {
			t.Log("File Descriptors test start")
			t.Log(string(fdsStart))
			t.Log("File Descriptors test end")
			t.Log(string(fds))
			t.Fatalf("Leaked %d TCP file descriptor", diffFDs)
		}
	}()

	// create two containers
	c1 := createContainer(is, location, "stowtest"+randName(10))
	c2 := createContainer(is, location, "stowtest"+randName(10))
	c3 := createContainer(is, location, "stowtest"+randName(10))
	is.NotEqual(c1.ID(), c2.ID())

	defer func() {
		err := location.RemoveContainer(c1.ID())
		is.NoErr(err)
		err = location.RemoveContainer(c2.ID())
		is.NoErr(err)
		err = location.RemoveContainer(c3.ID())
		is.NoErr(err)
	}()

	// add three items to c1
	item1 := putItem(is, c1, "a_first/the item", "item one")
	item2 := putItem(is, c1, "a_second/the item", "item two")
	item3 := putItem(is, c1, "the_third/the item", "item three")
	is.OK(item1, item2, item3)

	defer func() {
		err := c1.RemoveItem(item1.ID())
		is.NoErr(err)
		err = c1.RemoveItem(item2.ID())
		is.NoErr(err)
		err = c1.RemoveItem(item3.ID())
		is.NoErr(err)
	}()

	// make sure we can get a small set of paginated results
	items, cursor, err := c1.Items(stow.NoPrefix, stow.CursorStart, 1)
	is.NoErr(err)
	is.Equal(len(items), 1)
	is.NotEqual(cursor, "")

	// get the items with a prefix (should only get 2)
	items, _, err = c1.Items("a_", stow.CursorStart, stow.NumberItems)
	is.NoErr(err)
	is.Equal(len(items), 2)

	// make sure we get these three items from the container
	items, _, err = c1.Items(stow.NoPrefix, stow.CursorStart, stow.NumberItems)
	is.NoErr(err)
	is.Equal(len(items), 3)

	// make sure the items are identical
	is.OK(item1.ID())
	is.OK(item1.Name())
	is.Equal(items[0].ID(), item1.ID())
	is.Equal(items[0].Name(), item1.Name())
	is.Equal(size(is, items[0]), 8)
	is.Equal(readItemContents(is, item1), "item one")
	is.NoErr(acceptableTime(t, is, items[0], item1))

	is.OK(item2.ID())
	is.OK(item2.Name())
	is.Equal(items[1].ID(), item2.ID())
	is.Equal(items[1].Name(), item2.Name())
	is.Equal(size(is, items[1]), 8)
	is.Equal(readItemContents(is, item2), "item two")
	is.NoErr(acceptableTime(t, is, items[1], item2))

	is.OK(item3.ID())
	is.OK(item3.Name())
	is.Equal(items[2].ID(), item3.ID())
	is.Equal(items[2].Name(), item3.Name())
	is.Equal(size(is, items[2]), 10)
	is.Equal(readItemContents(is, item3), "item three")
	is.NoErr(acceptableTime(t, is, items[2], item3))

	// check ETags from items retrieved by the Items() method
	is.OK(etag(t, is, items[0]))
	is.OK(etag(t, is, items[1]))
	is.OK(etag(t, is, items[2]))

	// get container by ID
	c1copy, err := location.Container(c1.ID())
	is.NoErr(err)
	is.OK(c1copy)
	is.Equal(c1copy.ID(), c1.ID())

	// get container that doesn't exist
	noContainer, err := location.Container(c1.ID() + "nope")
	is.Equal(stow.ErrNotFound, err)
	is.Nil(noContainer)

	// get item by ID
	item1copy, err := c1copy.Item(item1.ID())
	is.NoErr(err)
	is.OK(item1copy)
	is.Equal(item1copy.ID(), item1.ID())
	is.Equal(item1copy.Name(), item1.Name())
	is.Equal(size(is, item1copy), size(is, item1))
	is.Equal(readItemContents(is, item1copy), "item one")
	is.OK(etag(t, is, item1copy))

	// get an item by ID that doesn't exist
	noItem, err := c1copy.Item(item1.ID() + "nope")
	is.Equal(stow.ErrNotFound, err)
	is.Nil(noItem)

	// get items by URL
	u1 := item1.URL()

	item1b, err := location.ItemByURL(u1)
	is.NoErr(err)
	is.OK(item1b)
	is.Equal(item1b.ID(), item1.ID())
	is.Equal(etag(t, is, item1b), etag(t, is, item1copy))

	// **************************************************
	// Walking
	// **************************************************

	// test walking
	var walkedItems []stow.Item
	err = stow.Walk(c1, stow.NoPrefix, stow.NumberItems, func(item stow.Item, err error) error {
		if err != nil {
			return err
		}
		walkedItems = append(walkedItems, item)
		return nil
	})
	is.NoErr(err)
	is.Equal(len(walkedItems), 3)
	is.Equal(readItemContents(is, walkedItems[0]), "item one")
	is.Equal(readItemContents(is, walkedItems[1]), "item two")
	is.Equal(readItemContents(is, walkedItems[2]), "item three")

	// test walking with a prefix
	walkedItems = make([]stow.Item, 0)
	err = stow.Walk(c1, "a_", stow.NumberItems, func(item stow.Item, err error) error {
		if err != nil {
			return err
		}
		walkedItems = append(walkedItems, item)
		return nil
	})
	is.NoErr(err)
	is.Equal(len(walkedItems), 2)
	is.Equal(readItemContents(is, walkedItems[0]), "item one")
	is.Equal(readItemContents(is, walkedItems[1]), "item two")

	// test walking error
	testErr := errors.New("test error")
	err = stow.Walk(c1, stow.NoPrefix, stow.NumberItems, func(item stow.Item, err error) error {
		return testErr
	})
	is.Equal(testErr, err)

	// container walking
	found := 0
	err = stow.WalkContainers(location, stow.NoPrefix, stow.NumberItems, func(c stow.Container, err error) error {
		if err != nil {
			return err
		}
		switch c.Name() {
		case c1.Name(), c2.Name(), c3.Name():
			found++
		}
		return nil
	})
	is.NoErr(err)
	is.Equal(found, 3) // should find three items
}

func totalNetFDs(t *testing.T) (int, []byte) {
	lsof, err := exec.Command("lsof", "-b", "-n", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			t.Log(string(err.Stderr))
		}
		t.Errorf("error finding or running lsof: %s", err.Error())
	}
	return bytes.Count(lsof, []byte("TCP")), lsof
}

func createContainer(is is.I, location stow.Location, name string) stow.Container {
	container, err := location.CreateContainer(name)
	is.NoErr(err)
	is.OK(container)
	is.OK(container.ID())
	is.Equal(container.Name(), name)
	return container
}

func putItem(is is.I, container stow.Container, name, content string) stow.Item {
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

func etag(t *testing.T, is is.I, item stow.Item) string {

	etag, err := item.ETag()
	is.NoErr(err)

	t.Logf("ETag value: %s", etag)

	if strings.HasPrefix(etag, `W/`) {
		is.Failf(`Item(%s) Etag value (%s) contains weak string prefix W/`, item.Name(), etag)
	} else if strings.HasPrefix(etag, `"`) || strings.HasSuffix(etag, `"`) {
		is.Failf("Item(%s) Etag value (%s) contains quotation mark in prefix or suffix", item.Name(), etag)
	}
	return etag
}

func size(is is.I, item stow.Item) int64 {
	size, err := item.Size()
	is.NoErr(err)
	return size
}

func acceptableTime(t *testing.T, is is.I, item1, item2 stow.Item) error {
	item1LastMod, err := item1.LastMod()
	is.NoErr(err)

	item2LastMod, err := item2.LastMod()
	is.NoErr(err)

	timeDiff := item2LastMod.Sub(item1LastMod)

	threshold := time.Duration(8 * time.Second)

	if timeDiff > threshold {
		t.Logf("LastModified time for item1: %s", item1LastMod.String())
		t.Logf("LastModified time for item2: %s", item2LastMod.String())
		t.Logf("Difference: %s", timeDiff.String())

		return fmt.Errorf("last modified time exceeds threshold (%v)", threshold.Seconds())
	}

	return nil
}

func lastMod(is is.I, item stow.Item) time.Time {
	lastMod, err := item.LastMod()
	is.NoErr(err)
	return lastMod
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

func randName(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
