package sftp

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/graymeta/stow"
)

type container struct {
	name     string
	location *location
}

// ID returns a string value which represents the name of the container.
func (c *container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *container) Name() string {
	return c.name
}

// Item returns a stow.Item instance of a container based on the name of the
// container and the file.
func (c *container) Item(id string) (stow.Item, error) {
	p := path.Join(c.location.config.basePath, c.name, id)
	info, err := c.location.sftpClient.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, stow.ErrNotFound
		}
		return nil, err
	}

	if info.IsDir() {
		return nil, errors.New("unexpected directory")
	}

	return &item{
		container: c,
		path:      id,
		size:      info.Size(),
		modTime:   info.ModTime(),
		md:        getFileMetadata(info),
	}, nil
}

// Items sends a request to retrieve a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	var entries []entry
	entries, cursor, err := c.getFolderItems([]entry{}, prefix, "", path.Join(c.location.config.basePath, c.name), cursor, count, false)
	if err != nil {
		return nil, "", err
	}

	// Convert entries to []stow.Item
	sItems := make([]stow.Item, len(entries))
	for i := range entries {
		sItems[i] = entries[i].item
	}

	return sItems, cursor, nil
}

const separator = "/"

// we use this struct to keep track of stuff when walking.
type entry struct {
	Name    string
	ID      string
	RelPath string
	item    stow.Item
}

func (c *container) getFolderItems(entries []entry, prefix, relPath, id, cursor string, limit int, initialStart bool) ([]entry, string, error) {
	relCursor := cursor
	if relPath != "" {
		relCursor = strings.TrimPrefix(cursor, relPath+separator)
	}
	cursorPieces := strings.Split(relCursor, separator)
	files, err := c.location.sftpClient.ReadDir(id)
	if err != nil {
		return nil, "", err
	}

	var start bool
	if cursorPieces[0] == "" {
		start = true
	}

	for i, file := range files {
		fileRelPath := path.Join(relPath, file.Name())
		if !start && cursorPieces[0] != "" && (file.Name() >= cursorPieces[0] || fileRelPath >= cursor) {
			start = true
			if file.Name() == cursorPieces[0] && !file.IsDir() {
				continue
			}
		}

		if !start {
			continue
		}

		if file.IsDir() {
			var err error
			var retCursor string
			entries, retCursor, err = c.getFolderItems(entries, prefix, fileRelPath, path.Join(id, file.Name()), cursor, limit, true)
			if err != nil {
				return nil, "", err
			}
			if len(entries) != limit {
				continue
			}

			if retCursor == "" && i < len(files)-1 {
				retCursor = entries[len(entries)-1].RelPath
			}

			return entries, retCursor, nil
		}

		// TODO: prefix could be optimized to not look down paths that don't match,
		// but this is a quick/cheap first implementation.
		filePath := strings.TrimPrefix(path.Join(id, file.Name()), path.Join(c.location.config.basePath, c.name)+"/")
		if !strings.HasPrefix(filePath, prefix) {
			continue
		}

		entries = append(
			entries,
			entry{
				Name:    file.Name(),
				ID:      path.Join(id, file.Name()),
				RelPath: fileRelPath,
				item: &item{
					container: c,
					path:      filePath,
					size:      file.Size(),
					modTime:   file.ModTime(),
					md:        getFileMetadata(file),
				},
			},
		)
		if len(entries) == limit {
			retCursor := entries[len(entries)-1].RelPath
			if i == len(files)-1 {
				retCursor = ""
			}
			return entries, retCursor, nil
		}
	}

	return entries, "", nil
}

// RemoveItem removes a file from the remote server.
func (c *container) RemoveItem(id string) error {
	return c.location.sftpClient.Remove(path.Join(c.location.config.basePath, c.name, id))
}

// Put sends a request to upload content to the container.
func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	if len(metadata) > 0 {
		return nil, stow.NotSupported("metadata")
	}

	p := path.Join(c.location.config.basePath, c.name, name)
	item := &item{
		container: c,
		path:      name,
		size:      size,
	}
	err := c.location.sftpClient.MkdirAll(path.Dir(p))
	if err != nil {
		return nil, err
	}
	f, err := c.location.sftpClient.Create(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, errors.New("bad size")
	}

	info, err := c.location.sftpClient.Stat(p)
	if err != nil {
		return nil, err
	}
	item.modTime = info.ModTime()
	item.md = getFileMetadata(info)

	return item, nil
}
