package local_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
)

func TestItemsPaging(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()
	cfg := stow.ConfigMap{"path": testDir}
	l, err := stow.Dial(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	// get the first actual container to work with (not "All" container)
	containers, _, err := l.Containers("", stow.CursorStart, 10)
	is.NoErr(err)
	is.True(len(containers) > 0)
	container := containers[1]

	// make 25 items
	for i := 0; i < 25; i++ {
		_, err := container.Put(fmt.Sprintf("item-%02d", i), strings.NewReader(`item`), 4, nil)
		is.NoErr(err)
	}

	// get the first page
	items, cursor, err := container.Items("item-", stow.CursorStart, 10)
	is.NoErr(err)
	is.OK(items)
	is.Equal(len(items), 10)
	is.OK(cursor)
	is.Equal(cursor, "item-10")

	// get the next page
	items, cursor, err = container.Items("item-", cursor, 10)
	is.NoErr(err)
	is.OK(items)
	is.Equal(len(items), 10)
	is.OK(cursor)
	is.Equal(cursor, "item-20")

	// get the last page
	items, cursor, err = container.Items("item-", cursor, 10)
	is.NoErr(err)
	is.OK(items)
	is.Equal(len(items), 5)
	is.True(stow.IsCursorEnd(cursor))

	// bad cursor
	_, _, err = container.Items("item-", "made up cursor", 10)
	is.Equal(err, stow.ErrBadCursor)

}

type chunkOfZero int

func (c *chunkOfZero) Read(b []byte) (n int, err error) {
	if *c <= 0 {
		return 0, io.EOF
	}
	toFill := len(b)
	if toFill > int(*c) {
		toFill = int(*c)
	}
	leftBytes := int(*c) - toFill
	*c = chunkOfZero(leftBytes)
	for i := 0; i < toFill; i++ {
		b[i] = 0
	}
	return toFill, nil
}

func TestAtomicWrites(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()
	cfg := stow.ConfigMap{"path": testDir}
	l, err := stow.Dial(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	container, err := l.Container("one")
	is.NoErr(err)
	is.OK(container)

	initialVer := chunkOfZero(1000000)
	initialItem, err := container.Put("test-file", &initialVer, 1000000, nil)
	is.NoErr(err)
	sz, err := initialItem.Size()
	is.NoErr(err)
	is.True(sz == 1000000)

	end := make(chan error)
	go func() {
		for i := 0; i < 10; i++ {
			c := chunkOfZero(1000000)
			_, err := container.Put("test-file", &c, 1000000, nil)
			if err != nil {
				end <- err
			}
		}

		end <- nil
	}()

loop:
	for {
		items, _, err := container.Items("", "", 10)
		is.NoErr(err)
		is.True(len(items) < 10)
		found := false
		for _, i := range items {
			if i.Name() != "test-file" {
				continue
			}

			found = true
			sz, err := i.Size()
			is.NoErr(err)
			is.True(sz == 1000000)
		}
		is.True(found)

		select {
		case err := <-end:
			is.NoErr(err)
			break loop
		default:
		}
	}
}
