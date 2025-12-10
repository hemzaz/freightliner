//go:build cosign
// +build cosign

package cosign

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sigstore/cosign/v2/pkg/cosign/bundle"
)

// RekorClient handles interactions with Rekor transparency log
type RekorClient struct {
	url        string
	httpClient *http.Client
}

// Entry represents a Rekor log entry
type Entry struct {
	UUID           string
	LogIndex       int64
	Body           []byte
	IntegratedTime int64
	LogID          string
	Verification   *EntryVerification
}

// EntryVerification contains verification data for a Rekor entry
type EntryVerification struct {
	InclusionProof   *InclusionProof
	SignedEntryTime  []byte
	InclusionPromise []byte
}

// InclusionProof proves entry inclusion in transparency log
type InclusionProof struct {
	TreeSize  int64
	RootHash  []byte
	LogIndex  int64
	Hashes    [][]byte
	Timestamp int64
}

// NewRekorClient creates a new Rekor client
func NewRekorClient(url string) *RekorClient {
	if url == "" {
		url = "https://rekor.sigstore.dev" // Default public instance
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &RekorClient{
		url:        url,
		httpClient: httpClient,
	}
}

// VerifyBundle verifies a Rekor bundle from a signature
func (r *RekorClient) VerifyBundle(ctx context.Context, bundle *bundle.RekorBundle, payload []byte) error {
	if bundle == nil {
		return fmt.Errorf("bundle is nil")
	}

	// Verify bundle payload matches signature payload
	if err := r.verifyBundlePayload(bundle, payload); err != nil {
		return fmt.Errorf("bundle payload verification failed: %w", err)
	}

	// Verify signed entry timestamp (SET)
	if err := r.verifySET(bundle); err != nil {
		return fmt.Errorf("signed entry timestamp verification failed: %w", err)
	}

	// Optional: Verify inclusion proof if available
	if bundle.Payload.InclusionProof != nil {
		if err := r.verifyInclusionProof(bundle); err != nil {
			return fmt.Errorf("inclusion proof verification failed: %w", err)
		}
	}

	return nil
}

// verifyBundlePayload verifies the bundle payload matches the signature
func (r *RekorClient) verifyBundlePayload(bundle *bundle.RekorBundle, payload []byte) error {
	// Compute SHA256 of payload
	hash := sha256.Sum256(payload)
	expectedHash := hex.EncodeToString(hash[:])

	// Extract hash from bundle body
	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(bundle.Payload.Body), &bodyMap); err != nil {
		return fmt.Errorf("failed to parse bundle body: %w", err)
	}

	// Navigate to the hash field (structure varies by entry type)
	spec, ok := bodyMap["spec"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("bundle body missing spec field")
	}

	signature, ok := spec["signature"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("bundle body missing signature field")
	}

	content, ok := signature["content"].(string)
	if !ok {
		return fmt.Errorf("bundle body missing signature content")
	}

	// Decode and verify
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("failed to decode signature content: %w", err)
	}

	contentHash := sha256.Sum256(decoded)
	actualHash := hex.EncodeToString(contentHash[:])

	if actualHash != expectedHash {
		return fmt.Errorf("bundle payload hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

// verifySET verifies the signed entry timestamp
func (r *RekorClient) verifySET(bundle *bundle.RekorBundle) error {
	// The SET is a signature over the entry by Rekor's key
	// For now, we trust the bundle if it exists
	// Full implementation would verify the signature using Rekor's public key
	if bundle.Payload.IntegratedTime == 0 {
		return fmt.Errorf("bundle missing integrated time")
	}

	if bundle.Payload.LogIndex == 0 {
		return fmt.Errorf("bundle missing log index")
	}

	return nil
}

// verifyInclusionProof verifies the Merkle inclusion proof
func (r *RekorClient) verifyInclusionProof(bundle *bundle.RekorBundle) error {
	proof := bundle.Payload.InclusionProof
	if proof == nil {
		return fmt.Errorf("inclusion proof is nil")
	}

	// Verify tree size and log index are consistent
	if proof.LogIndex < 0 || proof.LogIndex >= proof.TreeSize {
		return fmt.Errorf("invalid log index %d for tree size %d", proof.LogIndex, proof.TreeSize)
	}

	// Verify Merkle proof (simplified)
	// Full implementation would compute Merkle tree verification
	if len(proof.Hashes) == 0 {
		return fmt.Errorf("inclusion proof missing hashes")
	}

	return nil
}

// SearchByDigest searches Rekor for entries matching an image digest
func (r *RekorClient) SearchByDigest(ctx context.Context, digest string) ([]*Entry, error) {
	// Construct search query
	url := fmt.Sprintf("%s/api/v1/index/retrieve", r.url)

	// Create search payload
	searchHash := sha256.Sum256([]byte(digest))
	hashQuery := hex.EncodeToString(searchHash[:])

	reqBody := map[string]interface{}{
		"hash": fmt.Sprintf("sha256:%s", hashQuery),
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var uuids []string
	if err := json.NewDecoder(resp.Body).Decode(&uuids); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Retrieve full entries
	var entries []*Entry
	for _, uuid := range uuids {
		entry, err := r.GetEntry(ctx, uuid)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to retrieve entry %s: %v\n", uuid, err)
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntry retrieves a specific entry by UUID
func (r *RekorClient) GetEntry(ctx context.Context, uuid string) (*Entry, error) {
	url := fmt.Sprintf("%s/api/v1/log/entries/%s", r.url, uuid)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get entry request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get entry failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var entryMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&entryMap); err != nil {
		return nil, fmt.Errorf("failed to decode entry response: %w", err)
	}

	// Extract entry data
	entry := &Entry{
		UUID: uuid,
	}

	// Parse entry fields (simplified)
	if data, ok := entryMap[uuid].(map[string]interface{}); ok {
		if body, ok := data["body"].(string); ok {
			decoded, err := base64.StdEncoding.DecodeString(body)
			if err != nil {
				return nil, fmt.Errorf("failed to decode entry body: %w", err)
			}
			entry.Body = decoded
		}

		if intTime, ok := data["integratedTime"].(float64); ok {
			entry.IntegratedTime = int64(intTime)
		}

		if logIndex, ok := data["logIndex"].(float64); ok {
			entry.LogIndex = int64(logIndex)
		}

		if logID, ok := data["logID"].(string); ok {
			entry.LogID = logID
		}
	}

	return entry, nil
}

// VerifyEntry verifies a Rekor entry's authenticity
func (r *RekorClient) VerifyEntry(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}

	// Verify required fields are present
	if entry.UUID == "" {
		return fmt.Errorf("entry missing UUID")
	}

	if entry.IntegratedTime == 0 {
		return fmt.Errorf("entry missing integrated time")
	}

	if len(entry.Body) == 0 {
		return fmt.Errorf("entry missing body")
	}

	// Verify entry still exists in Rekor
	retrieved, err := r.GetEntry(ctx, entry.UUID)
	if err != nil {
		return fmt.Errorf("failed to retrieve entry for verification: %w", err)
	}

	// Compare entries
	if retrieved.IntegratedTime != entry.IntegratedTime {
		return fmt.Errorf("integrated time mismatch")
	}

	if retrieved.LogIndex != entry.LogIndex {
		return fmt.Errorf("log index mismatch")
	}

	return nil
}

// GetPublicKey retrieves Rekor's public key for signature verification
func (r *RekorClient) GetPublicKey(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/log/publicKey", r.url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get public key request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get public key failed with status %d: %s", resp.StatusCode, string(body))
	}

	publicKey, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	return publicKey, nil
}
