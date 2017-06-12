package google

import (
	"errors"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/graymeta/stow"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

// Kind represents the name of the location/storage type.
const Kind = "google"

const (
	// The service account json blob
	ConfigJSON      = "json"
	ConfigProjectId = "project_id"
	ConfigScopes    = "scopes"
)

func init() {

	makefn := func(config stow.Config) (stow.Location, error) {
		_, ok := config.Config(ConfigJSON)
		if !ok {
			return nil, errors.New("missing JSON configuration")
		}

		_, ok = config.Config(ConfigProjectId)
		if !ok {
			return nil, errors.New("missing Project ID")
		}

		// Create a new client
		client, err := newGoogleStorageClient(config)
		if err != nil {
			return nil, err
		}

		// Create a location with given config and client
		loc := &Location{
			config: config,
			client: client,
		}

		return loc, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn)
}

// Attempts to create a session based on the information given.
func newGoogleStorageClient(config stow.Config) (*storage.Client, error) {
	json, _ := config.Config(ConfigJSON)

	scopes := []string{storage.ScopeReadWrite}
	if s, ok := config.Config(ConfigScopes); ok && s != "" {
		scopes = strings.Split(s, ",")
	}

	var opts []option.ClientOption
	if json != "" {
		opts = append(opts, option.WithServiceAccountFile(json))
	}
	if len(scopes) > 0 {
		opts = append(opts, option.WithScopes(scopes...))
	}

	return storage.NewClient(context.Background(), opts...)
}
