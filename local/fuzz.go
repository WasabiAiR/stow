// +build gofuzz

package local

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"crypto/md5"
	"encoding/hex"

	"github.com/dvyukov/go-fuzz/examples/fuzz"
	"github.com/graymeta/stow"
)

func Fuzz(data []byte) int {
	testDir, teardown, _ := setup()
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.New(Kind, cfg)
	if err != nil {
		panic("couldn't make new local stow")
	}

	containers, _, err := l.Containers("t", 0)
	if err != nil {
		panic("couldn't get containers from prefix")
	}
	c1 := containers[0]
	items, _, err := c1.Items(0)
	if err != nil {
		panic("couldn't get an item")
	}
	beforecount := len(items)

	_, w, err := c1.CreateItem("new_item")
	if err != nil {
		panic("couldn't create an item")
	}
	defer w.Close()
	_, err = bytes.NewReader(data).WriteTo(w)
	if err != nil {
		panic("couldn't write to write-closer")
	}

	// get the container again
	containers, _, err = l.Containers("t", 0)
	if err != nil {
		panic("couldn't get containers")
	}
	c1 = containers[0]
	items, _, err = c1.Items(0)
	if err != nil {
		panic("couldn't get items")
	}
	aftercount := len(items)

	if aftercount != beforecount+1 {
		panic("item count doesn't match")
	}

	// get new item
	calculatedMD5 := calcMD5(data)
	item := items[len(items)-1]
	md5, err := item.MD5()
	if err != nil {
		panic("couldn't get MD5")
	}
	if calculatedMD5 != md5 {
		panic("md5 doesn't match")
	}
	r, err := item.Open()
	if err != nil {
		panic("couldn't read contents")
	}
	defer r.Close()
	itemContents, err := ioutil.ReadAll(r)
	if !fuzz.DeepEqual(itemContents, data) {
		panic("contents doesn't match")
	}
	return 1
}

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

	// make testpath absolute
	absdir, err := filepath.Abs(dir)
	if err != nil {
		return dir, done, err
	}
	return absdir, done, nil
}

func calcMD5(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
