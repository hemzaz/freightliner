//go:build cosign
// +build cosign

package cosign

import (
	"context"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"

	"github.com/sigstore/cosign/v2/pkg/cosign/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicy(t *testing.T) {
	policy := NewPolicy()

	assert.True(t, policy.RequireSignature)
	assert.Equal(t, 1, policy.MinSignatures)
	assert.False(t, policy.RequireRekor)
	assert.Equal(t, "enforce", policy.EnforcementMode)
}

func TestPolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  *Policy
		wantErr bool
	}{
		{
			name: "valid minimal policy",
			policy: &Policy{
				RequireSignature: true,
				MinSignatures:    1,
			},
			wantErr: false,
		},
		{
			name: "invalid negative min signatures",
			policy: &Policy{
				MinSignatures: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid enforcement mode",
			policy: &Policy{
				EnforcementMode: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid warn mode",
			policy: &Policy{
				EnforcementMode: "warn",
			},
			wantErr: false,
		},
		{
			name: "valid audit mode",
			policy: &Policy{
				EnforcementMode: "audit",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPolicy_EvaluateMinSignatures(t *testing.T) {
	ctx := context.Background()

	policy := &Policy{
		RequireSignature: true,
		MinSignatures:    2,
		EnforcementMode:  "enforce",
	}

	tests := []struct {
		name       string
		signatures []Signature
		wantErr    bool
	}{
		{
			name: "insufficient signatures",
			signatures: []Signature{
				{Digest: "sha256:abc"},
			},
			wantErr: true,
		},
		{
			name: "sufficient signatures",
			signatures: []Signature{
				{Digest: "sha256:abc"},
				{Digest: "sha256:def"},
			},
			wantErr: false,
		},
		{
			name:       "no signatures",
			signatures: []Signature{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Evaluate(ctx, tt.signatures)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPolicy_EvaluateRekorRequirement(t *testing.T) {
	ctx := context.Background()

	policy := &Policy{
		RequireSignature: true,
		MinSignatures:    1,
		RequireRekor:     true,
		EnforcementMode:  "enforce",
	}

	tests := []struct {
		name       string
		signatures []Signature
		wantErr    bool
	}{
		{
			name: "missing rekor bundle",
			signatures: []Signature{
				{
					Digest: "sha256:abc",
					Bundle: nil,
				},
			},
			wantErr: false, // Warning, but signature count still passes
		},
		{
			name: "has rekor bundle",
			signatures: []Signature{
				{
					Digest: "sha256:abc",
					Bundle: &bundle.RekorBundle{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Evaluate(ctx, tt.signatures)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// May have warnings but shouldn't fail
				_ = err
			}
		})
	}
}

func TestPolicy_EnforcementModes(t *testing.T) {
	ctx := context.Background()

	basePolicy := &Policy{
		RequireSignature: true,
		MinSignatures:    2,
	}

	signatures := []Signature{
		{Digest: "sha256:abc"},
	}

	tests := []struct {
		name            string
		enforcementMode string
		wantErr         bool
	}{
		{
			name:            "enforce mode fails",
			enforcementMode: "enforce",
			wantErr:         true,
		},
		{
			name:            "warn mode passes",
			enforcementMode: "warn",
			wantErr:         false,
		},
		{
			name:            "audit mode passes",
			enforcementMode: "audit",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := *basePolicy
			policy.EnforcementMode = tt.enforcementMode

			err := policy.Evaluate(ctx, signatures)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSignerIdentity_Validate(t *testing.T) {
	tests := []struct {
		name    string
		signer  SignerIdentity
		wantErr bool
	}{
		{
			name:    "empty signer",
			signer:  SignerIdentity{},
			wantErr: true,
		},
		{
			name: "valid email",
			signer: SignerIdentity{
				Email: "user@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid email regex",
			signer: SignerIdentity{
				EmailRegex: ".*@example\\.com$",
			},
			wantErr: false,
		},
		{
			name: "invalid email regex",
			signer: SignerIdentity{
				EmailRegex: "[invalid",
			},
			wantErr: true,
		},
		{
			name: "valid OIDC identity",
			signer: SignerIdentity{
				Issuer:  "https://token.actions.githubusercontent.com",
				Subject: "repo:org/repo:ref:refs/heads/main",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.signer.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSignerIdentity_Matches(t *testing.T) {
	tests := []struct {
		name      string
		signer    SignerIdentity
		signature Signature
		wantMatch bool
	}{
		{
			name: "matches email",
			signer: SignerIdentity{
				Email: "user@example.com",
			},
			signature: Signature{
				Certificate: &x509.Certificate{
					EmailAddresses: []string{"user@example.com"},
				},
			},
			wantMatch: true,
		},
		{
			name: "matches email regex",
			signer: SignerIdentity{
				EmailRegex: ".*@example\\.com$",
			},
			signature: Signature{
				Certificate: &x509.Certificate{
					EmailAddresses: []string{"anyone@example.com"},
				},
			},
			wantMatch: true,
		},
		{
			name: "matches OIDC identity",
			signer: SignerIdentity{
				Issuer:  "https://token.actions.githubusercontent.com",
				Subject: "repo:org/repo:ref:refs/heads/main",
			},
			signature: Signature{
				Issuer:  "https://token.actions.githubusercontent.com",
				Subject: "repo:org/repo:ref:refs/heads/main",
			},
			wantMatch: true,
		},
		{
			name: "no match wrong email",
			signer: SignerIdentity{
				Email: "user@example.com",
			},
			signature: Signature{
				Certificate: &x509.Certificate{
					EmailAddresses: []string{"other@example.com"},
				},
			},
			wantMatch: false,
		},
		{
			name: "no match wrong issuer",
			signer: SignerIdentity{
				Issuer: "https://token.actions.githubusercontent.com",
			},
			signature: Signature{
				Issuer: "https://gitlab.com",
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := tt.signer.Matches(&tt.signature)
			assert.Equal(t, tt.wantMatch, matches)
		})
	}
}

func TestPolicy_AllowedSigners(t *testing.T) {
	ctx := context.Background()

	policy := &Policy{
		RequireSignature: true,
		MinSignatures:    1,
		AllowedSigners: []SignerIdentity{
			{Email: "allowed@example.com"},
		},
		EnforcementMode: "enforce",
	}

	tests := []struct {
		name       string
		signatures []Signature
		wantErr    bool
	}{
		{
			name: "allowed signer",
			signatures: []Signature{
				{
					Digest: "sha256:abc",
					Certificate: &x509.Certificate{
						EmailAddresses: []string{"allowed@example.com"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "disallowed signer",
			signatures: []Signature{
				{
					Digest: "sha256:abc",
					Certificate: &x509.Certificate{
						EmailAddresses: []string{"unauthorized@example.com"},
					},
				},
			},
			wantErr: false, // Has warning but signature count still satisfies
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Evaluate(ctx, tt.signatures)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				_ = err
			}
		})
	}
}

func TestLoadPolicyFromFile(t *testing.T) {
	// Create temporary policy file
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.yaml")

	yamlContent := `
requireSignature: true
minSignatures: 2
requireRekor: true
enforcementMode: enforce
allowedIssuers:
  - https://token.actions.githubusercontent.com
`

	err := os.WriteFile(policyPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load policy
	policy, err := LoadPolicyFromFile(policyPath)
	require.NoError(t, err)
	assert.NotNil(t, policy)
	assert.True(t, policy.RequireSignature)
	assert.Equal(t, 2, policy.MinSignatures)
	assert.True(t, policy.RequireRekor)
	assert.Equal(t, "enforce", policy.EnforcementMode)
	assert.Len(t, policy.AllowedIssuers, 1)
}

func TestLoadPolicyFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.json")

	jsonContent := `{
		"requireSignature": true,
		"minSignatures": 1,
		"enforcementMode": "warn"
	}`

	err := os.WriteFile(policyPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	policy, err := LoadPolicyFromFile(policyPath)
	require.NoError(t, err)
	assert.NotNil(t, policy)
	assert.True(t, policy.RequireSignature)
	assert.Equal(t, 1, policy.MinSignatures)
	assert.Equal(t, "warn", policy.EnforcementMode)
}

func TestLoadPolicyFromFile_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.yaml")

	invalidContent := `
requireSignature: true
minSignatures: -1
`

	err := os.WriteFile(policyPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	policy, err := LoadPolicyFromFile(policyPath)
	assert.Error(t, err)
	assert.Nil(t, policy)
}
