package azure

import (
	"errors"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	az "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/flyteorg/stow"
)

// ConfigAccount should be the name of your storage account in the Azure portal
// ConfigKey should be an access key
// ConfigBaseUrl is the base URL of the cloud you want to connect to. The default
// is Azure Public cloud
// ConfigDefaultAPIVersion is the Azure Storage API version string used when a
// client is created.
// ConfigUseHTTPS specifies whether you want to use HTTPS to connect
const (
	ConfigAccount           = "account"
	ConfigKey               = "key"
	ConfigBaseUrl           = "base_url"
	ConfigDefaultAPIVersion = "default_api_version"
	ConfigUseHTTPS          = "use_https"
)

// Kind is the kind of Location this package provides.
const Kind = "azure"

func init() {
	validatefn := func(config stow.Config) error {
		_, ok := config.Config(ConfigAccount)
		if !ok {
			return errors.New("missing account id")
		}
		_, ok = config.Config(ConfigKey)
		if !ok {
			return errors.New("missing auth key")
		}
		return nil
	}
	makefn := func(config stow.Config) (stow.Location, error) {
		_, ok := config.Config(ConfigAccount)
		if !ok {
			return nil, errors.New("missing account id")
		}
		_, ok = config.Config(ConfigKey)
		if !ok {
			return nil, errors.New("missing auth key")
		}
		l := &location{
			config: config,
		}

		acc, key, env, api, https, err := getAccount(l.config)
		if err != nil {
			return nil, err
		}

		l.account = acc

		l.client, err = newBlobStorageClient(acc, key, env, api, https)
		if err != nil {
			return nil, err
		}

		l.sharedCreds, err = azblob.NewSharedKeyCredential(acc, key)
		if err != nil {
			return nil, err
		}

		// test the connection
		_, _, err = l.Containers("", stow.CursorStart, 1)
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

func getAccount(cfg stow.Config) (account, key string, baseUrl string, defaultAPIVersion string, useHTTPS bool, err error) {
	acc, ok := cfg.Config(ConfigAccount)
	if !ok {
		return "", "", "", "", false, errors.New("missing account id")
	}

	key, ok = cfg.Config(ConfigKey)
	if !ok {
		return "", "", "", "", false, errors.New("missing auth key")
	}

	baseUrl, ok = cfg.Config(ConfigBaseUrl)
	if !ok {
		baseUrl = "core.windows.net"
	}
	defaultAPIVersion, ok = cfg.Config(ConfigDefaultAPIVersion)
	if !ok {
		defaultAPIVersion = "2018-03-28"
	}

	var useHTTPSStr string
	useHTTPSStr, ok = cfg.Config(ConfigUseHTTPS)
	if !ok {
		useHTTPSStr = ""
	}
	switch useHTTPSStr {
	case "true":
		useHTTPS = true
	case "false":
		useHTTPS = false
	default:
		useHTTPS = true
	}
	return acc, key, baseUrl, defaultAPIVersion, useHTTPS, nil
}

func newBlobStorageClient(account, key string, baseUrl string, defaultAPIVersion string, useHTTPS bool) (*az.BlobStorageClient, error) {
	basicClient, err := az.NewClient(account, key, baseUrl, defaultAPIVersion, useHTTPS)
	if err != nil {
		return nil, errors.New("bad credentials")
	}

	client := basicClient.GetBlobService()
	return &client, err
}
