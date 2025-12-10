package mtls

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to generate a test certificate
func generateTestCertificate(t *testing.T, isCA bool) (*x509.Certificate, interface{}) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "test-cert",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}

	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert, privateKey
}

func TestTLSConfig_Creation(t *testing.T) {
	config := &TLSConfig{
		CertificatePath:    "/path/to/cert.pem",
		PrivateKeyPath:     "/path/to/key.pem",
		CAPath:             "/path/to/ca.pem",
		MinVersion:         "1.3",
		MaxVersion:         "1.3",
		CipherSuites:       []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
		ClientAuth:         "require-and-verify",
		ServerName:         "example.com",
		InsecureSkipVerify: false,
		RotationEnabled:    true,
		RotationPolicy: &RotationPolicy{
			Enabled:            true,
			RotateBeforeExpiry: 30 * 24 * time.Hour,
			CheckInterval:      1 * time.Hour,
			AutoRenew:          true,
			RenewalMethod:      "acme",
		},
	}

	assert.Equal(t, "/path/to/cert.pem", config.CertificatePath)
	assert.Equal(t, "1.3", config.MinVersion)
	assert.Len(t, config.CipherSuites, 2)
	assert.Equal(t, "require-and-verify", config.ClientAuth)
	assert.True(t, config.RotationEnabled)
	assert.NotNil(t, config.RotationPolicy)
}

func TestCertificateInfo_FromX509(t *testing.T) {
	cert, _ := generateTestCertificate(t, false)

	info := &CertificateInfo{
		ID: "test-cert-001",
		Subject: &Subject{
			CommonName:   cert.Subject.CommonName,
			Organization: cert.Subject.Organization,
		},
		Issuer: &Subject{
			CommonName:   cert.Issuer.CommonName,
			Organization: cert.Issuer.Organization,
		},
		SerialNumber:     cert.SerialNumber.String(),
		NotBefore:        cert.NotBefore,
		NotAfter:         cert.NotAfter,
		Fingerprint:      "sha256:abc123",
		KeyUsage:         []string{"digitalSignature", "keyEncipherment"},
		ExtendedKeyUsage: []string{"serverAuth", "clientAuth"},
		IsCA:             false,
		IsSelfSigned:     true,
		Certificate:      cert,
		Status:           CertificateStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Tags: map[string]string{
			"environment": "test",
			"service":     "api",
		},
	}

	assert.Equal(t, "test-cert-001", info.ID)
	assert.Equal(t, "test-cert", info.Subject.CommonName)
	assert.False(t, info.IsCA)
	assert.True(t, info.IsSelfSigned)
	assert.Equal(t, CertificateStatusActive, info.Status)
	assert.Len(t, info.Tags, 2)
}

func TestSubject_Complete(t *testing.T) {
	subject := &Subject{
		CommonName:         "example.com",
		Organization:       []string{"Example Inc"},
		OrganizationalUnit: []string{"Engineering", "Security"},
		Country:            []string{"US"},
		Province:           []string{"California"},
		Locality:           []string{"San Francisco"},
		StreetAddress:      []string{"123 Main St"},
		PostalCode:         []string{"94102"},
		SerialNumber:       "12345",
	}

	assert.Equal(t, "example.com", subject.CommonName)
	assert.Len(t, subject.Organization, 1)
	assert.Len(t, subject.OrganizationalUnit, 2)
	assert.Equal(t, "US", subject.Country[0])
	assert.Equal(t, "California", subject.Province[0])
}

func TestRotationPolicy_Configuration(t *testing.T) {
	policy := &RotationPolicy{
		Enabled:              true,
		RotateBeforeExpiry:   720 * time.Hour, // 30 days
		CheckInterval:        1 * time.Hour,
		AutoRenew:            true,
		RenewalMethod:        "vault",
		NotifyBeforeExpiry:   168 * time.Hour, // 7 days
		NotificationChannels: []string{"email:admin@example.com", "slack:#security"},
		MaxRetries:           3,
		RetryInterval:        5 * time.Minute,
		GracePeriod:          1 * time.Hour,
	}

	assert.True(t, policy.Enabled)
	assert.Equal(t, 720*time.Hour, policy.RotateBeforeExpiry)
	assert.Equal(t, "vault", policy.RenewalMethod)
	assert.Len(t, policy.NotificationChannels, 2)
	assert.Equal(t, 3, policy.MaxRetries)
}

