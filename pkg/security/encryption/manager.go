package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"sync"

	"freightliner/pkg/helper/errors"
)

// Manager handles encryption and decryption operations using cloud KMS providers
type Manager struct {
	providers map[string]Provider
	config    EncryptionConfig
	mu        sync.RWMutex
}

// NewManager creates a new encryption manager
func NewManager(providers map[string]Provider, config EncryptionConfig) *Manager {
	return &Manager{
		providers: providers,
		config:    config,
	}
}

// RegisterProvider adds a provider to the manager
func (m *Manager) RegisterProvider(name string, provider Provider) {
	// Validate input before locking to fail fast
	if name == "" {
		return // Silently ignore empty provider names
	}
	if provider == nil {
		return // Silently ignore nil providers
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[name] = provider
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (Provider, error) {
	// Validate input before locking to fail fast
	if name == "" {
		return nil, errors.InvalidInputf("provider name cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, errors.NotFoundf("encryption provider not found: %s", name)
	}

	return provider, nil
}

// GetDefaultProvider returns the default provider configured for this manager
func (m *Manager) GetDefaultProvider() (Provider, error) {
	// Validate configuration before attempting to get provider
	if m.config.Provider == "" {
		return nil, errors.InvalidInputf("no default provider configured")
	}

	return m.GetProvider(m.config.Provider)
}

// EncryptData encrypts data with the default provider or the specified provider
func (m *Manager) EncryptData(ctx context.Context, data []byte, opts *EncryptOptions) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.InvalidInputf("data cannot be empty")
	}

	provider, err := m.resolveProvider(opts)
	if err != nil {
		return nil, err
	}

	// If envelope encryption is disabled, encrypt directly
	if !m.config.EnvelopeEncryption {
		return provider.Encrypt(ctx, data)
	}

	// Generate data key for envelope encryption
	plainDataKey, encryptedDataKey, err := provider.GenerateDataKey(ctx, m.config.DataKeyLength)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate data key")
	}

	// Create cipher block
	block, err := aes.NewCipher(plainDataKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to generate nonce")
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Create envelope
	envelope := Envelope{
		EncryptedKey:   encryptedDataKey,
		Ciphertext:     ciphertext,
		ProviderInfo:   provider.GetKeyInfo(),
		EnvelopeFormat: "AES-GCM",
	}

	// Serialize envelope
	return json.Marshal(envelope)
}

// DecryptData decrypts data with the appropriate provider
func (m *Manager) DecryptData(ctx context.Context, data []byte, opts *DecryptOptions) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.InvalidInputf("data cannot be empty")
	}

	// Try to parse as envelope first
	var envelope Envelope
	err := json.Unmarshal(data, &envelope)

	// If not an envelope or envelope encryption is disabled, try direct decryption
	if err != nil || !m.config.EnvelopeEncryption {
		provider, providerErr := m.resolveProvider(opts)
		if providerErr != nil {
			return nil, providerErr
		}

		return provider.Decrypt(ctx, data)
	}

	// For envelope encryption, get the provider from the envelope
	var provider Provider
	if providerName, ok := envelope.ProviderInfo["provider"]; ok {
		provider, err = m.GetProvider(providerName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get provider from envelope")
		}
	} else if opts != nil && opts.Provider != "" {
		provider, err = m.GetProvider(opts.Provider)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get specified provider")
		}
	} else {
		provider, err = m.GetDefaultProvider()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get default provider and no provider specified in envelope")
		}
	}

	// Decrypt the data key
	plainDataKey, err := provider.Decrypt(ctx, envelope.EncryptedKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt data key")
	}

	// Create cipher block
	block, err := aes.NewCipher(plainDataKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}

	// Extract nonce and ciphertext
	if len(envelope.Ciphertext) < gcm.NonceSize() {
		return nil, errors.InvalidInputf("ciphertext too short")
	}

	nonce := envelope.Ciphertext[:gcm.NonceSize()]
	ciphertext := envelope.Ciphertext[gcm.NonceSize():]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt data")
	}

	return plaintext, nil
}

// EncryptStream encrypts a stream using envelope encryption
func (m *Manager) EncryptStream(ctx context.Context, src io.Reader, dst io.Writer, opts *EncryptOptions) error {
	provider, err := m.resolveProvider(opts)
	if err != nil {
		return err
	}

	// Generate data key
	plainDataKey, encryptedDataKey, err := provider.GenerateDataKey(ctx, m.config.DataKeyLength)
	if err != nil {
		return errors.Wrap(err, "failed to generate data key")
	}

	// Create envelope header
	envelope := Envelope{
		EncryptedKey:   encryptedDataKey,
		ProviderInfo:   provider.GetKeyInfo(),
		EnvelopeFormat: "AES-GCM-STREAM",
	}

	// Write envelope header as JSON
	headerBytes, err := json.Marshal(envelope)
	if err != nil {
		return errors.Wrap(err, "failed to marshal envelope header")
	}

	// Write header length as 4 bytes
	if len(headerBytes) > math.MaxUint32 {
		return errors.InvalidInputf("header too large: %d bytes", len(headerBytes))
	}
	headerLen := uint32(len(headerBytes))
	headerLenBytes := []byte{
		byte(headerLen >> 24),
		byte(headerLen >> 16),
		byte(headerLen >> 8),
		byte(headerLen),
	}

	// Write header length and header
	if _, writeErr := dst.Write(headerLenBytes); writeErr != nil {
		return errors.Wrap(writeErr, "failed to write header length")
	}
	if _, writeErr := dst.Write(headerBytes); writeErr != nil {
		return errors.Wrap(writeErr, "failed to write header")
	}

	// Create cipher block
	block, err := aes.NewCipher(plainDataKey)
	if err != nil {
		return errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.Wrap(err, "failed to create GCM")
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return errors.Wrap(err, "failed to generate nonce")
	}

	// Write nonce
	if _, err := dst.Write(nonce); err != nil {
		return errors.Wrap(err, "failed to write nonce")
	}

	// Encrypt data in chunks
	buf := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := src.Read(buf)
		if n > 0 {
			// Encrypt chunk
			encryptedChunk := gcm.Seal(nil, nonce, buf[:n], nil)

			// Write chunk length
			if len(encryptedChunk) > math.MaxUint32 {
				return errors.InvalidInputf("encrypted chunk too large: %d bytes", len(encryptedChunk))
			}
			chunkLen := uint32(len(encryptedChunk))
			chunkLenBytes := []byte{
				byte(chunkLen >> 24),
				byte(chunkLen >> 16),
				byte(chunkLen >> 8),
				byte(chunkLen),
			}
			if _, writeErr := dst.Write(chunkLenBytes); writeErr != nil {
				return errors.Wrap(writeErr, "failed to write chunk length")
			}

			// Write encrypted chunk
			if _, writeErr := dst.Write(encryptedChunk); writeErr != nil {
				return errors.Wrap(writeErr, "failed to write encrypted chunk")
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to read from source")
		}
	}

	return nil
}

