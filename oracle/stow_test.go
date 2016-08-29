package swift

import (
	"net/http"
	"testing"

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
