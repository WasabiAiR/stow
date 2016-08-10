package azure

import (
	"fmt"
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

func TestEtagCleanup(t *testing.T) {
	etagValue := "9c51403a2255f766891a1382288dece4"
	permutations := []string{
		`"%s"`,       // Enclosing quotations
		`W/\"%s\"`,   // Weak tag identifier with escapted quotes
		`W/"%s"`,     // Weak tag identifier with quotes
		`"\"%s"\"`,   // Double quotes, inner escaped
		`""%s""`,     // Double quotes,
		`"W/"%s""`,   // Double quotes with weak identifier
		`"W/\"%s\""`, // Double quotes with weak identifier, inner escaped
	}
	for index, p := range permutations {
		testStr := fmt.Sprintf(p, etagValue)
		cleanTestStr := cleanEtag(testStr)
		if etagValue != cleanTestStr {
			t.Errorf(`Failure at permutation #%d (%s), result: %s`,
				index, permutations[index], cleanTestStr)
		}
	}
}
