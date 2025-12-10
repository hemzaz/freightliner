// Package network provides advanced networking features for registry operations
package network

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/opencontainers/go-digest"
)

// BlobMounter handles zero-copy blob mounting operations
type BlobMounter struct {
	// HTTPClient is the HTTP client for registry communication
	HTTPClient *http.Client

	// Transport is the transport with authentication
	Transport http.RoundTripper

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// EnableCrossMount enables cross-repository blob mounting
	EnableCrossMount bool
}

// MountOptions configures blob mounting behavior
type MountOptions struct {
	// SourceRepo is the source repository for cross-repository mounts
	SourceRepo string

	// VerifyMount verifies the blob exists after mounting
	VerifyMount bool

	// FallbackToCopy falls back to regular copy if mounting fails
	FallbackToCopy bool
}

// DefaultMountOptions returns default mount options
func DefaultMountOptions() *MountOptions {
	return &MountOptions{
		SourceRepo:     "",
		VerifyMount:    true,
		FallbackToCopy: true,
	}
}

// NewBlobMounter creates a new blob mounter
func NewBlobMounter(transport http.RoundTripper) *BlobMounter {
	return &BlobMounter{
		HTTPClient: &http.Client{
			Transport: transport,
		},
		Transport:        transport,
		MaxRetries:       3,
		EnableCrossMount: true,
	}
}

// Mount attempts to mount a blob from source registry to destination
// This implements the Docker Registry HTTP API V2 blob mounting feature
// which allows zero-copy transfer when both registries support it
func (m *BlobMounter) Mount(ctx context.Context, srcRef, dstRef name.Reference, blobDigest digest.Digest, opts *MountOptions) error {
	if opts == nil {
		opts = DefaultMountOptions()
	}

	// Build mount URL
	mountURL, err := m.buildMountURL(dstRef, blobDigest, srcRef)
	if err != nil {
		return fmt.Errorf("failed to build mount URL: %w", err)
	}

	// Attempt to mount the blob
	if err := m.attemptMount(ctx, mountURL); err != nil {
		if opts.FallbackToCopy {
			// Mounting failed, return error to trigger fallback
			return fmt.Errorf("mount failed, fallback to copy: %w", err)
		}
		return err
	}

	// Verify mount if requested
	if opts.VerifyMount {
		if err := m.verifyMount(ctx, dstRef, blobDigest); err != nil {
			return fmt.Errorf("mount verification failed: %w", err)
		}
	}

	return nil
}

// MountBlob attempts to mount a blob using descriptor information
func (m *BlobMounter) MountBlob(ctx context.Context, srcRef, dstRef name.Reference, desc v1.Descriptor, opts *MountOptions) error {
	// Extract digest from descriptor
	blobDigest, err := digest.Parse(desc.Digest.String())
	if err != nil {
		return fmt.Errorf("invalid blob digest: %w", err)
	}

	return m.Mount(ctx, srcRef, dstRef, blobDigest, opts)
}

// CanMount checks if blob mounting is supported by the destination registry
func (m *BlobMounter) CanMount(ctx context.Context, dstRef name.Reference) (bool, error) {
	// Check if the destination registry supports blob mounting
	// This is done by checking for the Docker-Upload-UUID header in a HEAD request

	// Build a test URL for the blobs endpoint
	registryURL := fmt.Sprintf("%s://%s", dstRef.Context().Registry.Scheme(), dstRef.Context().RegistryStr())
	repo := dstRef.Context().RepositoryStr()
	testURL := fmt.Sprintf("%s/v2/%s/blobs/uploads/", registryURL, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check mount support: %w", err)
	}
	defer resp.Body.Close()

	// If we get a 202 Accepted, the registry supports uploads and likely mounting
	if resp.StatusCode == http.StatusAccepted {
		// Check for Docker-Upload-UUID header which indicates blob upload support
		if uploadUUID := resp.Header.Get("Docker-Upload-UUID"); uploadUUID != "" {
			return true, nil
		}
	}

	return false, nil
}

// buildMountURL builds the blob mount URL according to Docker Registry HTTP API V2
func (m *BlobMounter) buildMountURL(dstRef name.Reference, blobDigest digest.Digest, srcRef name.Reference) (string, error) {
	// Build the base URL for the destination registry
	registryURL := fmt.Sprintf("%s://%s", dstRef.Context().Registry.Scheme(), dstRef.Context().RegistryStr())
	repo := dstRef.Context().RepositoryStr()

	// Build mount URL with query parameters
	// Format: POST /v2/<name>/blobs/uploads/?mount=<digest>&from=<source_repo>
	mountURL := fmt.Sprintf("%s/v2/%s/blobs/uploads/?mount=%s", registryURL, repo, blobDigest.String())

	// Add source repository for cross-repository mount
	if m.EnableCrossMount && srcRef != nil {
		sourceRepo := srcRef.Context().RepositoryStr()
		mountURL += fmt.Sprintf("&from=%s", url.QueryEscape(sourceRepo))
	}

	return mountURL, nil
}

