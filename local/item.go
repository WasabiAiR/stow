package local

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type item struct {
	name    string
	path    string
	md5once sync.Once // protects md5
	md5     string
}

func (i *item) ID() string {
	return i.path
}

func (i *item) Name() string {
	return i.name
}

func (i *item) URL() *url.URL {
	return &url.URL{
		Scheme: "file",
		Path:   filepath.Clean(i.path),
	}
}

func (i *item) ETag() (string, error) {
	info, err := os.Stat(i.path)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s:%v", i.path, info.ModTime().String()), nil
}

func (i *item) MD5() (string, error) {
	var err error
	i.md5once.Do(func() {
		var f io.ReadCloser
		f, err = i.Open()
		if err != nil {
			return
		}
		defer f.Close()
		h := md5.New()
		_, err = io.Copy(h, f)
		if err != nil {
			return
		}
		i.md5 = fmt.Sprintf("%x", h.Sum(nil))
	})
	return i.md5, err
}

// Open opens the file for reading.
func (i *item) Open() (io.ReadCloser, error) {
	return os.Open(i.path)
}
