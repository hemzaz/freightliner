package aws

import (
	"testing"

	"freightliner/pkg/helper/log"
)

func TestProviderInitialization(t *testing.T) {
	// Just a placeholder test to ensure compilation
	logger := log.NewLogger()
	if logger == nil {
		t.Fatal("Failed to create logger")
	}
}