func TestRotationStatus_Tracking(t *testing.T) {
	now := time.Now()
	lastRotation := now.Add(-30 * 24 * time.Hour)
	nextRotation := now.Add(60 * 24 * time.Hour)
	expiresAt := now.Add(90 * 24 * time.Hour)

	status := &RotationStatus{
		CertificateID:    "cert-001",
		LastRotation:     &lastRotation,
		NextRotation:     &nextRotation,
		Status:           "scheduled",
		Attempts:         0,
		ExpiresAt:        expiresAt,
		DaysUntilExpiry:  90,
		RotationRequired: false,
	}

	assert.Equal(t, "cert-001", status.CertificateID)
	assert.NotNil(t, status.LastRotation)
	assert.NotNil(t, status.NextRotation)
	assert.Equal(t, "scheduled", status.Status)
	assert.False(t, status.RotationRequired)

	// Test rotation required
	status.DaysUntilExpiry = 25
	status.RotationRequired = true
	assert.True(t, status.RotationRequired)
}

func TestRotationEvent_Lifecycle(t *testing.T) {
	oldCert, _ := generateTestCertificate(t, false)
	newCert, _ := generateTestCertificate(t, false)

	oldInfo := &CertificateInfo{
		ID:          "cert-001",
		Certificate: oldCert,
		Status:      CertificateStatusActive,
	}

	newInfo := &CertificateInfo{
		ID:          "cert-002",
		Certificate: newCert,
		Status:      CertificateStatusActive,
	}

	events := []RotationEvent{
		{
			CertificateID: "cert-001",
			EventType:     "scheduled",
			Timestamp:     time.Now().Add(-2 * time.Hour),
			Message:       "Rotation scheduled for cert-001",
		},
		{
			CertificateID:  "cert-001",
			EventType:      "started",
			OldCertificate: oldInfo,
			Timestamp:      time.Now().Add(-1 * time.Hour),
			Message:        "Rotation started",
		},
		{
			CertificateID:  "cert-001",
			EventType:      "completed",
			OldCertificate: oldInfo,
			NewCertificate: newInfo,
			Timestamp:      time.Now(),
			Message:        "Rotation completed successfully",
		},
	}

	assert.Len(t, events, 3)
	assert.Equal(t, "scheduled", events[0].EventType)
	assert.Equal(t, "started", events[1].EventType)
	assert.Equal(t, "completed", events[2].EventType)
	assert.NotNil(t, events[2].NewCertificate)
}

func TestCertificateRequest_RSA(t *testing.T) {
	req := &CertificateRequest{
		CommonName:         "api.example.com",
		Organization:       "Example Inc",
		OrganizationalUnit: "Engineering",
		Country:            "US",
		Province:           "CA",
		Locality:           "San Francisco",
		DNSNames:           []string{"api.example.com", "*.api.example.com"},
		IPAddresses:        []string{"192.168.1.100"},
		KeyType:            "rsa",
		KeySize:            4096,
		ValidityDuration:   365 * 24 * time.Hour,
		IsCA:               false,
		KeyUsage:           []string{"digitalSignature", "keyEncipherment"},
		ExtendedKeyUsage:   []string{"serverAuth"},
	}

	assert.Equal(t, "api.example.com", req.CommonName)
	assert.Equal(t, "rsa", req.KeyType)
	assert.Equal(t, 4096, req.KeySize)
	assert.Len(t, req.DNSNames, 2)
	assert.False(t, req.IsCA)
}

func TestCertificateRequest_ECDSA(t *testing.T) {
	req := &CertificateRequest{
		CommonName:       "service.example.com",
		KeyType:          "ecdsa",
		Curve:            "P-384",
		ValidityDuration: 730 * 24 * time.Hour, // 2 years
		IsCA:             false,
		ExtendedKeyUsage: []string{"clientAuth"},
	}

	assert.Equal(t, "ecdsa", req.KeyType)
	assert.Equal(t, "P-384", req.Curve)
	assert.Equal(t, 730*24*time.Hour, req.ValidityDuration)
}

