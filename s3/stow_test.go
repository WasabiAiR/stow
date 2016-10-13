package s3

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	config := stow.ConfigMap{
		"access_key_id": "AKIAIKXUQN43OZER6ZJQ",
		"secret_key":    "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",
		"region":        "us-west-1",
	}
	test.All(t, "s3", config)
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

func TestPrepMetadataSuccess(t *testing.T) {
	is := is.New(t)

	m := make(map[string]*string)
	m["one"] = aws.String("two")
	m["3"] = aws.String("4")
	m["ninety-nine"] = aws.String("100")

	m2 := make(map[string]interface{})
	for key, value := range m {
		str := *value
		m2[key] = str
	}

	returnedMap, err := prepMetadata(m2)
	is.NoErr(err)

	if !reflect.DeepEqual(m, returnedMap) {
		t.Error("Expected and returned maps are not equal.")
	}
}

func TestPrepMetadataFailure(t *testing.T) {
	is := is.New(t)

	m := make(map[string]interface{})
	m["name"] = "Corey"
	m["number"] = 9

	_, err := prepMetadata(m)
	is.Err(err)
}
