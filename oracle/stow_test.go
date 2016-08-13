package oracle

import (
	"testing"

	"github.com/graymeta/stow"
)

// TestOracleClient
func TestNewOracleClient(t *testing.T) {
	config := stow.ConfigMap{
		"username": "storage-a422618:corey@graymeta.com",
		"password": "Coriffany1!",
		"endpoint": `https://storage-a422618.storage.oraclecloud.com/auth/v1.0`,
	}

	client, err := newOracleClient(config)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Auth Key is %s", client.AuthInfo.AuthToken)
}