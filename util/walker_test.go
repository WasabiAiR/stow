package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	_ "github.com/graymeta/stow/local"
)

func TestWalker(t *testing.T) {
	is := is.New(t)
	testDir, done, err := setupLocalStowage()
	is.NoErr(err)
	defer done()

	config := stow.ConfigMap{"path": testDir}
	location, err := stow.New("local", config)
	is.NoErr(err)
	containers, _, err := location.Containers("three", 0)
	is.NoErr(err)
	is.Equal(len(containers), 1)
	threeContainer := containers[0]

	walker := NewWalker(threeContainer)
	var items []stow.Item

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for {
			select {
			case item := <-walker.ItemChan():
				items = append(items, item)
			case err := <-walker.ErrChan():
				is.NoErr(err)
			case <-walker.DoneChan():
				return
			}
		}
	}()

	walker.Start()

	// wait for walking to finish
	<-doneChan

	is.Equal(len(items), 3)
	is.Equal(items[0].Name(), "item1")
	is.Equal(items[1].Name(), "item2")
	is.Equal(items[2].Name(), "item3")

}

func setupLocalStowage() (string, func() error, error) {
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

	// make testpath absolute
	absdir, err := filepath.Abs(dir)
	if err != nil {
		return dir, done, err
	}
	return absdir, done, nil
}
