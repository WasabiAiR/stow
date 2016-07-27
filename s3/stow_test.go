package s3

import (
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	config := stow.ConfigMap{
		"access_key_id": "AKIAIKXUQN43OZER6ZJQ",
		"secret_key":    "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",
		"region":        "us-west-1",
	}

	test.All(t, "s3", config)
}
