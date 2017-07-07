package s3

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	accessKeyId := os.Getenv("S3ACCESSKEYID")
	secretKey := os.Getenv("S3SECRETKEY")
	region := os.Getenv("S3REGION")

	if accessKeyId == "" || secretKey == "" || region == "" {
		t.Skip("skipping test because missing one or more of S3ACCESSKEYID S3SECRETKEY S3REGION")
	}

	config := stow.ConfigMap{
		"access_key_id": accessKeyId,
		"secret_key":    secretKey,
		"region":        region,
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
		cleanTestStr := cleanETag(testStr)
		if etagValue != cleanTestStr {
			t.Errorf(`Failure at permutation #%d (%s), result: %s`,
				index, permutations[index], cleanTestStr)
		}
	}
}

func TestInvalidAuthtype(t *testing.T) {
	is := is.New(t)

	config := stow.ConfigMap{
		"auth_type": "foo",
	}
	_, err := stow.Dial("s3", config)
	is.Err(err)
	is.True(strings.Contains(err.Error(), "invalid auth_type"))
}
