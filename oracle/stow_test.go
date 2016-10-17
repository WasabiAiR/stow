package oracle

import (
	"encoding/json"
	"testing"

	"github.com/graymeta/stow"
)

var config = stow.ConfigMap{
	"username": "storage-a422618:corey@graymeta.com",
	"password": "",
	"endpoint": `https://storage-a422618.storage.oraclecloud.com/auth/v1.0`,
}

// TestOracleClient
func TestNewOracleClient(t *testing.T) {
	client, err := newOracleClient(config)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Auth  Key is %s", client.AuthInfo.AuthToken)
}

// TestListContainers
// change response to containers
func TestListContainers(t *testing.T) {
	client, err := newOracleClient(config)
	if err != nil {
		t.Error(err)
	}

	response, err := client.ListContainers()
	if err != nil {
		t.Error(err)
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Error(err)
	}

	t.Logf("Response Bytes:\n%s", responseJSON)
}

func TestListItems(t *testing.T) {
	client, err := newOracleClient(config)
	if err != nil {
		t.Error(err)
	}

	// add prefixes and other header stuff
	_, err = client.ListItems("c1")
	if err != nil {
		t.Error(err)
	}
}
