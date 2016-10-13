package swift

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

var cfgUnmetered = stow.ConfigMap{
	"username":               "aaron@graymeta.com",
	"password":               "HPdq85BwQ66r",
	"authorization_endpoint": "https://storage-a422618.storage.oraclecloud.com/auth/v1.0",
}

var cfgMetered = stow.ConfigMap{
	"username":               "aaron@graymeta.com",
	"password":               "Wj41xKQdYwny",
	"authorization_endpoint": "https://usoraclegm1.storage.oraclecloud.com/auth/v1.0",
}

func TestStow(t *testing.T) {
	test.All(t, "oracle", cfgMetered)
}

func TestGetItemUTCLastModified(t *testing.T) {
	tr := http.DefaultTransport
	http.DefaultTransport = &bogusLastModifiedTransport{tr}
	defer func() {
		http.DefaultTransport = tr
	}()

	test.All(t, "oracle", cfgUnmetered)
}

type bogusLastModifiedTransport struct {
	http.RoundTripper
}

func (t *bogusLastModifiedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	res, err := t.RoundTripper.RoundTrip(r)
	if err != nil {
		return res, err
	}
	res.Header.Set("Last-Modified", "Tue, 23 Aug 2016 15:12:44 UTC")
	return res, err
}

func (t *bogusLastModifiedTransport) CloseIdleConnections() {
	if tr, ok := t.RoundTripper.(interface {
		CloseIdleConnections()
	}); ok {
		tr.CloseIdleConnections()
	}
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

func TestPrepMetadataFailure(t *testing.T) {
	is := is.New(t)

	m := make(map[string]interface{})
	m["name"] = "Corey"
	m["number"] = 9

	_, err := prepMetadata(m)
	is.Err(err)
}
