package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/graymeta/stow"
	_ "github.com/graymeta/stow/azure"
	_ "github.com/graymeta/stow/google"
	_ "github.com/graymeta/stow/local"
	_ "github.com/graymeta/stow/oracle"
	_ "github.com/graymeta/stow/s3"
	_ "github.com/graymeta/stow/swift"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	kindFlag   = kingpin.Flag("kind", "Kind of storage to use. Required.").Short('k').Required().String()
	configFlag = kingpin.Flag("config", "Config in format \"key=value,key2=value2\"").Short('c').Required().String()

	list                = kingpin.Command("list", "Lists resources")
	listContainer       = list.Command("containers", "Lists containers")
	listContainerPrefix = listContainer.Flag("prefix", "Prefix used for listing containers and items.").String()
	listContainerCursor = listContainer.Flag("cursor", "Cursor to next page of results.").String()
	listContainerCount  = listContainer.Flag("count", "Count of items returned. Defaults to 25.").Default("25").Int()

	listItems          = list.Command("items", "Lists items in container.")
	listItemsContainer = listItems.Arg("container", "Container to get items from.").Required().String()
)

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("0.0.1").Author("Piotr Rojek")
	kingpin.CommandLine.Help = "CLI for cloud storage"
	cmd := kingpin.Parse()

	l, err := dial()
	if err != nil {
		fmt.Println("dial error", err.Error())
		os.Exit(1)
	}

	switch cmd {
	case "list containers":
		listContainersFunc(l)
	}
}

func dial() (stow.Location, error) {
	cfg, err := parseConfig()
	if err != nil {
		return nil, err
	}
	return stow.Dial(*kindFlag, cfg)
}

var ErrInvalidConfig = errors.New("invalid config")

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

func listContainersFunc(l stow.Location) {
	var prefix string

	if *listContainerPrefix != "" {
		prefix = *listContainerPrefix
	} else {
		prefix = stow.NoPrefix
	}
	cs, cursor, err := l.Containers(prefix, stow.CursorStart, *listContainerCount)
	if err != nil {
		fmt.Println("Unexpected error:", err.Error())
	}
	for i, c := range cs {
		fmt.Printf("%d: %s (ID: %s)\n", i+1, c.Name(), c.ID())
	}

	if cursor != "" {
		fmt.Println("\n\tNext cursor:", cursor)
	}
}
