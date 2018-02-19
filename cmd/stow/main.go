package main

import (
	"errors"
	"fmt"
	"io"
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
	listContainerPrefix = listContainer.Flag("prefix", "Prefix used for listing containers.").String()
	listContainerCursor = listContainer.Flag("cursor", "Cursor to next page of results.").String()
	listContainerCount  = listContainer.Flag("count", "Count of containers returned.").Default("25").Int()
	listItems           = list.Command("items", "Lists items in container.")
	listItemsContainer  = listItems.Flag("container", "Container to get items from.").Required().String()
	listItemsPrefix     = listItems.Flag("prefix", "Prefix used for listing items.").String()
	listItemsCursor     = listItems.Flag("cursor", "Cursor to next page of results.").String()
	listItemsCount      = listItems.Flag("count", "Count of items returned.").Default("25").Int()

	download          = kingpin.Command("download", "Download an item")
	downloadOutput    = download.Flag("output", "Output file to write to. If none provided, the file will be written to stdin.").String()
	downloadContainer = download.Flag("container", "Container to get items from.").Required().String()
	downloadItem      = download.Flag("item", "Item to download.").Required().String()

	upload          = kingpin.Command("upload", "Upload an item")
	uploadInput     = upload.Flag("input", "File to upload").String()
	uploadContainer = upload.Flag("container", "Container to which upload the file.").Required().String()
	uploadName      = upload.Flag("name", "Name of an item to upload").String()

	item          = kingpin.Command("item", "Get details of an item")
	itemContainer = item.Flag("container", "Container in which is the item.").Required().String()
	itemID        = item.Arg("item-id", "Item ID.").Required().String()
)

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("0.0.1").Author("GrayMeta Inc.")
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
	case "list items":
		listItemsFunc(l)
	case "download":
		downloadFunc(l)
	case "upload":
		uploadFunc(l)
	case "item":
		itemDetailsFunc(l)
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
	var (
		prefix string
		cursor string
	)

	if *listContainerPrefix != "" {
		prefix = *listContainerPrefix
	} else {
		prefix = stow.NoPrefix
	}

	if *listContainerCursor != "" {
		cursor = *listContainerCursor
	} else {
		prefix = stow.CursorStart
	}

	cs, cursor, err := l.Containers(prefix, cursor, *listContainerCount)
	if err != nil {
		fmt.Println("Unexpected error:", err.Error())
		os.Exit(1)
	}
	for i, c := range cs {
		fmt.Printf("%d: %s (ID: %s)\n", i+1, c.Name(), c.ID())
	}

	if cursor != "" {
		fmt.Println("\n\tNext cursor:", cursor)
	}
}

func listItemsFunc(l stow.Location) {
	c, err := l.Container(*listItemsContainer)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var (
		prefix string
		cursor string
	)

	if *listItemsPrefix != "" {
		prefix = *listItemsPrefix
	} else {
		prefix = stow.NoPrefix
	}

	if *listItemsCursor != "" {
		cursor = *listItemsCursor
	} else {
		prefix = stow.CursorStart
	}

	is, cursor, err := c.Items(prefix, cursor, *listItemsCount)

	for i, item := range is {
		fmt.Printf("%d: %s (ID: %s)\n", i+1, item.Name(), item.ID())
	}
	if cursor != "" {
		fmt.Println("\n\tNext cursor:", cursor)
	}
}

func downloadFunc(l stow.Location) {
	c, err := l.Container(*downloadContainer)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	i, err := c.Item(*downloadItem)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	rc, err := i.Open()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer rc.Close()

	if *downloadOutput == "" {
		io.Copy(os.Stdin, rc)
		os.Exit(0)
	}

	f, err := os.Create(*downloadOutput)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	io.Copy(f, rc)
}

func uploadFunc(l stow.Location) {
	var name string

	c, err := l.Container(*uploadContainer)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	f, err := os.Open(*uploadInput)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	info, err := f.Stat()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if *uploadName == "" {
		name = info.Name()
	} else {
		name = *uploadName
	}

	_, err = c.Put(name, f, info.Size(), nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func itemDetailsFunc(l stow.Location) {
	c, err := l.Container(*itemContainer)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	item, err := c.Item(*itemID)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	lastmod, _ := item.LastMod()
	size, _ := item.Size()
	fmt.Println("Item:\t\t", item.Name())
	fmt.Println("Item ID:\t", item.ID())
	fmt.Println("Last modified:\t", lastmod)
	fmt.Println("Size:\t\t", size)
	fmt.Println("URL:\t\t", item.URL())
}
