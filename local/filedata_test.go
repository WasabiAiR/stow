package local

import (
	"os"
	"testing"

	"github.com/cheekybits/is"
)

func TestFileData(t *testing.T) {
	is := is.New(t)

	info, err := os.Stat("./filedata_test.go")
	is.NoErr(err)
	is.OK(info)

	data := getFileMetadata("./filedata_test.go", info)

	is.Equal(data["is_dir"], false)
	is.Equal(data["ext"], ".go")
	is.Equal(data["path"], "filedata_test.go")
	is.Equal(data["name"], "filedata_test.go")
	is.Equal(data["mode"], "644")
	is.Equal(data["mode_d"], "420")
	is.Equal(data["perm"], "-rw-r--r--")
	is.OK(data["inode"])
	is.False(data["is_hardlink"])
	is.False(data["is_symlink"])
	is.OK(data["size"])

}

func TestDotFile(t *testing.T) {
	is := is.New(t)

	info, err := os.Stat("./testdata/.dotfile")
	is.NoErr(err)
	is.OK(info)

	data := getFileMetadata("./testdata/.dotfile", info)

	is.Equal(data["ext"], ".dotfile")
	is.Equal(data["name"], ".dotfile")
}
