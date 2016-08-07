package b2

import (
	"io"
	"time"

	"github.com/graymeta/stow"
	"gopkg.in/kothar/go-backblaze.v0"
)

type container struct {
	bucket *backblaze.Bucket
}

var _ stow.Container = (*container)(nil)

func (c *container) ID() string {
	// Although backblaze does give an ID for buckets, some operations deal with bucket
	// names instead of the ID (specifically the B2.Bucket method). For that reason,
	// return name instead of ID. We can still use the id field internally when necessary
	return c.bucket.Name
}

func (c *container) Name() string {
	return c.bucket.Name
}

func (c *container) Item(id string) (stow.Item, error) {
	return c.getItem(id)
}

func (c *container) Items(prefix, cursor string) ([]stow.Item, string, error) {

	// At this time, there is no support for prefix in the B2 API. There is support for a cursor
	numitems := 10

	response, err := c.bucket.ListFileNames(cursor, numitems)
	if err != nil {
		return nil, "", err
	}

	items := make([]stow.Item, len(response.Files))

	for i, obj := range response.Files {
		items[i] = &item{
			id:           obj.ID,
			name:         obj.Name,
			size:         int64(obj.Size),
			lastModified: time.Unix(obj.UploadTimestamp/1000, 0),
			bucket:       c.bucket,
		}
	}

	marker := response.NextFileName

	return items, marker, nil
}

func (c *container) Put(name string, r io.Reader, size int64) (stow.Item, error) {

	file, err := c.bucket.UploadFile(name, nil, r)
	if err != nil {
		return nil, err
	}

	item := &item{
		id:     file.ID,
		name:   file.Name,
		size:   file.ContentLength,
		bucket: c.bucket,
	}

	return item, nil
}

func (c *container) RemoveItem(id string) error {
	item, err := c.getItem(id)
	if err != nil {
		return err
	}

	// TODO: files can have multiple versions in backblaze. It appears you have to delete
	// files one version at a time. We should loop over the *FileInfo result until it is nil
	_, err = c.bucket.DeleteFileVersion(item.name, item.id)

	return err
}

func (c *container) getItem(id string) (*item, error) {

	file, err := c.bucket.GetFileInfo(id)

	if err != nil {
		return nil, stow.ErrNotFound
	}

	item := &item{
		id:     file.ID,
		name:   file.Name,
		size:   file.ContentLength,
		bucket: c.bucket,
	}

	return item, nil
}
