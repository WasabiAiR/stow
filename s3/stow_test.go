package s3

import (
	"testing"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {

	// The only field not required for a session is 'token'.
	config := stow.ConfigMap{
		"access_key_id": "AKIAIZCA2DCPCRUODSYA",
		"secret_key":    "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",
		//		"token":         "",
		"region": "us-east-1",
	}

	test.All(t, "s3", config)
}
