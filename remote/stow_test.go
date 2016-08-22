// +build int

package remote

// Needs a NFS filesytem to run the test
// works on vagrant setup
import (
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	config := stow.ConfigMap{
		"source":  "192.168.50.1:/tmp",
		"target":  "/tmp/test/",
		"type":    "nfs",
		"options": "",
	}

	test.All(t, "remote", config)
}
