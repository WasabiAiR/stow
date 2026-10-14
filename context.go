package stow

import (
	"context"
	"io"
	"net/url"
	"time"
)

// DEV NOTE: tests for the implementations are in test/test.go

// The Location, Container and Item methods take no context, so a caller
// cannot cancel a request or give it a deadline. The interfaces below add a
// variant with a context to every method that talks to the storage, and are
// implemented by the locations, containers and items that can pass the
// context on to their requests.
//
// The functions of the same names call these methods when the value has
// them, and fall back to the method without a context otherwise. They are
// what calling code should use.

// ContextLocation is a Location whose requests take a context.
type ContextLocation interface {
	// CreateContainerContext is CreateContainer with a context.
	CreateContainerContext(ctx context.Context, name string) (Container, error)
	// ContainersContext is Containers with a context.
	ContainersContext(ctx context.Context, prefix, cursor string, count int) ([]Container, string, error)
	// ContainerContext is Container with a context.
	ContainerContext(ctx context.Context, id string) (Container, error)
	// RemoveContainerContext is RemoveContainer with a context.
	RemoveContainerContext(ctx context.Context, id string) error
	// ItemByURLContext is ItemByURL with a context.
	ItemByURLContext(ctx context.Context, url *url.URL) (Item, error)
}

// ContextContainer is a Container whose requests take a context.
type ContextContainer interface {
	// ItemContext is Item with a context.
	ItemContext(ctx context.Context, id string) (Item, error)
	// ItemsContext is Items with a context.
	ItemsContext(ctx context.Context, prefix, cursor string, count int) ([]Item, string, error)
	// RemoveItemContext is RemoveItem with a context.
	RemoveItemContext(ctx context.Context, id string) error
	// PutContext is Put with a context.
	PutContext(ctx context.Context, name string, r io.Reader, size int64, metadata map[string]any) (Item, error)
}

// ContextItem is an Item whose requests take a context.
type ContextItem interface {
	// OpenContext is Open with a context. The context also covers reading
	// from the returned io.ReadCloser, so it must not be cancelled before
	// the content is read.
	OpenContext(ctx context.Context) (io.ReadCloser, error)
	// ETagContext is ETag with a context.
	ETagContext(ctx context.Context) (string, error)
	// LastModContext is LastMod with a context.
	LastModContext(ctx context.Context) (time.Time, error)
	// MetadataContext is Metadata with a context.
	MetadataContext(ctx context.Context) (map[string]any, error)
}

// ContextItemRanger is an ItemRanger whose requests take a context.
type ContextItemRanger interface {
	// OpenRangeContext is OpenRange with a context. The context also covers
	// reading from the returned io.ReadCloser.
	OpenRangeContext(ctx context.Context, start, end uint64) (io.ReadCloser, error)
}

// ContextTaggable is a Taggable whose requests take a context.
type ContextTaggable interface {
	// TagsContext is Tags with a context.
	TagsContext(ctx context.Context) (map[string]any, error)
}

