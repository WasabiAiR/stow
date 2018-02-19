package main

import (
	"testing"

	"github.com/cheekybits/is"
)

func TestParseConfig(t *testing.T) {
	is := is.New(t)

	*configFlag = "key1=value1,key2=value2"

	cfg, err := parseConfig()
	is.NoErr(err)
	is.OK(cfg)

	s, ok := cfg.Config("key1")
	is.True(ok)
	is.Equal(s, "value1")
}
