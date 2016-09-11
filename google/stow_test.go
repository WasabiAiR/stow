package google

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {

	credFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")

	if credFile == "" {
		t.Skip("skipping test because GOOGLE_CREDENTIALS_FILE not set.")
	}

	b, err := ioutil.ReadFile(credFile)
	if err != nil {
		t.Fatal(err)
	}

	config := stow.ConfigMap{
		"json":       string(b),
		"project_id": "testproject-142822",
	}
	test.All(t, "google", config)
}
