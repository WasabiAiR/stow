package swift

import (
	"net/http"
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

func TestGetItemUTCLastModified(t *testing.T) {
	cfg := stow.ConfigMap{
		"username":        "corey@graymeta.com",
		"password":        "BHBmQ585mbctcfvD",
		"identity_domain": "storage-a422618",
	}
	tr := http.DefaultTransport
	http.DefaultTransport = &bogusLastModifiedTransport{tr}
	defer func() {
		http.DefaultTransport = tr
	}()

	test.All(t, "oracle", cfg)
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
