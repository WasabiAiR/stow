package remote

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type item struct {
	path     string
	infoOnce sync.Once // protects info
	info     os.FileInfo
	infoErr  error
}

func (i *item) ID() string {
	return i.path
}

func (i *item) Name() string {
	return filepath.Base(i.path)
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

// Open opens the file for reading.
func (i *item) Open() (io.ReadCloser, error) {
	return os.Open(i.path)
}

func (i *item) LastMod() (time.Time, error) {
	info, err := i.getInfo()
	if err != nil {
		return time.Time{}, nil
	}

	return info.ModTime(), nil
}