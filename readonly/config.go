package readonly

import (
	"errors"
	"net/url"

	"github.com/graymeta/stow"
)

const (
	// ConfigWrappedKind is a key whose value is the type of backend being wrapped.
	ConfigWrappedKind = "wrapped_kind"
	// ConfigWrapped is the key whose value is the config for the backend being wrapped.
	ConfigWrapped = "wrapped"
)

// Kind is the kid of Location this package provides
const Kind = "readonly"

func init() {
	validatefn := func(config stow.Config) error {
		wrappedKind, ok := config.Config(ConfigWrappedKind)
		if !ok || wrappedKind == "" {
			return errors.New("missing wrapped kind")
		}
		wrappedConfig, ok := config.NestedConfig(ConfigWrapped)
		if !ok {
			return errors.New("missing config for wrapped")
		}
		return stow.Validate(wrappedKind, wrappedConfig)
	}
	makefn := func(config stow.Config) (stow.Location, error) {
		wrappedKind, ok := config.Config(ConfigWrappedKind)
		if !ok || wrappedKind == "" {
			return nil, errors.New("missing wrapped kind")
		}
		wrappedConfig, ok := config.NestedConfig(ConfigWrapped)
		if !ok {
			return nil, errors.New("missing config for wrapped")
		}

		w, err := stow.Dial(wrappedKind, wrappedConfig)
		if err != nil {
			return nil, err
		}

		return &location{wrapped: w}, nil
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn, validatefn)
}
