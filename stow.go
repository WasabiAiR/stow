package stow

import "fmt"

// Locations is a map of installed location providers,
// supplying a function that creates a new instance of
// that Location.
var Locations = map[string]func(Config) Location{}

// Location represents a storage location.
type Location interface {
	// Containers gets the first page of containers
	// with the specified prefix from this Location.
	Containers(prefix string) (ContainerList, error)
	// Container gets the Container with the specified
	// identifier.
	Container(id string) (Container, error)
}

// New gets a new Location with the given kind and
// kind specific configuration.
func New(kind string, config Config) (Location, error) {
	for k, fn := range Locations {
		if k == kind {
			return fn(config), nil
		}
	}
	return nil, fmt.Errorf("stow: unknown kind %s", kind)
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
}

// ItemList represents a list of Item objects.
type ItemList interface {
	Items() []Item
	More() bool
}

// Item represents an item inside a Container.
// Such as a file.
type Item interface {
	// ID gets a unique string describing this Container.
	ID() string
	// Name gets a human-readable name describing this Container.
	Name() string
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

var _ Config = (ConfigMap)(nil)

// Config gets a string configuration value and a
// bool indicating whether the value was present or not.
func (c ConfigMap) Config(name string) (string, bool) {
	val, ok := c[name]
	return val, ok
}
