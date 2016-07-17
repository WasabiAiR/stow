
## Location = Amazon S3
## Container = Bucket
## Item = File

- AWS doesn't always use MD5 hashes for the ETag field.
- Paging for the list of containers doesn't exist yet.

TODO:
- refactor to add the owner field, better to not disregard it. Line 49 container.go is nil

---

###### Dev Notes

The init function of every implementation of `stow` must call `stow.Register`.

`stow.Register` accepts a few things: 

### Kind, a string argument respresenting the name of the location.

`makefn` a function that accepts any type that conforms to the stow.Config
interface. It first validates the values of the `Config` argument, and then
attempts to use the configuration to create a new client. If successful, An
instance of a data type that conforms to the `stow.Location` interface is
created. This Location should have fields that contain the client and
configuration.

Further calls in the hierarchy of a Location, Container, and Item depend
on the values of the configuration + the client to send and receive information.

- `kingmatchfn` a function that ensures that a given URL matches the `Kind` of the type of storage.

---

**stow.Register(kind string, makefn func(Config) (Locaion, error), kindmatchfn func(*url.URL) bool)**

- Adds `kind` and `makefn` into a map that contains a list of locations.

- Adds `kind` to a slice that contains all of the different kinds.

- Adds `kind` as part of an anonymous function which validates the scheme of the url.URL

Once the `stow.Register` function is completed, a location of the given.
---
