package s3

import (
	"log"
	"testing"

	"github.com/graymeta/stow"

	_ "github.com/graymeta/stow/test"
)

func TestGeneral(t *testing.T) {
	config := stow.ConfigMap{
		"access_key_id": "AKIAIKXUQN43OZER6ZJQ",
		"secret_key":    "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",
		"region":        "us-west-1",
	}

	loc, err := stow.Dial("s3", config)
	if err != nil {
		panic(err.Error())
	}

	conList, _, err := loc.Containers("stowtest", "")
	if err != nil {
		panic(err.Error())
	}

	log.Printf("Container List:\n")
	for _, c := range conList {
		log.Printf("%v", c.Name())
	}

	testCon, err := loc.CreateContainer("stowtest")
	if err != nil {
		panic(err.Error())
	}

	err = loc.RemoveContainer(testCon.Name())
	if err != nil {
		panic(err.Error())
	}
}
