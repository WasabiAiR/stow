package stow

import (
	"errors"
	"io"
	"net/url"
	"sync"
)

var (
	lock sync.RWMutex // protects locations and kindmatches
	// kinds holds a list of location kinds.
	kinds = []string{}
	// locations is a map of installed location providers,
	// supplying a function that creates a new instance of
	// that Location.
	locations = map[string]func(Config) (Location, error){}
	// kindmatches is a slice of functions that take turns
	// trying to match the kind of Location for a given
	// URL. Functions return an empty string if it does not
	// match.
	kindmatches []func(*url.URL) string
)

var (
	// ErrNotFound is returned when something could not be found.
	ErrNotFound = errors.New("not found")
)

// Location represents a storage location.
type Location interface {
	// CreateContainer creates a new Container with the
	// specified name.
	CreateContainer(name string) (Container, error)
	// Containers gets a page of containers
	// with the specified prefix from this Location.
	// The page starts at zero.
	// The bool returned indicates whether there might be
	// another page of results or not.
	// If false, there definitely are no more containers.
	Containers(prefix string, page int) ([]Container, bool, error)
	// Container gets the Container with the specified
	// identifier.
	Container(id string) (Container, error)
	// ItemByURL gets an Item at this location with the
	// specified URL.
	ItemByURL(url *url.URL) (Item, error)
}

// Register adds a Location implementation, with two helper functions.
// makefn should make a Location with the given Config.
// kindmatchfn should inspect a URL and return whether it represents a Location
// of this kind or not. Code can call KindByURL to get a kind string
// for any given URL and all registered implementations will be consulted.
// Register is usually called in an implementation package's init method.
func Register(kind string, makefn func(Config) (Location, error), kindmatchfn func(*url.URL) bool) {
	lock.Lock()
	defer lock.Unlock()
	locations[kind] = makefn
	kinds = append(kinds, kind)
	kindmatches = append(kindmatches, func(u *url.URL) string {
		if kindmatchfn(u) {
			return kind // match
		}
		return "" // empty string means no match
	})
}

// Dial gets a new Location with the given kind and
// configuration.
func Dial(kind string, config Config) (Location, error) {
	fn, ok := locations[kind]
	if !ok {
		return nil, errUnknownKind(kind)
	}
	return fn(config)
}

// Kinds gets a list of installed location kinds.
func Kinds() []string {
	return kinds
}

// KindByURL gets the kind represented by the given URL.
// It consults all registered locations.
// Error returned if no match is found.
func KindByURL(u *url.URL) (string, error) {
	lock.RLock()
	defer lock.RUnlock()
	for _, fn := range kindmatches {
		kind := fn(u)
		if kind == "" {
			continue
		}
		return kind, nil
	}
	return "", errUnknownKind("")
}

// Container represents a container.
type Container interface {
	// ID gets a unique string describing this Container.
	ID() string
	// Name gets a human-readable name describing this Container.
	Name() string
	// Items gets a page of items for this
	// Container. The page starts at zero.
	// The returned bool indicates whether there might be another
	// page of items or not. If false, there definitely are no more items.
	Items(page int) ([]Item, bool, error)
	// Put creates a new Item with the specified name, and contents
	// read from the reader.
	Put(name string, r io.Reader, size int64) (Item, error)
}

// Item represents an item inside a Container.
// Such as a file.
type Item interface {
	// ID gets a unique string describing this Item.
	ID() string
	// Name gets a human-readable name describing this Item.
	Name() string
	// URL gets a URL for this item.
	// For example:
	// local: file:///path/to/something
	// azure: azure://host:port/api/something
	//    s3: s3://host:post/etc
	URL() *url.URL
	// Open opens the Item for reading.
	// Calling code must close the io.ReadCloser.
	Open() (io.ReadCloser, error)
	// ETag is a string that is different when the Item is
	// different, and the same when the item is the same.
	// Usually this is the last modified datetime.
	ETag() (string, error)
	// MD5 gets a hash of the contents of the file.
	MD5() (string, error)
}

// Config represents key/value configuraiton.
type Config interface {
	// Config gets a string configuration value and a
	// bool indicating whether the value was present or not.
	Config(name string) (string, bool)
}

// ConfigMap is a map[string]string that implements
// the Config method.
type ConfigMap map[string]string

// Config gets a string configuration value and a
// bool indicating whether the value was present or not.
func (c ConfigMap) Config(name string) (string, bool) {
	val, ok := c[name]
	return val, ok
}
