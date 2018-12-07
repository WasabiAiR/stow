package b2

import (
	"errors"
	"net/url"

	"github.com/graymeta/stow"
	"gopkg.in/kothar/go-backblaze.v0"
)

// Config key constants.
const (
	ConfigAccountID      = "account_id"
	ConfigApplicationKey = "application_key"
	HashType             = "sha1"
)

// Kind is the kind of Location this package provides.
const Kind = "b2"

func init() {
	validatefn := func(config stow.Config) error {
		_, ok := config.Config(ConfigAccountID)
		if !ok {
			return errors.New("missing account id")
		}
		_, ok = config.Config(ConfigApplicationKey)
		if !ok {
			return errors.New("missing application key")
		}
		return nil
	}
	makefn := func(config stow.Config) (stow.Location, error) {
		_, ok := config.Config(ConfigAccountID)
		if !ok {
			return nil, errors.New("missing account ID")
		}
		_, ok = config.Config(ConfigApplicationKey)
		if !ok {
			return nil, errors.New("missing application key")
		}
		l := &location{
			config: config,
		}
		var err error
		l.client, err = newB2Client(l.config)
		if err != nil {
			return nil, err
		}
		return l, nil
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}
	stow.Register(Kind, makefn, kindfn, validatefn)
}

func newB2Client(cfg stow.Config) (*backblaze.B2, error) {
	accountID, _ := cfg.Config(ConfigAccountID)
	applicationKey, _ := cfg.Config(ConfigApplicationKey)

	client, err := backblaze.NewB2(backblaze.Credentials{
		AccountID:      accountID,
		ApplicationKey: applicationKey,
	})

	if err != nil {
		return nil, errors.New("Unable to create client")
	}

	err = client.AuthorizeAccount()
	if err != nil {
		return nil, errors.New("Unable to authenticate")
	}

	return client, nil
}
