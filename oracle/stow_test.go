package oracle

import (
	"testing"

	"github.com/graymeta/stow"
)

func TestNewOracleClient(t *testing.T) {
	config := stow.ConfigMap{
		"username": "storage-a422618:username",
		"password": "password",
		"endpoint": `https://storage-a422618.storage.oraclecloud.com/auth/v1.0`,
	}

	client, err := newOracleClient(config)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Auth Key is %s", client.AuthInfo.AuthToken)
}
