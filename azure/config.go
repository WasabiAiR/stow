package azure

import (
	"errors"
	"net/url"

	"github.com/graymeta/stow"
)

// ConfigAccount and ConfigKey are the supported configuration items for
// Azure blob storage.
const (
	ConfigAccount = "account"
	ConfigKey     = "key"
)

// Kind is the kind of Location this package provides.
const Kind = "azure"

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {
		_, ok := config.Config(ConfigAccount)
		if !ok {
			return nil, errors.New("missing account id")
		}
		_, ok = config.Config(ConfigKey)
		if !ok {
			return nil, errors.New("missing auth key")
		}
		return &location{
			config: config,
		}, nil
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == "azure"
	}
	stow.Register(Kind, makefn, kindfn)
}
