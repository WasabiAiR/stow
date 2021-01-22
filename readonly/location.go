package readonly

import (
	"errors"
	"net/url"

	"github.com/graymeta/stow"
)

type location struct {
	wrapped stow.Location
}

// Closes the underlying location.
func (l *location) Close() error {
	return l.wrapped.Close()
}

// CreateContainer creates a new Container with the
// specified name.
func (l *location) CreateContainer(name string) (stow.Container, error) {
	return nil, errors.New("readonly")
}

// Containers gets a page of containers
// with the specified prefix from this Location.
// The specified cursor is a pointer to the start of
// the containers to get. It it obtained from a previous
// call to this method, or should be CursorStart for the
// first page.
// count is the number of items to return per page.
// The returned cursor can be checked with IsCursorEnd to
// decide if there are any more items or not.
func (l *location) Containers(prefix string, cursor string, count int) ([]stow.Container, string, error) {
	cs, startAfter, err := l.wrapped.Containers(prefix, cursor, count)
	if err != nil {
		return nil, "", err
	}

	wrapped := make([]stow.Container, len(cs))
	for i, val := range cs {
		wrapped[i] = &container{wrapped: val}
	}

	return wrapped, startAfter, nil
}

// Container gets the Container with the specified
// identifier.
func (l *location) Container(id string) (stow.Container, error) {
	c, err := l.wrapped.Container(id)
	if err != nil {
		return nil, err
	}

	return &container{wrapped: c}, nil
}

// RemoveContainer removes the container with the specified ID.
func (l *location) RemoveContainer(id string) error {
	return errors.New("readonly")
}

// ItemByURL gets an Item at this location with the
// specified URL.
func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	i, err := l.wrapped.ItemByURL(url)
	if err != nil {
		return nil, err
	}

	return &item{wrapped: i}, nil
}
