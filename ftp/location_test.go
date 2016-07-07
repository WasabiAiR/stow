package ftp

import (
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

func TestContainers(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"address": ftpaddr, "user": ftpuser, "password": ftppass}
	location, err := stow.Dial("ftp", cfg)
	is.NoErr(err)
	is.OK(location)
	containers, _, err := location.Containers("", 0)
	is.NoErr(err)
	is.OK(containers)
	is.Equal(len(containers), 7)
}

func TestContainersWithPrefix(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"address": ftpaddr, "user": ftpuser, "password": ftppass}
	location, err := stow.Dial("ftp", cfg)
	is.NoErr(err)
	is.OK(location)
	containers, _, err := location.Containers("ub", 0)
	is.NoErr(err)
	is.OK(containers)
	is.Equal(len(containers), 2)
}

func TestSingleContainer(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"address": ftpaddr, "user": ftpuser, "password": ftppass}
	location, err := stow.Dial("ftp", cfg)
	is.NoErr(err)
	is.OK(location)
	cont, err := location.Container("ubuntu")
	is.NoErr(err)
	is.OK(cont)
	is.Equal(cont.ID(), "ubuntu")
}
