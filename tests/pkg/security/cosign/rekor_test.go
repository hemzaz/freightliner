//go:build cosign
// +build cosign

package cosign

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sigstore/cosign/v2/pkg/cosign/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRekorClient(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedURL string
	}{
		{
			name:        "default URL",
			url:         "",
			expectedURL: "https://rekor.sigstore.dev",
		},
		{
			name:        "custom URL",
			url:         "https://custom-rekor.example.com",
			expectedURL: "https://custom-rekor.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewRekorClient(tt.url)
			assert.NotNil(t, client)
			assert.Equal(t, tt.expectedURL, client.url)
			assert.NotNil(t, client.httpClient)
		})
	}
}

func TestRekorClient_VerifyBundle(t *testing.T) {
	ctx := context.Background()
	client := NewRekorClient("")

	tests := []struct {
		name    string
		bundle  *bundle.RekorBundle
		payload []byte
		wantErr bool
	}{
		{
			name:    "nil bundle",
			bundle:  nil,
			payload: []byte("test"),
			wantErr: true,
		},
		{
			name: "missing integrated time",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					Body:           "eyJ0ZXN0IjoidmFsdWUifQ==", // base64 encoded JSON
					IntegratedTime: 0,
				},
			},
			payload: []byte("test"),
			wantErr: true,
		},
		{
			name: "missing log index",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					Body:           "eyJ0ZXN0IjoidmFsdWUifQ==",
					IntegratedTime: 1234567890,
					LogIndex:       0,
				},
			},
			payload: []byte("test"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.VerifyBundle(ctx, tt.bundle, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRekorClient_VerifySET(t *testing.T) {
	client := NewRekorClient("")

	tests := []struct {
		name    string
		bundle  *bundle.RekorBundle
		wantErr bool
	}{
		{
			name: "valid SET",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					IntegratedTime: 1234567890,
					LogIndex:       100,
				},
			},
			wantErr: false,
		},
		{
			name: "missing integrated time",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					IntegratedTime: 0,
					LogIndex:       100,
				},
			},
			wantErr: true,
		},
		{
			name: "missing log index",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					IntegratedTime: 1234567890,
					LogIndex:       0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.verifySET(tt.bundle)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRekorClient_VerifyInclusionProof(t *testing.T) {
	client := NewRekorClient("")

	tests := []struct {
		name    string
		bundle  *bundle.RekorBundle
		wantErr bool
	}{
		{
			name: "nil inclusion proof",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					InclusionProof: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log index",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					InclusionProof: &bundle.InclusionProof{
						LogIndex: -1,
						TreeSize: 100,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "log index exceeds tree size",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					InclusionProof: &bundle.InclusionProof{
						LogIndex: 200,
						TreeSize: 100,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing hashes",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					InclusionProof: &bundle.InclusionProof{
						LogIndex: 50,
						TreeSize: 100,
						Hashes:   []string{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid inclusion proof",
			bundle: &bundle.RekorBundle{
				Payload: bundle.RekorPayload{
					InclusionProof: &bundle.InclusionProof{
						LogIndex: 50,
						TreeSize: 100,
						Hashes:   []string{"hash1", "hash2"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.verifyInclusionProof(tt.bundle)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRekorClient_GetEntry(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/v1/log/entries/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"test-uuid": {
				"body": "dGVzdCBib2R5",
				"integratedTime": 1234567890,
				"logIndex": 100,
				"logID": "test-log-id"
			}
		}`))
	}))
	defer server.Close()

	client := NewRekorClient(server.URL)
	ctx := context.Background()

	entry, err := client.GetEntry(ctx, "test-uuid")
	require.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "test-uuid", entry.UUID)
	assert.Equal(t, int64(1234567890), entry.IntegratedTime)
	assert.Equal(t, int64(100), entry.LogIndex)
	assert.Equal(t, "test-log-id", entry.LogID)
	assert.NotEmpty(t, entry.Body)
}

func TestRekorClient_GetEntry_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "entry not found"}`))
	}))
	defer server.Close()

	client := NewRekorClient(server.URL)
	ctx := context.Background()

	entry, err := client.GetEntry(ctx, "nonexistent-uuid")
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "404")
}

func TestRekorClient_SearchByDigest(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/index/retrieve" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`["uuid1", "uuid2"]`))
		} else if r.URL.Path == "/api/v1/log/entries/uuid1" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"uuid1": {
					"body": "dGVzdCBib2R5",
					"integratedTime": 1234567890,
					"logIndex": 100
				}
			}`))
		} else if r.URL.Path == "/api/v1/log/entries/uuid2" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"uuid2": {
					"body": "dGVzdCBib2R5Mg==",
					"integratedTime": 1234567891,
					"logIndex": 101
				}
			}`))
		}
	}))
	defer server.Close()

	client := NewRekorClient(server.URL)
	ctx := context.Background()

	entries, err := client.SearchByDigest(ctx, "sha256:abc123")
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestRekorClient_VerifyEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"test-uuid": {
				"body": "dGVzdCBib2R5",
				"integratedTime": 1234567890,
				"logIndex": 100
			}
		}`))
	}))
	defer server.Close()

	client := NewRekorClient(server.URL)
	ctx := context.Background()

	tests := []struct {
		name    string
		entry   *Entry
		wantErr bool
	}{
		{
			name:    "nil entry",
			entry:   nil,
			wantErr: true,
		},
		{
			name: "missing UUID",
			entry: &Entry{
				UUID: "",
			},
			wantErr: true,
		},
		{
			name: "missing integrated time",
			entry: &Entry{
				UUID:           "test-uuid",
				IntegratedTime: 0,
			},
			wantErr: true,
		},
		{
			name: "missing body",
			entry: &Entry{
				UUID:           "test-uuid",
				IntegratedTime: 1234567890,
				Body:           []byte{},
			},
			wantErr: true,
		},
		{
			name: "valid entry",
			entry: &Entry{
				UUID:           "test-uuid",
				IntegratedTime: 1234567890,
				LogIndex:       100,
				Body:           []byte("test body"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.VerifyEntry(ctx, tt.entry)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRekorClient_GetPublicKey(t *testing.T) {
	expectedKey := "-----BEGIN PUBLIC KEY-----\ntest key data\n-----END PUBLIC KEY-----"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/log/publicKey", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedKey))
	}))
	defer server.Close()

	client := NewRekorClient(server.URL)
	ctx := context.Background()

	pubKey, err := client.GetPublicKey(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, string(pubKey))
}

func TestRekorClient_GetPublicKey_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := NewRekorClient(server.URL)
	ctx := context.Background()

	pubKey, err := client.GetPublicKey(ctx)
	assert.Error(t, err)
	assert.Nil(t, pubKey)
	assert.Contains(t, err.Error(), "500")
}