// CreateContainerContext creates a container like Location.CreateContainer.
// A location that is not a ContextLocation only has the context checked
// before the call.
func CreateContainerContext(ctx context.Context, l Location, name string) (Container, error) {
	if cl, ok := l.(ContextLocation); ok {
		return cl.CreateContainerContext(ctx, name)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return l.CreateContainer(name)
}

// ContainersContext gets a page of containers like Location.Containers.
// A location that is not a ContextLocation only has the context checked
// before the call.
func ContainersContext(ctx context.Context, l Location, prefix, cursor string, count int) ([]Container, string, error) {
	if cl, ok := l.(ContextLocation); ok {
		return cl.ContainersContext(ctx, prefix, cursor, count)
	}
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	return l.Containers(prefix, cursor, count)
}

// ContainerContext gets a container like Location.Container. A location that
// is not a ContextLocation only has the context checked before the call.
func ContainerContext(ctx context.Context, l Location, id string) (Container, error) {
	if cl, ok := l.(ContextLocation); ok {
		return cl.ContainerContext(ctx, id)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return l.Container(id)
}

// RemoveContainerContext removes a container like Location.RemoveContainer.
// A location that is not a ContextLocation only has the context checked
// before the call.
func RemoveContainerContext(ctx context.Context, l Location, id string) error {
	if cl, ok := l.(ContextLocation); ok {
		return cl.RemoveContainerContext(ctx, id)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return l.RemoveContainer(id)
}

// ItemByURLContext gets an item like Location.ItemByURL. A location that is
// not a ContextLocation only has the context checked before the call.
func ItemByURLContext(ctx context.Context, l Location, url *url.URL) (Item, error) {
	if cl, ok := l.(ContextLocation); ok {
		return cl.ItemByURLContext(ctx, url)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return l.ItemByURL(url)
}

// ItemContext gets an item like Container.Item. A container that is not a
// ContextContainer only has the context checked before the call.
func ItemContext(ctx context.Context, c Container, id string) (Item, error) {
	if cc, ok := c.(ContextContainer); ok {
		return cc.ItemContext(ctx, id)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.Item(id)
}

// ItemsContext gets a page of items like Container.Items. A container that
// is not a ContextContainer only has the context checked before the call.
func ItemsContext(ctx context.Context, c Container, prefix, cursor string, count int) ([]Item, string, error) {
	if cc, ok := c.(ContextContainer); ok {
		return cc.ItemsContext(ctx, prefix, cursor, count)
	}
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	return c.Items(prefix, cursor, count)
}

// RemoveItemContext removes an item like Container.RemoveItem. A container
// that is not a ContextContainer only has the context checked before the
// call.
func RemoveItemContext(ctx context.Context, c Container, id string) error {
	if cc, ok := c.(ContextContainer); ok {
		return cc.RemoveItemContext(ctx, id)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.RemoveItem(id)
}

// PutContext creates an item like Container.Put. A container that is not a
// ContextContainer only has the context checked before the call.
func PutContext(ctx context.Context, c Container, name string, r io.Reader, size int64, metadata map[string]any) (Item, error) {
	if cc, ok := c.(ContextContainer); ok {
		return cc.PutContext(ctx, name, r, size, metadata)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.Put(name, r, size, metadata)
}

// OpenContext opens an item like Item.Open. For a ContextItem the context
// also covers reading from the returned io.ReadCloser. An item that is not
// a ContextItem only has the context checked before the call.
func OpenContext(ctx context.Context, i Item) (io.ReadCloser, error) {
	if ci, ok := i.(ContextItem); ok {
		return ci.OpenContext(ctx)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return i.Open()
}

// ETagContext gets the ETag of an item like Item.ETag. An item that is not
// a ContextItem only has the context checked before the call.
func ETagContext(ctx context.Context, i Item) (string, error) {
	if ci, ok := i.(ContextItem); ok {
		return ci.ETagContext(ctx)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return i.ETag()
}

// LastModContext gets the last modified date of an item like Item.LastMod.
// An item that is not a ContextItem only has the context checked before the
// call.
func LastModContext(ctx context.Context, i Item) (time.Time, error) {
	if ci, ok := i.(ContextItem); ok {
		return ci.LastModContext(ctx)
	}
	if err := ctx.Err(); err != nil {
		return time.Time{}, err
	}
	return i.LastMod()
}

// MetadataContext gets the metadata of an item like Item.Metadata. An item
// that is not a ContextItem only has the context checked before the call.
func MetadataContext(ctx context.Context, i Item) (map[string]any, error) {
	if ci, ok := i.(ContextItem); ok {
		return ci.MetadataContext(ctx)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return i.Metadata()
}

// OpenRangeContext opens a part of an item like ItemRanger.OpenRange. For a
// ContextItemRanger the context also covers reading from the returned
// io.ReadCloser. An item that is not a ContextItemRanger only has the
// context checked before the call.
func OpenRangeContext(ctx context.Context, i ItemRanger, start, end uint64) (io.ReadCloser, error) {
	if ci, ok := i.(ContextItemRanger); ok {
		return ci.OpenRangeContext(ctx, start, end)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return i.OpenRange(start, end)
}

// TagsContext gets the tags of an item like Taggable.Tags. An item that is
// not a ContextTaggable only has the context checked before the call.
func TagsContext(ctx context.Context, i Taggable) (map[string]any, error) {
	if ci, ok := i.(ContextTaggable); ok {
		return ci.TagsContext(ctx)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return i.Tags()
}
