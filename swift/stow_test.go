package swift

import (
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	cfg := stow.ConfigMap{
		"username":        "sek-chai",
		"key":             "W33C6CNESJRE3H3FW67MLNNBERZRR4AGYDROX56LX2XWE4GQZC7YNS4CJEF4MPCYVYMJYPARZVUDNBOUI2XRRYWU4HTBPKQ4Q6NX",
		"tenant_name":     "graymeta_container1",
		"tenant_auth_url": "https://lax-01.identity.sohonet.com/v2.0/",
		//"tenant_id":       "b04239c7467548678b4822e9dad96030",
	}

	test.All(t, "swift", cfg)
}
