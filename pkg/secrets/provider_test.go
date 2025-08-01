package secrets

import (
	"testing"

	"freightliner/pkg/helper/log"
)

func TestProviderConstantsExist(t *testing.T) {
	// Verify provider constants are defined
	if AWSProvider != "aws" {
		t.Errorf("AWSProvider should equal 'aws', got %s", AWSProvider)
	}
	if GCPProvider != "gcp" {
		t.Errorf("GCPProvider should equal 'gcp', got %s", GCPProvider)
	}

	// Just a placeholder test to ensure compilation
	logger := log.NewLogger()
	if logger == nil {
		t.Fatal("Failed to create logger")
	}
}
