# stow
Storage abstraction

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