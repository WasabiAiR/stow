package azure

import (
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

const (
	azureaccount = "piotrsplaygroundblock"
	azurekey     = "u8M0wFeizJtIKWlSC1rQsJ0w1C+QZbVeL5eavup9fTusRcUp1RN5+JMNCv6lc5usdNCAOg05cKbuOo2nZNG2Sw=="
)

func TestStow(t *testing.T) {
	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}

	test.All(t, "azure", cfg)
}
