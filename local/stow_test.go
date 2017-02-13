package local_test

import (
	"io/ioutil"
	"os"
	"testing"

	"runtime"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	is := is.New(t)

	dir, err := ioutil.TempDir("testdata", "stow")
	is.NoErr(err)
	defer os.RemoveAll(dir)
	cfg := stow.ConfigMap{"path": dir}

	test.All(t, "local", cfg)
}
