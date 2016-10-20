package swift

import (
	"os"
	"reflect"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	cfg := stow.ConfigMap{
		"username":        os.Getenv("SWIFTUSERNAME"),
		"key":             os.Getenv("SWIFTKEY"),
		"tenant_name":     os.Getenv("SWIFTTENANTNAME"),
		"tenant_auth_url": os.Getenv("SWIFTTENANTAUTHURL"),
		//"tenant_id":       "b04239c7467548678b4822e9dad96030",
	}
	test.All(t, "swift", cfg)
}

func TestPrepMetadataSuccess(t *testing.T) {
	is := is.New(t)

	m := make(map[string]string)
	m["one"] = "two"
	m["3"] = "4"
	m["ninety-nine"] = "100"

	m2 := make(map[string]interface{})
	for key, value := range m {
		m2[key] = value
	}

	assertionM := make(map[string]string)
	assertionM["X-Object-Meta-one"] = "two"
	assertionM["X-Object-Meta-3"] = "4"
	assertionM["X-Object-Meta-ninety-nine"] = "100"

	//returns map[string]interface
	returnedMap, err := prepMetadata(m2)
	is.NoErr(err)

	if !reflect.DeepEqual(returnedMap, assertionM) {
		t.Errorf("Expected map (%+v) and returned map (%+v) are not equal.", assertionM, returnedMap)
	}
}

func TestPrepMetadataFailureWithNonStringValues(t *testing.T) {
	is := is.New(t)

	m := make(map[string]interface{})
	m["float"] = 8.9
	m["number"] = 9

	_, err := prepMetadata(m)
	is.Err(err)
}
