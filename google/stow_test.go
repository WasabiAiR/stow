package google

import (
	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
	"google.golang.org/api/storage/v1"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestStow(t *testing.T) {

	credFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	projectId := os.Getenv("GOOGLE_PROJECT_ID")

	if credFile == "" || projectId == "" {
		t.Skip("skipping test because GOOGLE_CREDENTIALS_FILE or GOOGLE_PROJECT_ID not set.")
	}

	b, err := ioutil.ReadFile(credFile)
	if err != nil {
		t.Fatal(err)
	}

	config := stow.ConfigMap{
		"json":       string(b),
		"project_id": projectId,
	}
	test.All(t, "google", config)
}

func TestParseMetadataSuccess(t *testing.T) {
	is := is.New(t)

	aclItem := &storage.ObjectAccessControl{}

	o := &storage.Object{
		Name:               "myobject",
		ContentType:        "text/html",
		ContentEncoding:    "gzip",
		ContentDisposition: "form-data",
		ContentLanguage:    "et",
		CacheControl:       "no-cache",
		Metadata: map[string]string{
			"myCustomKey": "myCustomvalue",
		},
		Acl: []*storage.ObjectAccessControl{
			aclItem,
		},
	}

	expected := map[string]interface{}{
		metaContentType:        "text/html",
		metaContentEncoding:    "gzip",
		metaContentDisposition: "form-data",
		metaContentLanguage:    "et",
		metaCacheControl:       "no-cache",
		metaACL: []*storage.ObjectAccessControl{
			aclItem,
		},
		"myCustomKey": "myCustomvalue",
	}

	itemMeta, err := parseMetadata(o)
	is.NoErr(err)
	if !reflect.DeepEqual(itemMeta, expected) {
		t.Errorf("Expected map (%+v) and returned map (%+v) are not equal.", expected, itemMeta)
	}
}

func TestPrepMetadataSuccess(t *testing.T) {
	is := is.New(t)

	o := &storage.Object{Name: "myobject"}

	m := map[string]interface{}{
		metaContentType:        "text/html",
		metaContentEncoding:    "gzip",
		metaContentDisposition: "form-data",
		metaContentLanguage:    "et",
		metaCacheControl:       "no-cache",
		"myCustomKey":          "myCustomvalue",
	}

	err := prepMetadata(o, m)
	if err != nil {
		t.Error("failed to prep metadata", err)
	}

	is.Equal(o.ContentType, "text/html")
	is.Equal(o.ContentEncoding, "gzip")
	is.Equal(o.ContentDisposition, "form-data")
	is.Equal(o.ContentLanguage, "et")
	is.Equal(o.CacheControl, "no-cache")

	is.Equal(len(o.Metadata), 1)
	customValue, ok := o.Metadata["myCustomKey"]
	is.OK(ok)
	is.Equal(customValue, "myCustomvalue")
}

func TestPrepMetadataFailureWithInvalidValues(t *testing.T) {
	is := is.New(t)

	o := &storage.Object{Name: "megaobject"}

	m := make(map[string]interface{})
	m["float"] = 8.9
	m["number"] = 9

	err := prepMetadata(o, m)
	is.Err(err)

	m = make(map[string]interface{})
	m[metaACL] = 8.9
	is.Err(err)
}
