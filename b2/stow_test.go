package b2

import (
	"os"
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {

	account_id := os.Getenv("B2_ACCOUNT_ID")
	application_key := os.Getenv("B2_APPLICATION_KEY")

	if account_id == "" || application_key == "" {
		t.Skip("Backblaze credentials missing from environment. Skipping tests")
	}

	cfg := stow.ConfigMap{
		"account_id":      account_id,
		"application_key": application_key,
	}

	test.All(t, "b2", cfg)
}
