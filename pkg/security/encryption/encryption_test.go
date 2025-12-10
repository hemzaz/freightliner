package encryption

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	name            string
	encryptFunc     func(ctx context.Context, plaintext []byte) ([]byte, error)
	decryptFunc     func(ctx context.Context, ciphertext []byte) ([]byte, error)
	generateKeyFunc func(ctx context.Context, keyLength int) ([]byte, []byte, error)
	keyInfo         map[string]string
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	if m.encryptFunc != nil {
		return m.encryptFunc(ctx, plaintext)
	}
	// Simple XOR encryption for testing
	encrypted := make([]byte, len(plaintext))
	for i := range plaintext {
		encrypted[i] = plaintext[i] ^ 0xAA
	}
	return encrypted, nil
}

func (m *MockProvider) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if m.decryptFunc != nil {
		return m.decryptFunc(ctx, ciphertext)
	}
	// Simple XOR decryption for testing
	decrypted := make([]byte, len(ciphertext))
	for i := range ciphertext {
		decrypted[i] = ciphertext[i] ^ 0xAA
	}
	return decrypted, nil
}

func (m *MockProvider) GenerateDataKey(ctx context.Context, keyLength int) ([]byte, []byte, error) {
	if m.generateKeyFunc != nil {
		return m.generateKeyFunc(ctx, keyLength)
	}
	plainKey := make([]byte, keyLength)
	_, err := io.ReadFull(rand.Reader, plainKey)
	if err != nil {
		return nil, nil, err
	}

	encryptedKey, err := m.Encrypt(ctx, plainKey)
	if err != nil {
		return nil, nil, err
	}

	return plainKey, encryptedKey, nil
}

func (m *MockProvider) GetKeyInfo() map[string]string {
	if m.keyInfo != nil {
		return m.keyInfo
	}
	return map[string]string{
		"provider": m.name,
		"keyID":    "test-key",
	}
}

func TestNewManager(t *testing.T) {
	providers := map[string]Provider{
		"test-provider": &MockProvider{name: "test-provider"},
	}
	config := EncryptionConfig{
		Provider: "test-provider",
	}

	manager := NewManager(providers, config)

	assert.NotNil(t, manager)
	assert.Equal(t, "test-provider", manager.config.Provider)
}

func TestManager_RegisterProvider(t *testing.T) {
	manager := NewManager(map[string]Provider{}, EncryptionConfig{})

	t.Run("valid provider", func(t *testing.T) {
		provider := &MockProvider{name: "new-provider"}
		manager.RegisterProvider("new-provider", provider)

		retrievedProvider, err := manager.GetProvider("new-provider")
		require.NoError(t, err)
		assert.Equal(t, provider, retrievedProvider)
	})

	t.Run("empty name", func(t *testing.T) {
		provider := &MockProvider{name: "test"}
		manager.RegisterProvider("", provider)

		// Should not register
		_, err := manager.GetProvider("")
		assert.Error(t, err)
	})

	t.Run("nil provider", func(t *testing.T) {
		manager.RegisterProvider("nil-provider", nil)

		// Should not register
		_, err := manager.GetProvider("nil-provider")
		assert.Error(t, err)
	})
}

