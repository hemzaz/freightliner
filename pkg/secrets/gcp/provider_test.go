package gcp

import (
	"testing"

	"freightliner/pkg/helper/log"
)

func TestProviderInitialization(t *testing.T) {
	// Just a placeholder test to ensure compilation
	logger := log.NewLogger(log.InfoLevel)
	if logger == nil {
		t.Fatal("Failed to create logger")
	}
}
