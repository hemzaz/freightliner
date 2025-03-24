package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
)

// Manager handles encryption operations, including selecting the right provider
// and managing envelope encryption
type Manager struct {
	providers map[string]Provider
	config    EncryptionConfig
	mu        sync.RWMutex
}

// NewManager creates a new encryption manager with the specified providers
func NewManager(providers map[string]Provider, config EncryptionConfig) *Manager {
	return &Manager{
		providers: providers,
		config:    config,
	}
}

// RegisterProvider adds a provider to the manager
func (m *Manager) RegisterProvider(name string, provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.providers == nil {
		m.providers = make(map[string]Provider)
	}
	
	m.providers[name] = provider
}

// GetProvider returns the provider with the given name
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("encryption provider not found: %s", name)
	}
	
	return provider, nil
}

// SetConfig updates the encryption configuration
func (m *Manager) SetConfig(config EncryptionConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.config = config
}

// GetConfig returns the current encryption configuration
func (m *Manager) GetConfig() EncryptionConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.config
}

// Encrypt encrypts the plaintext using the configured provider and key
func (m *Manager) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()
	
	// Get the provider
	provider, err := m.GetProvider(config.Provider)
	if err != nil {
		return nil, err
	}
	
	// If envelope encryption is enabled, use a data key
	if config.EnvelopeEncryption {
		return m.envelopeEncrypt(ctx, provider, plaintext, config.KeyID, config.DataKeyLength)
	}
	
	// Otherwise, use direct KMS encryption
	return provider.Encrypt(ctx, plaintext, config.KeyID)
}

// Decrypt decrypts the ciphertext using the configured provider and key
func (m *Manager) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()
	
	// Get the provider
	provider, err := m.GetProvider(config.Provider)
	if err != nil {
		return nil, err
	}
	
	// If envelope encryption is enabled, handle specially
	if config.EnvelopeEncryption {
		return m.envelopeDecrypt(ctx, provider, ciphertext, config.KeyID)
	}
	
	// Otherwise, use direct KMS decryption
	return provider.Decrypt(ctx, ciphertext, config.KeyID)
}

// GetKeyInfo retrieves information about the configured key
func (m *Manager) GetKeyInfo(ctx context.Context) (*KeyInfo, error) {
	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()
	
	// Get the provider
	provider, err := m.GetProvider(config.Provider)
	if err != nil {
		return nil, err
	}
	
	return provider.GetKeyInfo(ctx, config.KeyID)
}

// envelopeEncrypt encrypts data using envelope encryption:
// 1. Generate a data key using KMS
// 2. Encrypt the plaintext with the data key using AES-GCM
// 3. Prepend the encrypted data key to the ciphertext
func (m *Manager) envelopeEncrypt(ctx context.Context, provider Provider, plaintext []byte, keyID string, keyLength int) ([]byte, error) {
	// Generate a data key using KMS
	dataKey, err := provider.GenerateDataKey(ctx, keyID, keyLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate data key: %w", err)
	}
	
	// Use the plaintext data key to encrypt the data with AES-GCM
	block, err := aes.NewCipher(dataKey.Plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}
	
	// Create a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// Encrypt the plaintext
	contentCiphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	
	// Format: [Length of encrypted data key (4 bytes)][Encrypted data key][Content ciphertext (with nonce prepended)]
	encodedKeyLength := len(dataKey.Ciphertext)
	
	// Construct the final ciphertext
	result := make([]byte, 4+encodedKeyLength+len(contentCiphertext))
	
	// Store the key length as 4 bytes
	result[0] = byte(encodedKeyLength >> 24)
	result[1] = byte(encodedKeyLength >> 16)
	result[2] = byte(encodedKeyLength >> 8)
	result[3] = byte(encodedKeyLength)
	
	// Copy the encrypted data key
	copy(result[4:4+encodedKeyLength], dataKey.Ciphertext)
	
	// Copy the content ciphertext (including nonce)
	copy(result[4+encodedKeyLength:], contentCiphertext)
	
	return result, nil
}

