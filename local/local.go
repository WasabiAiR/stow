package local

import (
	"net/url"

	"github.com/graymeta/stow"
)

// ConfigKeys are the supported configuration items for
// local storage.
const (
	ConfigKeyPath = "path"
)

// Kind is the kind of Location this package provides.
const Kind = "local"

const (
	paramTypeValue = "item"
)

func init() {
	makefn := func(config stow.Config) stow.Location {
		return &location{
			config: config,
		}
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == "file"
	}
	stow.Register(Kind, makefn, kindfn)
}
