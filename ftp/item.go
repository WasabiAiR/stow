package ftp

import (
	"io"
	"net/url"

	"github.com/graymeta/stow"
	"github.com/jlaffaye/ftp"
)

type item struct {
	conn *ftp.ServerConn
}

var _ stow.Item = (*item)(nil)

func (i *item) ID() string {
	panic("not implemented")
}

func (i *item) Name() string {
	panic("not implemented")
}

func (i *item) URL() *url.URL {
	panic("not implemented")
}

func (i *item) Open() (io.ReadCloser, error) {
	panic("not implemented")
}

func (i *item) ETag() (string, error) {
	panic("not implemented")
}

func (i *item) MD5() (string, error) {
	panic("not implemented")
}
