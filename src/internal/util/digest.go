package util

import (
	"crypto/sha256"
	"fmt"
)

// CalculateDigest calculates a SHA256 digest of the given data
// Returns a digest string in the format "sha256:<hex-digest>"
func CalculateDigest(data []byte) (string, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return "", fmt.Errorf("failed to calculate digest: %w", err)
	}
	
	digest := fmt.Sprintf("sha256:%x", h.Sum(nil))
	return digest, nil
}

// ValidateDigest validates that the provided digest matches the given data
func ValidateDigest(data []byte, expectedDigest string) (bool, error) {
	actualDigest, err := CalculateDigest(data)
	if err != nil {
		return false, err
	}
	
	return actualDigest == expectedDigest, nil
}