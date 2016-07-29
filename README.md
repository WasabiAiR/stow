# stow
Blob storage abstraction.

## How it works

Stow provides implementations for storage services, blob stores, cloud storage etc.

## Implementations

* Local (folders are containers, files are items)
* Amazon S3
* Microsoft Azure Blob Storage
* Openstack Swift (with auth v2)

## Concepts

The concepts of Stow are modelled around the most popular object storage services, and are made up of three main objects:

* Location - a place where many `Container` objects are stored
* Container - a named group of `Item` objects
* Item - an individual file

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

* A Location contains many containers
* A container contains many items
* Containers do not contain other containers
* Items must belong to a container

## Guides

### Getting an `Item` by URL

If you have a stow URL, you can use it to lookup the kind of location:

```
kind, err := stow.KindByURL(url)
```

`kind` will be a string describing the kind of storage. You can then pass `kind` along with a `Config` to `stow.New` to create a new `Location` where the item for the URL is:

```
location, err := stow.New(kind, config)
```

You can then get the `Item` for the specified URL from the location:

```
item, err := location.ItemByURL(url)
```

### Cursors

Cursors are strings that provide a pointer to items in sets allowing for paging over the entire set.

Call such methods first passing in `stow.CursorStart` as the cursor, which indicates the first item/page. The method will, as one of its return arguments, provide a new cursor which you can pass into subsequent calls to the same method.

When `stow.IsCursorEnd(cursor)` returns `true`, you have reached the end of the set.