package util

import (
	"crypto/sha256"
	"fmt"

	"freightliner/pkg/helper/errors"
)

// CalculateDigest calculates a SHA256 digest of the given data
// Returns a digest string in the format "sha256:<hex-digest>"
func CalculateDigest(data []byte) (string, error) {
	if data == nil {
		return "", errors.InvalidInputf("data cannot be nil")
	}

	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return "", errors.Wrap(err, "failed to calculate digest")
	}

	digest := fmt.Sprintf("sha256:%x", h.Sum(nil))
	return digest, nil
}

// ValidateDigest validates that the provided digest matches the given data
func ValidateDigest(data []byte, expectedDigest string) (bool, error) {
	if data == nil {
		return false, errors.InvalidInputf("data cannot be nil")
	}

	if expectedDigest == "" {
		return false, errors.InvalidInputf("expected digest cannot be empty")
	}

	actualDigest, err := CalculateDigest(data)
	if err != nil {
		return false, errors.Wrap(err, "failed to calculate actual digest")
	}

	return actualDigest == expectedDigest, nil
}
