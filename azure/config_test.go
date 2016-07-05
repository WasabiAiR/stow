package azure

import (
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

const (
	azureaccount = "piotrsplaygroundblock"
	azurekey     = "u8M0wFeizJtIKWlSC1rQsJ0w1C+QZbVeL5eavup9fTusRcUp1RN5+JMNCv6lc5usdNCAOg05cKbuOo2nZNG2Sw=="
)

func TestConfig(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}
	location, err := stow.New("azure", cfg)
	is.NoErr(err)
	is.OK(location)
}

// func TestStow(t *testing.T) {
// 	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}

// 	test.All(t, "azure", cfg)
// }
