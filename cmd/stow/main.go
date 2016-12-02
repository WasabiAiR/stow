package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

var (
	kindFlag      = flag.String("kind", "", "Kind of storage to use. Required.")
	configFlag    = flag.String("config", "", "Config in format \"key=value,key2=value2\"")
	containerFlag = flag.String("container", "", "Name of a container to perform operation on. Optional.")
)

func main() {
	flag.Parse()
	if len(os.Args) == 1 {
		//TODO piotrrojek: write whole usage
		fmt.Println("Usage of stow command:")
		flag.PrintDefaults()
	}

	//location, err := dial()

	arg, args := pop(os.Args[1:])
	_ = args

	switch arg {
	case "list":
	case "upload":
	case "download":
	default:
	}
}

func pop(args []string) (string, []string) {
	if len(args) == 0 {
		return "", args
	}
	return args[0], args[1:]
}

func dial() (stow.Location, error) {
	cfg, err := parseConfig()
	if err != nil {
		return nil, err
	}
	return stow.Dial(*kindFlag, cfg)
}

var (
	ErrInvalidConfig = errors.New("invalid config")
)
