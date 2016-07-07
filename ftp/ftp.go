package ftp

import (
	"errors"
	"net/url"

	"github.com/graymeta/stow"
	"github.com/jlaffaye/ftp"
)

const (
	Kind           = "ftp"
	ConfigAddr     = "address"
	ConfigUser     = "user"
	ConfigPassword = "password"
	// ConfigReadOnly = "readOnly"
)

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {
		addr, ok := config.Config(ConfigAddr)
		if !ok {
			return nil, errors.New("missing address config")
		}
		user, ok := config.Config(ConfigUser)
		if !ok {
			return nil, errors.New("missing user config")
		}
		password, ok := config.Config(ConfigPassword)
		if !ok {
			return nil, errors.New("missing password config")
		}
		// readOnly, ok := config.Config(ConfigReadOnly)
		// if !ok {
		// 	return nil, errors.New("missing read only config")
		// }

		// connect and validate
		conn, err := connectToFTP(addr, user, password)
		if err != nil {
			return nil, errors.New("Couldn't connect to FTP: " + err.Error())
		}
		return &location{
			config: config,
			conn:   conn,
		}, nil
	}
	kindfn := func(u *url.URL) bool {
		return u.Scheme == "ftp"
	}
	stow.Register(Kind, makefn, kindfn)
}

func connectToFTP(addr, user, pass string) (*ftp.ServerConn, error) {
	conn, err := ftp.Dial(addr)
	if err != nil {
		return nil, err
	}
	if err := conn.Login(user, pass); err != nil {
		return nil, err
	}
	// create a function to keep connection alive
	// go func() {
	// 	for {
	// 		if err := conn.NoOp(); err != nil {
	// 			fmt.Println("keep alive")
	// 			connectToFTP(addr, user, pass)
	// 			break
	// 		}
	// 		time.Sleep(time.Second * 10)
	// 	}
	// }()
	return conn, nil
}
