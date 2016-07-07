package ftp

import (
	"net/url"

	"errors"

	"github.com/graymeta/stow"
	"github.com/jlaffaye/ftp"
)

type location struct {
	config stow.Config
	conn   *ftp.ServerConn
}

var _ stow.Location = (*location)(nil)

func (l *location) CreateContainer(name string) (stow.Container, error) {
	panic("not implemented")
}

func (l *location) Containers(prefix string, page int) ([]stow.Container, bool, error) {
	entries, err := l.conn.List("./")
	if err != nil {
		return nil, false, errors.New("Cannot list FTP containers")
	}
	var containers []stow.Container
	for _, entry := range entries {
		if entry.Type == ftp.EntryTypeFolder {
			if prefix == entry.Name[:len(prefix)] {
				toAdd := &container{
					conn: l.conn,
					name: entry.Name,
				}
				containers = append(containers, toAdd)
			}
		}
	}
	return containers, false, nil
}

func (l *location) Container(id string) (stow.Container, error) {
	entries, err := l.conn.List("./")
	if err != nil {
		return nil, errors.New("Cannot list FTP containers")
	}
	for _, entry := range entries {
		if entry.Name == id {
			return &container{
				conn: l.conn,
				name: entry.Name,
			}, nil
		}
	}
	// not found
	return nil, nil
}

func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	panic("not implemented")
}