// DecryptStream decrypts a stream using envelope encryption
func (m *Manager) DecryptStream(ctx context.Context, src io.Reader, dst io.Writer, opts *DecryptOptions) error {
	// Read header length
	headerLenBytes := make([]byte, 4)
	if _, err := io.ReadFull(src, headerLenBytes); err != nil {
		return errors.Wrap(err, "failed to read header length")
	}

	headerLen := uint32(headerLenBytes[0])<<24 |
		uint32(headerLenBytes[1])<<16 |
		uint32(headerLenBytes[2])<<8 |
		uint32(headerLenBytes[3])

	// Read header
	headerBytes := make([]byte, headerLen)
	if _, err := io.ReadFull(src, headerBytes); err != nil {
		return errors.Wrap(err, "failed to read header")
	}

	// Parse envelope header
	var envelope Envelope
	if err := json.Unmarshal(headerBytes, &envelope); err != nil {
		return errors.Wrap(err, "failed to unmarshal envelope header")
	}

	// Get provider from envelope or options
	var provider Provider
	var err error
	if providerName, ok := envelope.ProviderInfo["provider"]; ok {
		provider, err = m.GetProvider(providerName)
		if err != nil {
			return errors.Wrap(err, "failed to get provider from envelope")
		}
	} else if opts != nil && opts.Provider != "" {
		provider, err = m.GetProvider(opts.Provider)
		if err != nil {
			return errors.Wrap(err, "failed to get specified provider")
		}
	} else {
		provider, err = m.GetDefaultProvider()
		if err != nil {
			return errors.Wrap(err, "failed to get default provider and no provider specified in envelope")
		}
	}

	// Decrypt data key
	plainDataKey, err := provider.Decrypt(ctx, envelope.EncryptedKey)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt data key")
	}

	// Create cipher block
	block, err := aes.NewCipher(plainDataKey)
	if err != nil {
		return errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.Wrap(err, "failed to create GCM")
	}

	// Read nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(src, nonce); err != nil {
		return errors.Wrap(err, "failed to read nonce")
	}

	// Decrypt data in chunks
	for {
		// Read chunk length
		chunkLenBytes := make([]byte, 4)
		_, err := io.ReadFull(src, chunkLenBytes)
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to read chunk length")
		}

		chunkLen := uint32(chunkLenBytes[0])<<24 |
			uint32(chunkLenBytes[1])<<16 |
			uint32(chunkLenBytes[2])<<8 |
			uint32(chunkLenBytes[3])

		// Read encrypted chunk
		encryptedChunk := make([]byte, chunkLen)
		if _, readErr := io.ReadFull(src, encryptedChunk); readErr != nil {
			return errors.Wrap(readErr, "failed to read encrypted chunk")
		}

		// Decrypt chunk
		plainChunk, err := gcm.Open(nil, nonce, encryptedChunk, nil)
		if err != nil {
			return errors.Wrap(err, "failed to decrypt chunk")
		}

		// Write decrypted chunk
		if _, err := dst.Write(plainChunk); err != nil {
			return errors.Wrap(err, "failed to write decrypted chunk")
		}
	}

	return nil
}

// resolveProvider returns the provider to use based on options and configuration
func (m *Manager) resolveProvider(opts interface{}) (Provider, error) {
	var providerName string

	switch o := opts.(type) {
	case *EncryptOptions:
		if o != nil && o.Provider != "" {
			providerName = o.Provider
		}
	case *DecryptOptions:
		if o != nil && o.Provider != "" {
			providerName = o.Provider
		}
	}

	if providerName == "" {
		providerName = m.config.Provider
	}

	if providerName == "" {
		return nil, errors.InvalidInputf("no provider specified")
	}

	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get provider")
	}

	return provider, nil
}

// EncryptBase64 encrypts data and returns a base64-encoded string
func (m *Manager) EncryptBase64(ctx context.Context, data []byte, opts *EncryptOptions) (string, error) {
	ciphertext, err := m.EncryptData(ctx, data, opts)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptBase64 decrypts a base64-encoded string
func (m *Manager) DecryptBase64(ctx context.Context, data string, opts *DecryptOptions) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base64 data")
	}
	return m.DecryptData(ctx, ciphertext, opts)
}

// Close closes all providers
func (m *Manager) Close() error {
	// No input validation needed as this is a no-parameter method

	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastErr error
	for _, p := range m.providers {
		if closer, ok := p.(io.Closer); ok {
			err := closer.Close()
			if err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}
