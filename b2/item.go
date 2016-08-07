package b2

import (
	"errors"
	"io"
	"net/url"
	"time"

	"github.com/graymeta/stow"
	"gopkg.in/kothar/go-backblaze.v0"
)

type item struct {
	id           string
	name         string
	size         int64
	lastModified time.Time
	bucket       *backblaze.Bucket
	hash         string

	//	container *container
	// url          url.URL
	// lastModified time.Time
}

var _ stow.Item = (*item)(nil)

func (i *item) ID() string {
	return i.id
}

func (i *item) Name() string {
	return i.name
}

func (i *item) URL() *url.URL {

	str, err := i.bucket.FileURL(i.name)
	if err != nil {
		return nil
	}

	url, _ := url.Parse(str)
	url.Scheme = Kind

	return url
}

func (i *item) Size() (int64, error) {
	return i.size, nil
}

func (i *item) Open() (io.ReadCloser, error) {
	_, r, err := i.bucket.DownloadFileByName(i.name)
	return r, err
}

func (i *item) ETag() (string, error) {
	return i.lastModified.String(), nil
}

func (i *item) MD5() (string, error) {
	// Call fileinfo...they give us a sha1, no md5
	if i.hash == "" {
		f, err := i.bucket.GetFileInfo(i.id)
		if err != nil {
			return "", err
		}
		i.hash = f.ContentSha1
	}
	return i.hash, nil
}

func (i *item) LastMod() (time.Time, error) {
	if i.lastModified.IsZero() {
		response, err := i.bucket.ListFileNames(i.name, 1)
		if err != nil {
			return time.Now(), err
		}

		if len(response.Files) != 1 {
			return time.Now(), errors.New("Unable to determine lastModified time for item")
		}

		i.lastModified = time.Unix(response.Files[0].UploadTimestamp/1000, 0)
	}

	return i.lastModified, nil
}