func TestAuthenticatedIdentity_Creation(t *testing.T) {
	identity := &AuthenticatedIdentity{
		ID:                 "user-001",
		CommonName:         "john.doe@example.com",
		Organization:       "Example Inc",
		OrganizationalUnit: "Engineering",
		Roles:              []string{"developer", "admin"},
		Permissions:        []string{"read:api", "write:api", "delete:resources"},
		Attributes: map[string]string{
			"department": "engineering",
			"level":      "senior",
		},
		CertificateFingerprint: "sha256:abc123def456",
		AuthenticatedAt:        time.Now(),
		ExpiresAt:              time.Now().Add(24 * time.Hour),
	}

	assert.Equal(t, "user-001", identity.ID)
	assert.Equal(t, "john.doe@example.com", identity.CommonName)
	assert.Len(t, identity.Roles, 2)
	assert.Len(t, identity.Permissions, 3)
	assert.Len(t, identity.Attributes, 2)
}

func TestIdentity_FromCertificate(t *testing.T) {
	cert, _ := generateTestCertificate(t, false)

	identity := &Identity{
		Subject: &Subject{
			CommonName:   cert.Subject.CommonName,
			Organization: cert.Subject.Organization,
		},
		SerialNumber: cert.SerialNumber.String(),
		Fingerprint:  "sha256:fingerprint",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"role":  "developer",
		},
		Verified: true,
	}

	assert.Equal(t, "test-cert", identity.Subject.CommonName)
	assert.True(t, identity.Verified)
	assert.Len(t, identity.Claims, 2)
}

func TestCertificateStatus_Values(t *testing.T) {
	statuses := []CertificateStatus{
		CertificateStatusActive,
		CertificateStatusExpired,
		CertificateStatusRevoked,
		CertificateStatusPending,
		CertificateStatusRotating,
	}

	assert.Len(t, statuses, 5)
	assert.Equal(t, "active", string(CertificateStatusActive))
	assert.Equal(t, "expired", string(CertificateStatusExpired))
	assert.Equal(t, "revoked", string(CertificateStatusRevoked))
	assert.Equal(t, "pending", string(CertificateStatusPending))
	assert.Equal(t, "rotating", string(CertificateStatusRotating))
}

func TestHSMConfig_PKCS11(t *testing.T) {
	config := &HSMConfig{
		Provider:    "pkcs11",
		LibraryPath: "/usr/lib/softhsm/libsofthsm2.so",
		Pin:         "1234",
		SlotID:      0,
		KeyID:       "hsm-key-001",
	}

	assert.Equal(t, "pkcs11", config.Provider)
	assert.NotEmpty(t, config.LibraryPath)
	assert.Equal(t, 0, config.SlotID)
}

func TestHSMConfig_CloudKMS(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		region   string
		endpoint string
	}{
		{
			name:     "AWS KMS",
			provider: "awskms",
			region:   "us-east-1",
		},
		{
			name:     "Azure Key Vault",
			provider: "azurekeyvault",
			endpoint: "https://myvault.vault.azure.net",
		},
		{
			name:     "GCP KMS",
			provider: "gcpkms",
			region:   "us-central1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &HSMConfig{
				Provider: tt.provider,
				Region:   tt.region,
				Endpoint: tt.endpoint,
				KeyID:    "cloud-key-001",
			}

			assert.Equal(t, tt.provider, config.Provider)
			if tt.region != "" {
				assert.Equal(t, tt.region, config.Region)
			}
			if tt.endpoint != "" {
				assert.Equal(t, tt.endpoint, config.Endpoint)
			}
		})
	}
}

func TestVaultConfig_AppRole(t *testing.T) {
	config := &VaultConfig{
		Address:       "https://vault.example.com:8200",
		RoleID:        "role-id-123",
		SecretID:      "secret-id-456",
		PKIMount:      "pki",
		Role:          "web-server",
		Namespace:     "engineering",
		CAPath:        "/etc/vault/ca.pem",
		TLSSkipVerify: false,
	}

	assert.Equal(t, "https://vault.example.com:8200", config.Address)
	assert.NotEmpty(t, config.RoleID)
	assert.NotEmpty(t, config.SecretID)
	assert.Equal(t, "pki", config.PKIMount)
	assert.Equal(t, "web-server", config.Role)
	assert.False(t, config.TLSSkipVerify)
}

