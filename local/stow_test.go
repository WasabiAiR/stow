package local_test

import (
	"os"
	"testing"

	"github.com/cheekybits/is"
	"github.com/flyteorg/stow"
	"github.com/flyteorg/stow/test"
)

func TestStow(t *testing.T) {
	is := is.New(t)

	dir, err := os.MkdirTemp("testdata", "stow")
	is.NoErr(err)
	defer os.RemoveAll(dir)
	cfg := stow.ConfigMap{"path": dir}

	test.All(t, "local", cfg)
}
