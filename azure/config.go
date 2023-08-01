package azure

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/go-autorest/autorest/azure"

	az "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/flyteorg/stow"
)

// ConfigAccount and ConfigKey are the supported configuration items for
// Azure blob storage.
// SovereignCloud only supports "us" but there are other options listed
// https://pkg.go.dev/github.com/Azure/go-autorest/autorest/azure#Environment
const (
	ConfigAccount = "account"
	ConfigKey     = "key"
	ConfigCloud   = "cloud"
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

		acc, key, env, err := getAccount(l.config)
		if err != nil {
			return nil, err
		}

		l.account = acc

		l.client, err = newBlobStorageClient(acc, key, env)
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

func getAccount(cfg stow.Config) (account, key string, env azure.Environment, err error) {
	acc, ok := cfg.Config(ConfigAccount)
	if !ok {
		return "", "", azure.Environment{}, errors.New("missing account id")
	}

	key, ok = cfg.Config(ConfigKey)
	if !ok {
		return "", "", azure.Environment{}, errors.New("missing auth key")
	}

	cloud, ok := cfg.Config(ConfigCloud)
	if !ok {
		return "", "", azure.Environment{}, errors.New("missing auth key")
	}
	switch cloud {
	case "us":
		env = azure.USGovernmentCloud
	case "german":
		env = azure.GermanCloud
	case "china":
		env = azure.ChinaCloud
	case "public":
		env = azure.PublicCloud
	default:
		return "", "", azure.Environment{}, fmt.Errorf("invalid cloud %s", cloud)
	}
	return acc, key, env, nil
}

func newBlobStorageClient(account, key string, env azure.Environment) (*az.BlobStorageClient, error) {
	basicClient, err := az.NewBasicClientOnSovereignCloud(account, key, env)
	if err != nil {
		return nil, errors.New("bad credentials")
	}

	client := basicClient.GetBlobService()
	return &client, err
}
