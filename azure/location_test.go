package azure

import (
	"testing"

	"strings"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

func TestContainers(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}
	location, err := stow.Dial("azure", cfg)
	is.NoErr(err)
	is.OK(location)
	containers, _, err := location.Containers("c", 0)
	is.NoErr(err)
	is.OK(containers)
}

func TestContainer(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}
	location, err := stow.Dial("azure", cfg)
	is.NoErr(err)
	is.OK(location)
	container, err := location.Container("container1")
	is.NoErr(err)
	is.OK(container)
}

func TestCreateContainer(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}
	location, err := stow.Dial("azure", cfg)
	is.NoErr(err)
	is.OK(location)
	newContainer, err := location.CreateContainer("testing")
	if err != nil {
		if strings.Contains(err.Error(), "ErrorCode=ContainerAlreadyExists") {
			// ignore for testing purposes
			return
		}
	}
	is.NoErr(err)
	is.OK(newContainer)
}
