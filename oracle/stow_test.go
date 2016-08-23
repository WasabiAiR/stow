package swift

import (
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	cfg := stow.ConfigMap{
		"username":        "corey@graymeta.com",
		"password":        "BHBmQ585mbctcfvD",
		"identity_domain": "storage-a422618",
	}

	test.All(t, "oracle", cfg)
}
