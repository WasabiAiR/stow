package azure

import (
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
)

func TestItems(t *testing.T) {
	is := is.New(t)
	cfg := stow.ConfigMap{"account": azureaccount, "key": azurekey}
	location, err := stow.New("azure", cfg)
	is.NoErr(err)
	is.OK(location)
	container, err := location.Container("container1")
	is.NoErr(err)
	is.OK(container)
	items, _, err := container.Items(0)
	is.NoErr(err)
	is.OK(items)
	// in that container should be more than 100 items
	if len(items) < 100 {
		t.Error("Test container has less than 100 items")
	}
}
