package swift

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/graymeta/stow"
	"github.com/ncw/swift"
)

// Config key constants.
const (
	ConfigUsername      = "username"
	ConfigKey           = "key"
	ConfigTenantName    = "tenant_name"
	ConfigTenantAuthURL = "tenant_auth_url"

	ConfigApplicationCredentialID     = "application_credential_id"
	ConfigApplicationCredentialSecret = "application_credential_secret"
	ConfigApplicationCredentialName   = "application_credential_name"
)

// Kind is the kind of Location this package provides.
const Kind = "swift"

func validateConfig(config stow.Config) error {
	_, ok := config.Config(ConfigTenantAuthURL)
	if !ok {
		return errors.New("missing tenant auth url")
	}

	_, ok = config.Config(ConfigUsername)

	if ok {
		_, ok = config.Config(ConfigKey)
		if !ok {
			return errors.New("missing api key")
		}
		_, ok = config.Config(ConfigTenantName)
		if !ok {
			return errors.New("missing tenant name")
		}
		return nil
	}

	_, ok = config.Config(ConfigApplicationCredentialName)

	if ok {
		_, ok = config.Config(ConfigApplicationCredentialID)
		if !ok {
			return errors.New("missing application credential id")
		}
		_, ok = config.Config(ConfigApplicationCredentialSecret)
		if !ok {
			return errors.New("missing application credential secret")
		}
		return nil

	}

	return errors.New("missing config to auth with account or application credential")
}

func init() {
	validatefn := validateConfig
	makefn := func(config stow.Config) (stow.Location, error) {
		err := validateConfig(config)
		if err != nil {
			return nil, err
		}
		l := &location{
			config: config,
		}

		l.client, err = newSwiftClient(l.config)
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

func newSwiftClient(cfg stow.Config) (*swift.Connection, error) {

	var client *swift.Connection
	tenantAuthURL, _ := cfg.Config(ConfigTenantAuthURL)

	applicationCredentialName, ok := cfg.Config(ConfigApplicationCredentialName)
	if ok {
		applicationCredentialID, _ := cfg.Config(ConfigApplicationCredentialID)
		applicationCredentialSecret, _ := cfg.Config(ConfigApplicationCredentialSecret)

		client = &swift.Connection{
			AuthUrl:                     tenantAuthURL,
			ApplicationCredentialId:     applicationCredentialID,
			ApplicationCredentialName:   applicationCredentialName,
			ApplicationCredentialSecret: applicationCredentialSecret,
			Transport:                   http.DefaultTransport,
		}
	}

	username, ok := cfg.Config(ConfigUsername)
	if ok {
		key, _ := cfg.Config(ConfigKey)
		tenantName, _ := cfg.Config(ConfigTenantName)

		client = &swift.Connection{
			UserName: username,
			ApiKey:   key,
			AuthUrl:  tenantAuthURL,
			//Domain:   "domain", // Name of the domain (v3 auth only)
			Tenant: tenantName, // Name of the tenant (v2 auth only)
			// Add Default transport
			Transport: http.DefaultTransport,
		}
	}

	if client == nil {
		return nil, errors.New("unable to create connection")
	}

	err := client.Authenticate()
	if err != nil {
		return nil, errors.New("Unable to authenticate")
	}
	return client, nil
}
