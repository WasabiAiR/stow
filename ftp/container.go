package ftp

import (
	"io"

	"github.com/graymeta/stow"
	"github.com/jlaffaye/ftp"
)

type container struct {
	conn *ftp.ServerConn
	name string
}

func (c *container) ID() string {
	return c.name
}

func (c *container) Name() string {
	return c.name
}

func (c *container) Items(page int) ([]stow.Item, bool, error) {
	panic("not implemented")
}

func (c *container) Put(name string, r io.Reader, size int64) (stow.Item, error) {
	panic("not implemented")
}
