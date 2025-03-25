package signing

import (
	"context"
	"fmt"
	"sync"
)

// Manager handles image signing and verification operations
type Manager struct {
	signers      map[string]Signer
	options      SignManagerOptions
	activeSigner string
	mu           sync.RWMutex
}

// SignManagerOptions contains options for the signing manager
type SignManagerOptions struct {
	// DefaultProvider is the name of the default signing provider to use
	DefaultProvider string

	// SignImages enables image signing when true
	SignImages bool

	// VerifyImages requires image verification when true
	VerifyImages bool

	// StrictVerification fails if verification isn't possible when true
	StrictVerification bool

	// SignatureStorePath is the path where signatures should be stored
	SignatureStorePath string

	// AllowedSigners is a list of identity patterns for acceptable signatures
	AllowedSigners []string

	// KeyPath is the path to the signing key
	KeyPath string

	// KeyID is the identifier of the signing key
	KeyID string
}

// NewManager creates a new signing manager
func NewManager(options SignManagerOptions) *Manager {
	return &Manager{
		signers:      make(map[string]Signer),
		options:      options,
		activeSigner: options.DefaultProvider,
	}
}

// RegisterSigner adds a signer to the manager
func (m *Manager) RegisterSigner(name string, signer Signer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.signers[name] = signer

	// If this is the first signer or matches the default, make it active
	if m.activeSigner == "" || m.activeSigner == name {
		m.activeSigner = name
	}
}

// GetSigner returns the signer with the given name
func (m *Manager) GetSigner(name string) (Signer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	signer, ok := m.signers[name]
	if !ok {
		return nil, fmt.Errorf("signer not found: %s", name)
	}

	return signer, nil
}

// GetActiveSigner returns the currently active signer
func (m *Manager) GetActiveSigner() (Signer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeSigner == "" {
		return nil, fmt.Errorf("no active signer configured")
	}

	signer, ok := m.signers[m.activeSigner]
	if !ok {
		return nil, fmt.Errorf("active signer not found: %s", m.activeSigner)
	}

	return signer, nil
}

// SetActiveSigner changes the active signer
func (m *Manager) SetActiveSigner(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.signers[name]
	if !ok {
		return fmt.Errorf("signer not found: %s", name)
	}

	m.activeSigner = name
	return nil
}

// IsSigningEnabled returns whether signing is enabled
func (m *Manager) IsSigningEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.options.SignImages
}

// IsVerificationEnabled returns whether verification is enabled
func (m *Manager) IsVerificationEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.options.VerifyImages
}

// SignImage signs a container image
func (m *Manager) SignImage(ctx context.Context, payload *SignaturePayload) (*Signature, error) {
	if !m.IsSigningEnabled() {
		return nil, nil // Signing is disabled, return nil without error
	}

	signer, err := m.GetActiveSigner()
	if err != nil {
		return nil, fmt.Errorf("failed to get active signer: %w", err)
	}

	signature, err := signer.Sign(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign image: %w", err)
	}

	// Store the signature if a storage path is configured
	if m.options.SignatureStorePath != "" {
		if storableSigner, ok := signer.(*CosignSigner); ok {
			sigPath := fmt.Sprintf("%s/%s-%s.sig",
				m.options.SignatureStorePath,
				sanitizeFileName(payload.Repository),
				sanitizeFileName(payload.Tag))

			if err := storableSigner.StoreSignature(ctx, signature, sigPath); err != nil {
				return nil, fmt.Errorf("failed to store signature: %w", err)
			}
		}
	}

	return signature, nil
}

// VerifyImageSignature verifies a container image signature
func (m *Manager) VerifyImageSignature(ctx context.Context, payload *SignaturePayload, signature *Signature) (bool, error) {
	if !m.IsVerificationEnabled() {
		return true, nil // Verification is disabled, return success without error
	}

	signer, err := m.GetActiveSigner()
	if err != nil {
		if m.options.StrictVerification {
			return false, fmt.Errorf("failed to get active signer: %w", err)
		}
		return false, nil
	}

	return signer.Verify(ctx, payload, signature)
}

// GetSignatureFromStorage retrieves a signature from storage
func (m *Manager) GetSignatureFromStorage(ctx context.Context, repository, tag string) (*Signature, error) {
	if m.options.SignatureStorePath == "" {
		return nil, fmt.Errorf("no signature storage path configured")
	}

	signer, err := m.GetActiveSigner()
	if err != nil {
		return nil, fmt.Errorf("failed to get active signer: %w", err)
	}

	// Only CosignSigner supports storage operations
	cosignSigner, ok := signer.(*CosignSigner)
	if !ok {
		return nil, fmt.Errorf("active signer does not support signature storage")
	}

	sigPath := fmt.Sprintf("%s/%s-%s.sig",
		m.options.SignatureStorePath,
		sanitizeFileName(repository),
		sanitizeFileName(tag))

	return cosignSigner.LoadSignature(ctx, sigPath)
}

// sanitizeFileName makes a string safe for use in a filename
func sanitizeFileName(s string) string {
	// Replace characters that are problematic in filenames
	replacements := map[byte]byte{
		'/':  '-',
		'\\': '-',
		':':  '_',
		'*':  '_',
		'?':  '_',
		'"':  '_',
		'<':  '_',
		'>':  '_',
		'|':  '_',
	}

	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if r, ok := replacements[s[i]]; ok {
			result[i] = r
		} else {
			result[i] = s[i]
		}
	}

	return string(result)
}
