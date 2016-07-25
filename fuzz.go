// +build gofuzz

package stow

import (
	"net/url"
)

func Fuzz(data []byte) int {
	u, err := url.Parse(string(data))
	if err != nil {
		if u != nil {
			panic("non nil url")
		}
		return 0
	}
	kind, err := KindByURL(u)
	if err != nil {
		if kind != "" {
			panic("non empty kind")
		}
		return 0
	}
	return 1
}
