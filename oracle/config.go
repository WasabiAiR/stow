package oracle

import (
	"fmt"
	"net/http"
	"time"

	"github.com/graymeta/stow"
)

const Kind = "oracle"

const (
	// <service type>-<namespace>:<username>
	// storage-a422618:corey@graymeta.com
	storageUsername = "username"

	// Raw text
	// foobar
	storagePassword = "password"

	// Obtained from container information page. Note, must be modified.
	// https://storage-a422618.storage.oraclecloud.com/auth/v1.0
	authEndpoint = "endpoint"
)

// AuthResponse encapsulates the data returned when requesting authorization
// information for accessing storage tokens.
type AuthResponse struct {
	ContentLength string `json:"Content-Length,"`
	Server        string `json:"Server,"`
	AuthToken     string `json:"X-Auth-Token,"`
	StorageToken  string `json:"X-StorageToken,"`
	StorageURL    string `json:"X-Storage-Url,"`
	Date          string `json:"date,"`
}

// Get auth key
func newOracleClient(cfg stow.Config) (ConnectionInfo, error) {
	username, _ := cfg.Config(storageUsername)
	password, _ := cfg.Config(storagePassword)
	endpoint, _ := cfg.Config(authEndpoint)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return ConnectionInfo{}, err
	}

	request.Header = make(map[string][]string)
	request.Header.Set("X-Storage-Pass", password)
	request.Header.Set("X-Storage-User", username)

	newClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	response, err := newClient.Do(request)
	if err != nil {
		return ConnectionInfo{}, err
	} else if response.StatusCode != 200 {
		return ConnectionInfo{}, fmt.Errorf("received status code %d", response.StatusCode)
	}

	var ar AuthResponse

	ar.Server = response.Header.Get("Server")
	ar.AuthToken = response.Header.Get("X-Auth-Token")
	ar.StorageURL = response.Header.Get("X-Storage-Url")

	c := ConnectionInfo{
		client:   *newClient,
		AuthInfo: ar,
	}

	return c, nil
}
