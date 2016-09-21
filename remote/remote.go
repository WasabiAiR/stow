package remote

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
)

// ConfigKeys are the supported configuration items for
// remote storage (NFS, CIFS, SAMBA).
const (
	ConfigKeySource  = "source"
	ConfigKeyType    = "type"
	ConfigKeyOptions = "options"
)

// Kind is the kind of Location this package provides.
const Kind = "remote"

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {
		source, ok := config.Config(ConfigKeySource)
		if !ok {
			return nil, errors.New("missing source in config")
		}
		target := calculateTargetMountPoint(source)

		fstype, ok := config.Config(ConfigKeyType)
		if !ok {
			return nil, errors.New("missing type in config")
		}

		options, _ := config.Config(ConfigKeyOptions)

		err := mount(source, target, fstype, options)
		if err != nil {
			return nil, err
		}
		return &location{
			target: target,
			config: config,
		}, nil
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == "remote"
	}
	stow.Register(Kind, makefn, kindfn)
}

func calculateTargetMountPoint(source string) string {
	const mountpath = "stow_mountpath"
	basepath := os.Getenv(mountpath)
	if len(basepath) == 0 {
		basepath = "/lib/graymeta/mounts"
	}

	mountpoint := filepath.Join(basepath, hash(source))
	return mountpoint
}

func hash(args ...string) string {
	h := md5.New()
	for _, arg := range args {
		io.WriteString(h, arg)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