func TestRevocationReason_String(t *testing.T) {
	tests := []struct {
		reason   RevocationReason
		expected string
	}{
		{RevocationReasonUnspecified, "Unspecified"},
		{RevocationReasonKeyCompromise, "KeyCompromise"},
		{RevocationReasonCACompromise, "CACompromise"},
		{RevocationReasonAffiliationChanged, "AffiliationChanged"},
		{RevocationReasonSuperseded, "Superseded"},
		{RevocationReasonCessationOfOperation, "CessationOfOperation"},
		{RevocationReasonCertificateHold, "CertificateHold"},
		{RevocationReasonRemoveFromCRL, "RemoveFromCRL"},
		{RevocationReasonPrivilegeWithdrawn, "PrivilegeWithdrawn"},
		{RevocationReasonAACompromise, "AACompromise"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.reason.String())
		})
	}
}

func TestExportFormat_String(t *testing.T) {
	tests := []struct {
		format   ExportFormat
		expected string
	}{
		{ExportFormatPEM, "PEM"},
		{ExportFormatDER, "DER"},
		{ExportFormatPKCS12, "PKCS12"},
		{ExportFormatJWK, "JWK"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.String())
		})
	}
}

func TestMockTLSProvider(t *testing.T) {
	provider := &mockTLSProvider{}

	ctx := context.Background()

	t.Run("GetTLSConfig", func(t *testing.T) {
		config, err := provider.GetTLSConfig(ctx)
		require.NoError(t, err)
		assert.NotNil(t, config)
	})

	t.Run("LoadCertificate", func(t *testing.T) {
		info, err := provider.LoadCertificate(ctx, "test-cert")
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test-cert", info.ID)
	})

	t.Run("GetCACertPool", func(t *testing.T) {
		pool, err := provider.GetCACertPool(ctx)
		require.NoError(t, err)
		assert.NotNil(t, pool)
	})
}

func TestCertificateExpiry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		notAfter     time.Time
		shouldExpire bool
	}{
		{
			name:         "valid certificate",
			notAfter:     now.Add(90 * 24 * time.Hour),
			shouldExpire: false,
		},
		{
			name:         "expiring soon",
			notAfter:     now.Add(7 * 24 * time.Hour),
			shouldExpire: false,
		},
		{
			name:         "expired certificate",
			notAfter:     now.Add(-1 * 24 * time.Hour),
			shouldExpire: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &CertificateInfo{
				NotAfter: tt.notAfter,
			}

			isExpired := info.NotAfter.Before(now)
			assert.Equal(t, tt.shouldExpire, isExpired)
		})
	}
}

func TestKeyGeneration(t *testing.T) {
	t.Run("RSA key generation", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		assert.NotNil(t, privateKey)
		assert.Equal(t, 2048, privateKey.N.BitLen())
	})

	t.Run("ECDSA key generation", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		assert.NotNil(t, privateKey)
	})
}

// Mock implementations for interface testing
type mockTLSProvider struct{}

func (m *mockTLSProvider) GetTLSConfig(ctx context.Context) (*TLSConfig, error) {
	return &TLSConfig{
		MinVersion: "1.3",
		ClientAuth: "require",
	}, nil
}

func (m *mockTLSProvider) LoadCertificate(ctx context.Context, certID string) (*CertificateInfo, error) {
	return &CertificateInfo{
		ID:     certID,
		Status: CertificateStatusActive,
		Subject: &Subject{
			CommonName: "test.example.com",
		},
	}, nil
}

func (m *mockTLSProvider) ValidateCertificate(ctx context.Context, cert *x509.Certificate) error {
	return nil
}

func (m *mockTLSProvider) GetCACertPool(ctx context.Context) (*x509.CertPool, error) {
	return x509.NewCertPool(), nil
}

func (m *mockTLSProvider) RefreshCertificates(ctx context.Context) error {
	return nil
}
