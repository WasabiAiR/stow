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
