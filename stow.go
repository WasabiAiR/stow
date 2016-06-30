package stow

import (
	"io"
	"net/url"
	"sync"
)

var (
	lock sync.RWMutex // protects locations and kindmatches
	// locations is a map of installed location providers,
	// supplying a function that creates a new instance of
	// that Location.
	locations = map[string]func(Config) Location{}
	// kindmatches is a slice of functions that take turns
	// trying to match the kind of Location for a given
	// URL. Functions return an empty string if it does not
	// match.
	kindmatches []func(*url.URL) string
)

// Location represents a storage location.
type Location interface {
	// CreateContainer creates a new Container with the
	// specified name.
	CreateContainer(name string) (Container, error)
	// Containers gets the first page of containers
	// with the specified prefix from this Location.
	Containers(prefix string) (ContainerList, error)
	// Container gets the Container with the specified
	// identifier.
	Container(id string) (Container, error)
	// ItemByURL gets an Item at this location with the
	// specified URL.
	ItemByURL(*url.URL) (Item, error)
	// NewContainer creates a container and returns it.
	// On error it returns nil and error.
	NewContainer(name string) (Container, error)
	// DeleteContainer deletes container with given name.
	DeleteContainer(name string) error
}

// Register adds a Location implementation, with two helper functions.
// makefn should make a Location with the given Config.
// kindmatchfn should inspect a URL and return whether it represents a Location
// of this kind or not. Code can call KindByURL to get a kind string
// for any given URL and all registered implementations will be consulted.
// Register is usually called in an implementation package's init method.
func Register(kind string, makefn func(Config) Location, kindmatchfn func(*url.URL) bool) {
	lock.Lock()
	defer lock.Unlock()
	locations[kind] = makefn
	kindmatches = append(kindmatches, func(u *url.URL) string {
		if kindmatchfn(u) {
			return kind // match
		}
		return "" // empty string means no match
	})
}

// New gets a new Location with the given kind and
// configuration.
func New(kind string, config Config) (Location, error) {
	fn, ok := locations[kind]
	if !ok {
		return nil, errUnknownKind(kind)
	}
	return fn(config), nil
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

// ContainerList represents a list of containers.
type ContainerList interface {
	// Items are the containers.
	Items() []Container
	// More indicates whether there are more containers
	// not included in this ContainerList.
	More() bool
}

// Container represents a container.
type Container interface {
	// ID gets a unique string describing this Container.
	ID() string
	// Name gets a human-readable name describing this Container.
	Name() string
	// Items gets the first page of items for this
	// Container.
	Items() (ItemList, error)
	// CreateItem creates a new Item with the
	// specified name and returns it along with a io.WriteCloser
	// which can be used to write contents to the Item.
	// The io.WriteCloser must always be closed.
	CreateItem(name string) (Item, io.WriteCloser, error)
}

// ItemList represents a list of Item objects.
type ItemList interface {
	// Items gets the slice of Item that make up
	// this ItemList.
	Items() []Item
	// More indicates whether there are more items
	// after this ItemList.
	More() bool
}

// Item represents an item inside a Container.
// Such as a file.
type Item interface {
	// ID gets a unique string describing this Item.
	ID() string
	// Name gets a human-readable name describing this Item.
	Name() string
	// URL gets the url for this item.
	URL() *url.URL
	// Open opens the Item for reading.
	// Calling code must close the io.ReadCloser.
	Open() (io.ReadCloser, error)
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
