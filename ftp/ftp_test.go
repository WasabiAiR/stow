package ftp

import (
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

const (
	ftpaddr = "ftp.agh.edu.pl:21"
	ftpuser = "anonymous"
	ftppass = "anonymous"
)

func TestConfig(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"address": ftpaddr, "user": ftpuser, "password": ftppass}
	location, err := stow.Dial("ftp", cfg)
	is.NoErr(err)
	is.OK(location)
}

// func TestStow(t *testing.T) {
// 	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}

// 	test.All(t, "azure", cfg)
// }
