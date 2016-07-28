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
	name     string
	path     string
	infoOnce sync.Once // protects info
	info     os.FileInfo
	infoErr  error
	md5once  sync.Once // protects md5
	md5      string
}

func (i *item) ID() string {
	return i.path
}

func (i *item) Name() string {
	return i.name
}

func (i *item) Size() (int64, error) {
	info, err := i.getInfo()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (i *item) URL() *url.URL {
	return &url.URL{
		Scheme: "file",
		Path:   filepath.Clean(i.path),
	}
}

func (i *item) getInfo() (os.FileInfo, error) {
	i.infoOnce.Do(func() {
		i.info, i.infoErr = os.Stat(i.path)
	})
	return i.info, i.infoErr
}

func (i *item) ETag() (string, error) {
	info, err := i.getInfo()
	if err != nil {
		return "", nil
	}
	return info.ModTime().String(), nil
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
