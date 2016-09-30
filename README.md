# stow [![GoDoc](https://godoc.org/github.com/graymeta/stow?status.svg)](https://godoc.org/github.com/graymeta/stow) [![Go Report Card](https://goreportcard.com/badge/github.com/graymeta/stow)](https://goreportcard.com/report/github.com/graymeta/stow)
Cloud storage abstraction package for Go. 

* Version: 0.1.0
* Project status: Approaching v1.0 release

## How it works

Stow provides implementations for storage services, blob stores, cloud storage etc.

## Implementations

* Local (folders are containers, files are items)
* Remote (mounted) drives (SMB, NFS, CIFS, etc.)
* Amazon S3
* Google Cloud Storage
* Microsoft Azure Blob Storage
* Openstack Swift (with auth v2)
* Oracle Storage Cloud Service

## Concepts

The concepts of Stow are modelled around the most popular object storage services, and are made up of three main objects:

* `Location` - a place where many `Container` objects are stored
* `Container` - a named group of `Item` objects
* `Item` - an individual file

```
location1 (e.g. Azure)
├── container1
├───── item1.1
├───── item1.2
├───── item1.3
├── container2
├───── item2.1
├───── item2.2
location2 (e.g. local storage)
├── container1
├───── item1.1
├───── item1.2
├───── item1.3
├── container2
├───── item2.1
├───── item2.2
```

* A location contains many containers
* A container contains many items
* Containers do not contain other containers
* Items must belong to a container
* Item names may be a path

## Guides

### Using Stow

Import Stow plus any of the implementation packages that you wish to provide. For example, to support Google Cloud Storage and Amazon S3 you would write:

```go
import (
	"github.com/graymeta/stow"
	_ "github.com/graymeta/stow/google"
	_ "github.com/graymeta/stow/s3"
)
```

The underscore indicates that you do not intend to use the package in your code. Importing it is enough, as the implementation packages register themselves with Stow during initialization.

* For more information about using Stow, see the [Best practices documentation](BestPractices.md).

### Connecting to locations

To connect to a location, you need to know the `kind` string (available by accessing the `Kind` constant in the implementation package) and a `stow.Config` object that contains any required configuration information (such as account names, API keys, credentials, etc). Configuration is implementation specific, so you should consult each implementation to see what fields are required.

```go
kind := "s3"
config := stow.ConfigMap{
	s3.ConfigAccessKeyID: "246810"
	s3.ConfigSecretKey:   "abc123",
	s3.ConfigRegion:      "eu-west-1"
}
location, err := stow.Dial(kind, config)
if err != nil {
	return err
}
defer location.Close()

// TODO: use location
```

### Walking containers

You can walk every Container using the `stow.WalkContainers` function:

```go
func WalkContainers(location Location, prefix string, pageSize int, fn WalkContainersFunc) error
```

For example:

```go
err = stow.WalkContainers(location, stow.NoPrefix, 100, func(c stow.Container, err error) error {
	if err != nil {
		return err
	}
	switch c.Name() {
	case c1.Name(), c2.Name(), c3.Name():
		found++
	}
	return nil
})
if err != nil {
	return err
}
```

### Walking items

Once you have a `Container`, you can walk every Item inside it using the `stow.Walk` function:

```go
func Walk(container Container, prefix string, pageSize int, fn WalkFunc) error
```

For example:

```go
err = stow.Walk(containers[0], stow.NoPrefix, 100, func(item stow.Item, err error) error {
	if err != nil {
		return err
	}
	log.Println(item.Name())
	return nil
})
if err != nil {
	return err
}
```

### Stow URLs

An `Item` can return a URL via the `URL()` method. While a valid URL, they are useful only within the context of Stow. Within a Location, you can get items using these URLs via the `Location.ItemByURL` method.

#### Getting an `Item` by URL

If you have a Stow URL, you can use it to lookup the kind of location:

```go
kind, err := stow.KindByURL(url)
```

`kind` will be a string describing the kind of storage. You can then pass `kind` along with a `Config` to `stow.New` to create a new `Location` where the item for the URL is:

```go
location, err := stow.Dial(kind, config)
```

You can then get the `Item` for the specified URL from the location:

```go
item, err := location.ItemByURL(url)
```

### Cursors

Cursors are strings that provide a pointer to items in sets allowing for paging over the entire set.

Call such methods first passing in `stow.CursorStart` as the cursor, which indicates the first item/page. The method will, as one of its return arguments, provide a new cursor which you can pass into subsequent calls to the same method.

When `stow.IsCursorEnd(cursor)` returns `true`, you have reached the end of the set.