func TestManager_GetProvider(t *testing.T) {
	provider := &MockProvider{name: "test-provider"}
	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{})

	t.Run("existing provider", func(t *testing.T) {
		p, err := manager.GetProvider("test-provider")
		require.NoError(t, err)
		assert.Equal(t, provider, p)
	})

	t.Run("non-existent provider", func(t *testing.T) {
		_, err := manager.GetProvider("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := manager.GetProvider("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestManager_GetDefaultProvider(t *testing.T) {
	t.Run("valid default provider", func(t *testing.T) {
		provider := &MockProvider{name: "default-provider"}
		manager := NewManager(map[string]Provider{
			"default-provider": provider,
		}, EncryptionConfig{
			Provider: "default-provider",
		})

		p, err := manager.GetDefaultProvider()
		require.NoError(t, err)
		assert.Equal(t, provider, p)
	})

	t.Run("no default provider configured", func(t *testing.T) {
		manager := NewManager(map[string]Provider{}, EncryptionConfig{})

		_, err := manager.GetDefaultProvider()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no default provider")
	})
}

func TestManager_EncryptDecryptData_DirectEncryption(t *testing.T) {
	provider := &MockProvider{name: "test-provider"}
	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{
		Provider:           "test-provider",
		EnvelopeEncryption: false,
	})

	ctx := context.Background()
	plaintext := []byte("secret data")

	t.Run("successful encryption and decryption", func(t *testing.T) {
		encrypted, err := manager.EncryptData(ctx, plaintext, nil)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, encrypted)

		decrypted, err := manager.DecryptData(ctx, encrypted, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("empty data", func(t *testing.T) {
		_, err := manager.EncryptData(ctx, []byte{}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("with provider options", func(t *testing.T) {
		opts := &EncryptOptions{Provider: "test-provider"}
		encrypted, err := manager.EncryptData(ctx, plaintext, opts)
		require.NoError(t, err)

		decryptOpts := &DecryptOptions{Provider: "test-provider"}
		decrypted, err := manager.DecryptData(ctx, encrypted, decryptOpts)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestManager_EncryptDecryptData_EnvelopeEncryption(t *testing.T) {
	provider := &MockProvider{name: "test-provider"}
	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{
		Provider:           "test-provider",
		EnvelopeEncryption: true,
		DataKeyLength:      32, // 256 bits
	})

	ctx := context.Background()
	plaintext := []byte("secret data with envelope encryption")

	t.Run("successful envelope encryption and decryption", func(t *testing.T) {
		encrypted, err := manager.EncryptData(ctx, plaintext, nil)
		require.NoError(t, err)

		// Verify it's a JSON envelope
		var envelope Envelope
		err = json.Unmarshal(encrypted, &envelope)
		require.NoError(t, err)
		assert.NotEmpty(t, envelope.EncryptedKey)
		assert.NotEmpty(t, envelope.Ciphertext)
		assert.Equal(t, "AES-GCM", envelope.EnvelopeFormat)

		decrypted, err := manager.DecryptData(ctx, encrypted, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("corrupted envelope - falls back to direct decryption", func(t *testing.T) {
		// Invalid JSON will not parse as envelope, so it attempts direct decryption
		encrypted := []byte("invalid json data")
		decrypted, err := manager.DecryptData(ctx, encrypted, nil)
		require.NoError(t, err)
		// XOR decryption will work on any data
		assert.NotNil(t, decrypted)
	})

	t.Run("short ciphertext in envelope", func(t *testing.T) {
		// Create a valid encrypted key for proper envelope parsing
		validKey := make([]byte, 32)
		encryptedKey, _ := provider.Encrypt(ctx, validKey)

		envelope := Envelope{
			EncryptedKey:   encryptedKey,
			Ciphertext:     []byte("short"), // Too short for GCM
			EnvelopeFormat: "AES-GCM",
			ProviderInfo:   map[string]string{"provider": "test-provider"},
		}
		envelopeBytes, _ := json.Marshal(envelope)

		_, err := manager.DecryptData(ctx, envelopeBytes, nil)
		assert.Error(t, err)
	})
}

func TestManager_EncryptDecryptStream(t *testing.T) {
	provider := &MockProvider{
		name: "test-provider",
		generateKeyFunc: func(ctx context.Context, keyLength int) ([]byte, []byte, error) {
			plainKey := make([]byte, keyLength)
			_, err := io.ReadFull(rand.Reader, plainKey)
			if err != nil {
				return nil, nil, err
			}
			encryptedKey := make([]byte, len(plainKey))
			for i := range plainKey {
				encryptedKey[i] = plainKey[i] ^ 0xAA
			}
			return plainKey, encryptedKey, nil
		},
		decryptFunc: func(ctx context.Context, ciphertext []byte) ([]byte, error) {
			decrypted := make([]byte, len(ciphertext))
			for i := range ciphertext {
				decrypted[i] = ciphertext[i] ^ 0xAA
			}
			return decrypted, nil
		},
	}

	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{
		Provider:      "test-provider",
		DataKeyLength: 32,
	})

	ctx := context.Background()
	plaintext := []byte("stream data to encrypt")

	t.Run("successful stream encryption and decryption", func(t *testing.T) {
		// Encrypt
		src := bytes.NewReader(plaintext)
		var encrypted bytes.Buffer
		err := manager.EncryptStream(ctx, src, &encrypted, nil)
		require.NoError(t, err)

		// Decrypt
		var decrypted bytes.Buffer
		err = manager.DecryptStream(ctx, &encrypted, &decrypted, nil)
		require.NoError(t, err)

		assert.Equal(t, plaintext, decrypted.Bytes())
	})

	t.Run("large stream", func(t *testing.T) {
		// Generate large data (200KB)
		largeData := make([]byte, 200*1024)
		_, err := io.ReadFull(rand.Reader, largeData)
		require.NoError(t, err)

		// Encrypt
		src := bytes.NewReader(largeData)
		var encrypted bytes.Buffer
		err = manager.EncryptStream(ctx, src, &encrypted, nil)
		require.NoError(t, err)

		// Decrypt
		var decrypted bytes.Buffer
		err = manager.DecryptStream(ctx, &encrypted, &decrypted, nil)
		require.NoError(t, err)

		assert.Equal(t, largeData, decrypted.Bytes())
	})
}

func TestManager_EncryptDecryptBase64(t *testing.T) {
	provider := &MockProvider{name: "test-provider"}
	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{
		Provider:           "test-provider",
		EnvelopeEncryption: false,
	})

	ctx := context.Background()
	plaintext := []byte("secret data")

	t.Run("successful base64 encryption and decryption", func(t *testing.T) {
		encrypted, err := manager.EncryptBase64(ctx, plaintext, nil)
		require.NoError(t, err)

		// Verify it's valid base64
		_, err = base64.StdEncoding.DecodeString(encrypted)
		require.NoError(t, err)

		decrypted, err := manager.DecryptBase64(ctx, encrypted, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("invalid base64", func(t *testing.T) {
		_, err := manager.DecryptBase64(ctx, "not-valid-base64!!!", nil)
		assert.Error(t, err)
	})
}

func TestManager_Close(t *testing.T) {
	closeCalled := false
	provider := &mockClosableProvider{
		MockProvider: MockProvider{name: "test-provider"},
		closeFunc: func() error {
			closeCalled = true
			return nil
		},
	}

	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{})

	err := manager.Close()
	require.NoError(t, err)
	assert.True(t, closeCalled)
}

type mockClosableProvider struct {
	MockProvider
	closeFunc func() error
}

func (m *mockClosableProvider) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestValidateKeyARN(t *testing.T) {
	tests := []struct {
		name    string
		arn     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid ARN",
			arn:     "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
			wantErr: false,
		},
		{
			name:    "valid ARN with aws-cn partition",
			arn:     "arn:aws-cn:kms:cn-north-1:123456789012:key/12345678-1234-1234-1234-123456789012",
			wantErr: false,
		},
		{
			name:    "valid ARN with aws-us-gov partition",
			arn:     "arn:aws-us-gov:kms:us-gov-west-1:123456789012:key/12345678-1234-1234-1234-123456789012",
			wantErr: false,
		},
		{
			name:    "empty ARN",
			arn:     "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "invalid format",
			arn:     "not:an:arn",
			wantErr: true,
			errMsg:  "invalid ARN format",
		},
		{
			name:    "invalid prefix",
			arn:     "invalid:aws:kms:us-east-1:123456789012:key/test",
			wantErr: true,
			errMsg:  "invalid ARN prefix",
		},
		{
			name:    "invalid partition",
			arn:     "arn:invalid:kms:us-east-1:123456789012:key/test",
			wantErr: true,
			errMsg:  "invalid ARN partition",
		},
		{
			name:    "invalid service",
			arn:     "arn:aws:invalid:us-east-1:123456789012:key/test",
			wantErr: true,
			errMsg:  "invalid ARN service",
		},
		{
			name:    "missing region",
			arn:     "arn:aws:kms::123456789012:key/test",
			wantErr: true,
			errMsg:  "missing region",
		},
		{
			name:    "missing account",
			arn:     "arn:aws:kms:us-east-1::key/test",
			wantErr: true,
			errMsg:  "missing account ID",
		},
		{
			name:    "invalid resource",
			arn:     "arn:aws:kms:us-east-1:123456789012:invalid/test",
			wantErr: true,
			errMsg:  "resource must start with 'key/'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeyARN(tt.arn)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGCPKMSKeyName(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid key name",
			keyName: "projects/test-project/locations/us-central1/keyRings/test-ring/cryptoKeys/test-key",
			wantErr: false,
		},
		{
			name:    "invalid format - too short",
			keyName: "projects/test-project/locations/us-central1",
			wantErr: true,
			errMsg:  "invalid GCP KMS key name format",
		},
		{
			name:    "invalid format - wrong prefix",
			keyName: "invalid/test-project/locations/us-central1/keyRings/test-ring/cryptoKeys/test-key",
			wantErr: true,
			errMsg:  "must start with 'projects/'",
		},
		{
			name:    "invalid format - missing locations",
			keyName: "projects/test-project/invalid/us-central1/keyRings/test-ring/cryptoKeys/test-key",
			wantErr: true,
			errMsg:  "missing 'locations/'",
		},
		{
			name:    "invalid format - missing keyRings",
			keyName: "projects/test-project/locations/us-central1/invalid/test-ring/cryptoKeys/test-key",
			wantErr: true,
			errMsg:  "missing 'keyRings/'",
		},
		{
			name:    "invalid format - missing cryptoKeys",
			keyName: "projects/test-project/locations/us-central1/keyRings/test-ring/invalid/test-key",
			wantErr: true,
			errMsg:  "missing 'cryptoKeys/'",
		},
		{
			name:    "empty project ID",
			keyName: "projects//locations/us-central1/keyRings/test-ring/cryptoKeys/test-key",
			wantErr: true,
			errMsg:  "empty project ID",
		},
		{
			name:    "empty location",
			keyName: "projects/test-project/locations//keyRings/test-ring/cryptoKeys/test-key",
			wantErr: true,
			errMsg:  "empty location",
		},
		{
			name:    "empty key ring",
			keyName: "projects/test-project/locations/us-central1/keyRings//cryptoKeys/test-key",
			wantErr: true,
			errMsg:  "empty key ring",
		},
		{
			name:    "empty key name",
			keyName: "projects/test-project/locations/us-central1/keyRings/test-ring/cryptoKeys/",
			wantErr: true,
			errMsg:  "empty key name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGCPKMSKeyName(tt.keyName)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRandomBytes(t *testing.T) {
	t.Run("generates random bytes", func(t *testing.T) {
		buf := make([]byte, 32)
		n, err := getRandomBytes(buf)
		require.NoError(t, err)
		assert.Equal(t, 32, n)

		// Check that not all bytes are zero
		allZero := true
		for _, b := range buf {
			if b != 0 {
				allZero = false
				break
			}
		}
		assert.False(t, allZero, "random bytes should not all be zero")
	})
}

func TestEnvelopeEncryptionIntegration(t *testing.T) {
	// Create a more realistic provider that uses actual AES encryption
	provider := &MockProvider{
		name: "aes-provider",
		generateKeyFunc: func(ctx context.Context, keyLength int) ([]byte, []byte, error) {
			// Generate a real AES key
			plainKey := make([]byte, keyLength)
			_, err := io.ReadFull(rand.Reader, plainKey)
			if err != nil {
				return nil, nil, err
			}

			// Encrypt the key (in real scenario, this would use KMS)
			encryptedKey := make([]byte, len(plainKey))
			for i := range plainKey {
				encryptedKey[i] = plainKey[i] ^ 0xAA
			}

			return plainKey, encryptedKey, nil
		},
		decryptFunc: func(ctx context.Context, ciphertext []byte) ([]byte, error) {
			// Decrypt the key (in real scenario, this would use KMS)
			decrypted := make([]byte, len(ciphertext))
			for i := range ciphertext {
				decrypted[i] = ciphertext[i] ^ 0xAA
			}
			return decrypted, nil
		},
	}

	manager := NewManager(map[string]Provider{
		"aes-provider": provider,
	}, EncryptionConfig{
		Provider:           "aes-provider",
		EnvelopeEncryption: true,
		DataKeyLength:      32,
	})

	ctx := context.Background()
	testData := []byte("This is secret data that needs envelope encryption")

	// Encrypt
	encrypted, err := manager.EncryptData(ctx, testData, nil)
	require.NoError(t, err)

	// Verify envelope structure
	var envelope Envelope
	err = json.Unmarshal(encrypted, &envelope)
	require.NoError(t, err)
	assert.NotEmpty(t, envelope.EncryptedKey)
	assert.NotEmpty(t, envelope.Ciphertext)
	assert.Equal(t, "AES-GCM", envelope.EnvelopeFormat)
	assert.Equal(t, "aes-provider", envelope.ProviderInfo["provider"])

	// Verify GCM structure (nonce + ciphertext + tag)
	block, _ := aes.NewCipher(make([]byte, 32))
	gcm, _ := cipher.NewGCM(block)
	assert.GreaterOrEqual(t, len(envelope.Ciphertext), gcm.NonceSize())

	// Decrypt
	decrypted, err := manager.DecryptData(ctx, encrypted, nil)
	require.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}

func TestConcurrentEncryption(t *testing.T) {
	provider := &MockProvider{name: "test-provider"}
	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{
		Provider:           "test-provider",
		EnvelopeEncryption: false,
	})

	ctx := context.Background()
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer func() { done <- true }()

			plaintext := []byte("concurrent test data")
			encrypted, err := manager.EncryptData(ctx, plaintext, nil)
			assert.NoError(t, err)

			decrypted, err := manager.DecryptData(ctx, encrypted, nil)
			assert.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestManager_ResolveProvider(t *testing.T) {
	provider1 := &MockProvider{name: "provider1"}
	provider2 := &MockProvider{name: "provider2"}

	manager := NewManager(map[string]Provider{
		"provider1": provider1,
		"provider2": provider2,
	}, EncryptionConfig{
		Provider: "provider1",
	})

	ctx := context.Background()

	t.Run("uses provider from options", func(t *testing.T) {
		opts := &EncryptOptions{Provider: "provider2"}
		plaintext := []byte("test data")

		encrypted, err := manager.EncryptData(ctx, plaintext, opts)
		require.NoError(t, err)

		// Decrypt with same provider option
		decryptOpts := &DecryptOptions{Provider: "provider2"}
		decrypted, err := manager.DecryptData(ctx, encrypted, decryptOpts)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("uses default provider when no options", func(t *testing.T) {
		plaintext := []byte("test data")
		encrypted, err := manager.EncryptData(ctx, plaintext, nil)
		require.NoError(t, err)

		decrypted, err := manager.DecryptData(ctx, encrypted, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("error when no provider available", func(t *testing.T) {
		emptyManager := NewManager(map[string]Provider{}, EncryptionConfig{})
		_, err := emptyManager.EncryptData(ctx, []byte("test"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no provider specified")
	})
}

func TestManager_DecryptStream_EdgeCases(t *testing.T) {
	provider := &MockProvider{
		name: "test-provider",
		generateKeyFunc: func(ctx context.Context, keyLength int) ([]byte, []byte, error) {
			plainKey := make([]byte, keyLength)
			_, err := io.ReadFull(rand.Reader, plainKey)
			if err != nil {
				return nil, nil, err
			}
			encryptedKey := make([]byte, len(plainKey))
			for i := range plainKey {
				encryptedKey[i] = plainKey[i] ^ 0xAA
			}
			return plainKey, encryptedKey, nil
		},
		decryptFunc: func(ctx context.Context, ciphertext []byte) ([]byte, error) {
			decrypted := make([]byte, len(ciphertext))
			for i := range ciphertext {
				decrypted[i] = ciphertext[i] ^ 0xAA
			}
			return decrypted, nil
		},
	}

	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{
		Provider:      "test-provider",
		DataKeyLength: 32,
	})

	ctx := context.Background()

	t.Run("decrypt stream with invalid header length", func(t *testing.T) {
		// Create a buffer with invalid data
		var buf bytes.Buffer
		buf.Write([]byte{0, 0, 0}) // Only 3 bytes instead of 4

		var decrypted bytes.Buffer
		err := manager.DecryptStream(ctx, &buf, &decrypted, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read header length")
	})

	t.Run("decrypt stream with corrupted header", func(t *testing.T) {
		// Create a buffer with length but no header
		var buf bytes.Buffer
		buf.Write([]byte{0, 0, 0, 10}) // Header length is 10
		buf.Write([]byte("short"))     // But only 5 bytes available

		var decrypted bytes.Buffer
		err := manager.DecryptStream(ctx, &buf, &decrypted, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read header")
	})
}

func TestManager_EnvelopeEncryption_ProviderSelection(t *testing.T) {
	provider1 := &MockProvider{
		name: "provider1",
		keyInfo: map[string]string{
			"provider": "provider1",
			"keyID":    "key1",
		},
	}
	provider2 := &MockProvider{
		name: "provider2",
		keyInfo: map[string]string{
			"provider": "provider2",
			"keyID":    "key2",
		},
	}

	manager := NewManager(map[string]Provider{
		"provider1": provider1,
		"provider2": provider2,
	}, EncryptionConfig{
		Provider:           "provider1",
		EnvelopeEncryption: true,
		DataKeyLength:      32,
	})

	ctx := context.Background()
	plaintext := []byte("test data")

	t.Run("decrypt uses provider from envelope", func(t *testing.T) {
		// Encrypt with provider2
		opts := &EncryptOptions{Provider: "provider2"}
		encrypted, err := manager.EncryptData(ctx, plaintext, opts)
		require.NoError(t, err)

		// Decrypt without specifying provider (should use envelope info)
		decrypted, err := manager.DecryptData(ctx, encrypted, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("decrypt with provider option overrides envelope", func(t *testing.T) {
		// Encrypt with provider1
		encrypted, err := manager.EncryptData(ctx, plaintext, nil)
		require.NoError(t, err)

		// Decrypt explicitly with provider1
		decryptOpts := &DecryptOptions{Provider: "provider1"}
		decrypted, err := manager.DecryptData(ctx, encrypted, decryptOpts)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("decrypt falls back to default provider", func(t *testing.T) {
		// Create envelope without provider info
		validKey := make([]byte, 32)
		encryptedKey, _ := provider1.Encrypt(ctx, validKey)

		// Encrypt data with the key
		block, _ := aes.NewCipher(validKey)
		gcm, _ := cipher.NewGCM(block)
		nonce := make([]byte, gcm.NonceSize())
		io.ReadFull(rand.Reader, nonce)
		ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

		envelope := Envelope{
			EncryptedKey:   encryptedKey,
			Ciphertext:     ciphertext,
			EnvelopeFormat: "AES-GCM",
			ProviderInfo:   map[string]string{}, // Empty provider info
		}
		envelopeBytes, _ := json.Marshal(envelope)

		// Should use default provider
		decrypted, err := manager.DecryptData(ctx, envelopeBytes, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestManager_EncryptStream_EdgeCases(t *testing.T) {
	provider := &MockProvider{
		name: "test-provider",
		generateKeyFunc: func(ctx context.Context, keyLength int) ([]byte, []byte, error) {
			plainKey := make([]byte, keyLength)
			_, err := io.ReadFull(rand.Reader, plainKey)
			if err != nil {
				return nil, nil, err
			}
			encryptedKey := make([]byte, len(plainKey))
			for i := range plainKey {
				encryptedKey[i] = plainKey[i] ^ 0xAA
			}
			return plainKey, encryptedKey, nil
		},
	}

	manager := NewManager(map[string]Provider{
		"test-provider": provider,
	}, EncryptionConfig{
		Provider:      "test-provider",
		DataKeyLength: 32,
	})

	ctx := context.Background()

	t.Run("empty stream", func(t *testing.T) {
		src := bytes.NewReader([]byte{})
		var encrypted bytes.Buffer
		err := manager.EncryptStream(ctx, src, &encrypted, nil)
		require.NoError(t, err)

		// Should be able to decrypt empty stream
		var decrypted bytes.Buffer
		err = manager.DecryptStream(ctx, &encrypted, &decrypted, nil)
		require.NoError(t, err)
		assert.Empty(t, decrypted.Bytes())
	})

	t.Run("single byte stream", func(t *testing.T) {
		plaintext := []byte{0x42}
		src := bytes.NewReader(plaintext)
		var encrypted bytes.Buffer
		err := manager.EncryptStream(ctx, src, &encrypted, nil)
		require.NoError(t, err)

		var decrypted bytes.Buffer
		err = manager.DecryptStream(ctx, &encrypted, &decrypted, nil)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted.Bytes())
	})
}
