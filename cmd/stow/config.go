package main

import (
	"strings"

	"github.com/graymeta/stow"
)

func parseConfig() (stow.ConfigMap, error) {
	if *configFlag == "" {
		return nil, ErrInvalidConfig
	}

	cfg := stow.ConfigMap{}

	configs := strings.Split(*configFlag, ",")
	for _, c := range configs {
		p := strings.Split(c, "=")
		if len(p) != 2 {
			return nil, ErrInvalidConfig
		}
		cfg[p[0]] = p[1]
	}

	return cfg, nil
}
