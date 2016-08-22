package swift

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/graymeta/stow"
	"github.com/ncw/swift"
)

const (
	// ConfigUsername is the username associated with the account
	ConfigUsername = "username"

	// ConfigPassword is the user password associated with the account
	ConfigPassword = "password"

	// ConfigIdentityDomain is the identity domain associated with the account
	ConfigIdentityDomain = "identity_domain"
)

// Kind is the kind of Location this package provides.
const Kind = "oracle"

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {
		_, ok := config.Config(ConfigUsername)
		if !ok {
			return nil, errors.New("missing account username")
		}

		_, ok = config.Config(ConfigPassword)
		if !ok {
			return nil, errors.New("missing account password")
		}

		_, ok = config.Config(ConfigIdentityDomain)
		if !ok {
			return nil, errors.New("missing identity domain")
		}

		l := &location{
			config: config,
		}

		var err error
		l.client, err = newSwiftClient(l.config)
		if err != nil {
			return nil, err
		}

		return l, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn)
}

func newSwiftClient(cfg stow.Config) (*swift.Connection, error) {
	cfgUsername, _ := cfg.Config(ConfigUsername)

	// The client's key is the user account's Password
	swiftKey, _ := cfg.Config(ConfigPassword)

	// The client's tenant field is the container's Identity Domain
	swiftTenantName, _ := cfg.Config(ConfigIdentityDomain)

	// The username field is a combo of the user's email + resource type + identity domain
	// storage-foobarbaz:santaclaus@northpole.com
	swiftUsername := strings.Join([]string{"storage-", swiftTenantName, ":", cfgUsername}, "")

	// The Package's auth URL includes a portion of the API endpoint.
	swiftAuthURL := fmt.Sprintf(`https://storage-%s.storage.oraclecloud.com/auth/v1.0`,
		swiftTenantName)

	client := swift.Connection{
		UserName: swiftUsername,
		ApiKey:   swiftKey,
		AuthUrl:  swiftAuthURL,
		Tenant:   swiftTenantName,
	}

	err := client.Authenticate()
	if err != nil {
		return nil, errors.New("Unable to authenticate")
	}
	return &client, nil
}
