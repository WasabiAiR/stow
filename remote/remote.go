package remote

import (
	"errors"
	"net/url"
	"path"

	"github.com/graymeta/stow"
)

// ConfigKeys are the supported configuration items for
// remote storage (NFS, CIFS, SAMBA).
const (
	ConfigKeySource  = "source"
	ConfigKeyTarget  = "target"
	ConfigKeyType    = "type"
	ConfigKeyOptions = "options"
)

// Kind is the kind of Location this package provides.
const Kind = "remote"

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {
		source, ok := config.Config(ConfigKeySource)
		if !ok {
			return nil, errors.New("missing source in config")
		}
		target, ok := config.Config(ConfigKeyTarget)
		if !ok {
			return nil, errors.New("missing target in config")
		}
		target = path.Clean(target)

		fstype, ok := config.Config(ConfigKeyType)
		if !ok {
			return nil, errors.New("missing type in config")
		}

		options, ok := config.Config(ConfigKeyOptions)
		if !ok {
			return nil, errors.New("missing options in config")
		}

		err := mount(source, target, fstype, options)
		if err != nil {
			return nil, err
		}
		return &location{
			config:   config,
			pagesize: 10,
		}, nil
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == "remote"
	}
	stow.Register(Kind, makefn, kindfn)
}
