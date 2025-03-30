package util

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestCalculateDigest(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string // Updated to include "sha256:" prefix
	}{
		{
			name:     "Empty data",
			input:    []byte{},
			expected: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Simple string",
			input:    []byte("hello world"),
			expected: "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "Binary data",
			input:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			expected: "sha256:17e88db187afd62c16e5debf3e6527cd006bc012bc90b51a810cd80c2d511f43",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			digest, err := CalculateDigest(tc.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if digest != tc.expected {
				t.Errorf("Expected digest %s, got %s", tc.expected, digest)
			}

			// Verify with native SHA256 to double-check
			h := sha256.New()
			h.Write(tc.input)
			nativeDigest := fmt.Sprintf("sha256:%x", h.Sum(nil))

			if digest != nativeDigest {
				t.Errorf("Digest doesn't match native SHA256 implementation. Got %s, expected %s", digest, nativeDigest)
			}
		})
	}
}

func TestValidateDigest(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		digest      string
		shouldMatch bool
	}{
		{
			name:        "Valid digest",
			data:        []byte("hello world"),
			digest:      "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			shouldMatch: true,
		},
		{
			name:        "Invalid digest",
			data:        []byte("hello world"),
			digest:      "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			shouldMatch: false,
		},
		{
			name:        "Invalid format digest",
			data:        []byte("hello world"),
			digest:      "not-a-valid-hex-digest",
			shouldMatch: false,
		},
		{
			name:        "Empty digest",
			data:        []byte("hello world"),
			digest:      "",
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matched, err := ValidateDigest(tc.data, tc.digest)

			if tc.shouldMatch {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if !matched {
					t.Error("Expected digest to match, but it didn't")
				}
			} else {
				if matched {
					t.Error("Expected digest to not match, but it did")
				}
			}
		})
	}
}
