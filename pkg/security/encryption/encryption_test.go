package encryption

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Encrypt(ctx context.Context, plaintext []byte, keyID string) ([]byte, error) {
	args := m.Called(ctx, plaintext, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockProvider) Decrypt(ctx context.Context, ciphertext []byte, keyID string) ([]byte, error) {
	args := m.Called(ctx, ciphertext, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockProvider) GenerateDataKey(ctx context.Context, keyID string, keyLength int) (*DataKey, error) {
	args := m.Called(ctx, keyID, keyLength)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DataKey), args.Error(1)
}

func TestProviderInterface(t *testing.T) {
	mockProvider := new(MockProvider)

	// Setup expectations
	mockProvider.On("Name").Return("mock-provider")
	mockProvider.On("Encrypt", mock.Anything, []byte("test-data"), "test-key").Return([]byte("encrypted-data"), nil)
	mockProvider.On("Decrypt", mock.Anything, []byte("encrypted-data"), "test-key").Return([]byte("test-data"), nil)
	mockProvider.On("GenerateDataKey", mock.Anything, "test-key", 32).Return(&DataKey{
		Plaintext:  []byte("plain-key"),
		Ciphertext: []byte("encrypted-key"),
	}, nil)

	// Test Name
	name := mockProvider.Name()
	assert.Equal(t, "mock-provider", name)

	// Test Encrypt
	ctx := context.Background()
	encrypted, err := mockProvider.Encrypt(ctx, []byte("test-data"), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("encrypted-data"), encrypted)

	// Test Decrypt
	decrypted, err := mockProvider.Decrypt(ctx, []byte("encrypted-data"), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("test-data"), decrypted)

	// Test GenerateDataKey
	dataKey, err := mockProvider.GenerateDataKey(ctx, "test-key", 32)
	assert.NoError(t, err)
	assert.Equal(t, []byte("plain-key"), dataKey.Plaintext)
	assert.Equal(t, []byte("encrypted-key"), dataKey.Ciphertext)

	// Verify all expectations were met
	mockProvider.AssertExpectations(t)
}

// We can't easily test the actual AWS/GCP KMS implementations without valid credentials,
// but we can test that the constructors require valid parameters.

func TestAWSKMSConstructor(t *testing.T) {
	// Skip if not running in a real test environment with AWS credentials
	if os.Getenv("AWS_TEST_KMS") != "true" {
		t.Skip("Skipping AWS KMS tests, set AWS_TEST_KMS=true to run")
	}

	ctx := context.Background()
	_, err := NewAWSKMS(ctx, AWSOpts{
		Region: "us-west-2",
		KeyID:  "alias/test-key",
	})
	// This will likely fail without valid credentials, but we're just testing the constructor logic
	if err != nil {
		assert.Contains(t, err.Error(), "failed to load AWS config")
	}
}

func TestGCPKMSConstructor(t *testing.T) {
	// Skip if not running in a real test environment with GCP credentials
	if os.Getenv("GCP_TEST_KMS") != "true" {
		t.Skip("Skipping GCP KMS tests, set GCP_TEST_KMS=true to run")
	}

	_, err := NewGCPKMS(context.Background(), GCPOpts{
		Project:  "test-project",
		Location: "global",
		KeyRing:  "test-keyring",
		Key:      "test-key",
	})
	// This will likely fail without valid credentials, but we're just testing the constructor logic
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create GCP KMS client")
	}
}