// attemptMount attempts to mount the blob using HTTP POST
func (m *BlobMounter) attemptMount(ctx context.Context, mountURL string) error {
	// Create POST request for blob mounting
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mountURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create mount request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Length", "0")

	// Execute request with retries
	var lastErr error
	for attempt := 0; attempt <= m.MaxRetries; attempt++ {
		resp, err := m.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("mount request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		// Check response status
		switch resp.StatusCode {
		case http.StatusCreated:
			// 201 Created - blob successfully mounted
			return nil

		case http.StatusAccepted:
			// 202 Accepted - mount initiated, may need to complete upload
			// Check Location header for upload URL
			location := resp.Header.Get("Location")
			if location != "" {
				// Complete the upload with a PUT request
				return m.completeMount(ctx, location)
			}
			return nil

		case http.StatusNotFound:
			// 404 Not Found - blob doesn't exist or cross-mount not supported
			lastErr = fmt.Errorf("blob not found or cross-mount not supported")
			break

		default:
			// Read error response body
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("mount failed with status %d: %s", resp.StatusCode, string(body))
		}
	}

	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("mount failed after %d attempts", m.MaxRetries)
}

// completeMount completes a blob mount by sending a PUT request
func (m *BlobMounter) completeMount(ctx context.Context, uploadURL string) error {
	// Create PUT request to complete the upload
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create completion request: %w", err)
	}

	req.Header.Set("Content-Length", "0")

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("completion request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("completion failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// verifyMount verifies that the blob exists at the destination
func (m *BlobMounter) verifyMount(ctx context.Context, dstRef name.Reference, blobDigest digest.Digest) error {
	// Build blob URL
	registryURL := fmt.Sprintf("%s://%s", dstRef.Context().Registry.Scheme(), dstRef.Context().RegistryStr())
	repo := dstRef.Context().RepositoryStr()
	blobURL := fmt.Sprintf("%s/v2/%s/blobs/%s", registryURL, repo, blobDigest.String())

	// Create HEAD request to check blob existence
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, blobURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create verification request: %w", err)
	}

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("verification request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blob verification failed with status %d", resp.StatusCode)
	}

	// Verify digest in Content-Digest header if present
	if contentDigest := resp.Header.Get("Docker-Content-Digest"); contentDigest != "" {
		if contentDigest != blobDigest.String() {
			return fmt.Errorf("digest mismatch: expected %s, got %s", blobDigest, contentDigest)
		}
	}

	return nil
}

// BulkMount mounts multiple blobs in parallel
func (m *BlobMounter) BulkMount(ctx context.Context, srcRef, dstRef name.Reference, descriptors []v1.Descriptor, opts *MountOptions) error {
	if opts == nil {
		opts = DefaultMountOptions()
	}

	// Create error channel for collecting errors
	errChan := make(chan error, len(descriptors))

	// Mount blobs in parallel
	for _, desc := range descriptors {
		desc := desc // Capture loop variable
		go func() {
			err := m.MountBlob(ctx, srcRef, dstRef, desc, opts)
			errChan <- err
		}()
	}

	// Collect errors
	var errors []error
	for range descriptors {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to mount %d blobs: %v", len(errors), errors)
	}

	return nil
}

// GetMountableBlobs returns a list of blobs that can be mounted
func (m *BlobMounter) GetMountableBlobs(ctx context.Context, srcRef, dstRef name.Reference, descriptors []v1.Descriptor) ([]v1.Descriptor, error) {
	// Check if destination supports mounting
	supportsMount, err := m.CanMount(ctx, dstRef)
	if err != nil {
		return nil, fmt.Errorf("failed to check mount support: %w", err)
	}

	if !supportsMount {
		return nil, nil // Return empty list if mounting not supported
	}

	// All blobs are potentially mountable if the registry supports it
	// In a more sophisticated implementation, we could check if each blob
	// already exists in the destination registry
	return descriptors, nil
}

// EstimateSavings estimates the bandwidth savings from blob mounting
func (m *BlobMounter) EstimateSavings(descriptors []v1.Descriptor) int64 {
	var totalSize int64
	for _, desc := range descriptors {
		totalSize += desc.Size
	}
	return totalSize
}

// isTransientError checks if an error is transient and should be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common transient error patterns
	errMsg := strings.ToLower(err.Error())
	transientPatterns := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"temporary failure",
		"too many requests",
		"503",
		"502",
		"429",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	// Check for specific error types
	if _, ok := err.(*transport.Error); ok {
		return true
	}

	return false
}