// envelopeDecrypt decrypts data that was encrypted using envelope encryption:
// 1. Extract the encrypted data key from the ciphertext
// 2. Decrypt the data key using KMS
// 3. Use the plaintext data key to decrypt the content
func (m *Manager) envelopeDecrypt(ctx context.Context, provider Provider, ciphertext []byte, keyID string) ([]byte, error) {
	// Minimum length check
	if len(ciphertext) < 4 {
		return nil, fmt.Errorf("invalid ciphertext format: too short")
	}
	
	// Extract the encrypted data key length
	keyLength := int(ciphertext[0])<<24 | int(ciphertext[1])<<16 | int(ciphertext[2])<<8 | int(ciphertext[3])
	
	// Validate the key length
	if keyLength <= 0 || 4+keyLength >= len(ciphertext) {
		return nil, fmt.Errorf("invalid ciphertext format: invalid key length")
	}
	
	// Extract the encrypted data key
	encryptedDataKey := ciphertext[4 : 4+keyLength]
	
	// Extract the content ciphertext
	contentCiphertext := ciphertext[4+keyLength:]
	
	// Decrypt the data key using KMS
	dataKey, err := provider.Decrypt(ctx, encryptedDataKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data key: %w", err)
	}
	
	// Use the plaintext data key to decrypt the content with AES-GCM
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}
	
	// Verify we have at least enough bytes for the nonce
	if len(contentCiphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid ciphertext format: content too short")
	}
	
	// Extract the nonce
	nonce := contentCiphertext[:gcm.NonceSize()]
	
	// Extract the actual ciphertext
	content := contentCiphertext[gcm.NonceSize():]
	
	// Decrypt the content
	plaintext, err := gcm.Open(nil, nonce, content, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content: %w", err)
	}
	
	return plaintext, nil
}

// ReKeyEncryptedData re-encrypts data with a new key
func (m *Manager) ReKeyEncryptedData(ctx context.Context, ciphertext []byte, sourceKeyID, destinationKeyID string) ([]byte, error) {
	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()
	
	// Get the provider
	provider, err := m.GetProvider(config.Provider)
	if err != nil {
		return nil, err
	}
	
	// For envelope encryption, we need to decrypt and re-encrypt
	if config.EnvelopeEncryption {
		// Decrypt the data
		plaintext, err := m.envelopeDecrypt(ctx, provider, ciphertext, sourceKeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt data for re-keying: %w", err)
		}
		
		// Re-encrypt with the new key
		return m.envelopeEncrypt(ctx, provider, plaintext, destinationKeyID, config.DataKeyLength)
	}
	
	// For direct KMS encryption, we can use the provider's ReEncrypt function
	return provider.ReEncrypt(ctx, ciphertext, sourceKeyID, destinationKeyID)
}

// ListAvailableKeys lists available keys from the specified provider
func (m *Manager) ListAvailableKeys(ctx context.Context, providerName string) ([]*KeyInfo, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	// This requires type-asserting to the specific provider type
	// Here we'll handle AWS KMS and GCP KMS specifically
	switch p := provider.(type) {
	case *AWSKMS:
		// For AWS, we need to implement a method to list keys
		// This is a placeholder - in a real implementation, you'd query AWS KMS ListKeys API
		// and collect KeyInfo for each key
		return nil, fmt.Errorf("listing keys not implemented for AWS KMS")
		
	case *GCPKMS:
		// For GCP, we'll use the helper methods we've implemented
		keyRings, err := p.ListKeyRings(ctx)
		if err != nil {
			return nil, err
		}
		
		var keys []*KeyInfo
		for _, keyRing := range keyRings {
			keyIds, err := p.ListKeys(ctx, keyRing)
			if err != nil {
				continue // Skip this key ring if we can't list keys
			}
			
			for _, keyId := range keyIds {
				keyInfo, err := p.GetKeyInfo(ctx, keyId)
				if err != nil {
					continue // Skip this key if we can't get info
				}
				
				keys = append(keys, keyInfo)
			}
		}
		
		return keys, nil
		
	default:
		return nil, fmt.Errorf("listing keys not implemented for provider: %s", providerName)
	}
}

// IsCustomerManagedKeyEnabled checks if the current key is a customer-managed key
func (m *Manager) IsCustomerManagedKeyEnabled(ctx context.Context) (bool, error) {
	keyInfo, err := m.GetKeyInfo(ctx)
	if err != nil {
		return false, err
	}
	
	return keyInfo.CustomerManaged, nil
}