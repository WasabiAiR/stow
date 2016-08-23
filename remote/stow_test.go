// +build int

package remote

// Needs a NFS filesytem to run the test
// works on vagrant setup
import (
	"os"
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	os.Setenv("stow_mountpath", "/tmp/test")
	config := stow.ConfigMap{
		"source":  "192.168.50.1:/tmp",
		"type":    "nfs",
		"options": "",
	}

	test.All(t, "remote", config)
}
