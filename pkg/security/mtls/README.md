# Mutual TLS (mTLS) Implementation Guide

## Overview

This package provides comprehensive interfaces and types for implementing mutual TLS authentication in the Freightliner platform. mTLS provides strong authentication and encryption for service-to-service communication by requiring both client and server to present valid certificates.

## Architecture

### Core Interfaces

1. **TLSProvider** - Certificate and TLS configuration management
2. **MutualTLSAuthenticator** - Client and server authentication
3. **CertificateRotator** - Automated certificate rotation
4. **CertificateManager** - Combined high-level management
5. **CertificateStore** - Persistent certificate storage
6. **TrustManager** - Trust anchor and revocation management

## Implementation Guide

### Step 1: Implement Certificate Storage

```go
package vault

import (
    "context"
    "crypto/x509"

    "github.com/freightliner/pkg/security/mtls"
)

type VaultCertStore struct {
    client *vault.Client
    mount  string
}

func (v *VaultCertStore) Store(ctx context.Context, certID string, cert *x509.Certificate, privateKey interface{}) error {
    // Store certificate and key in Vault
    // Use transit secrets engine for key encryption
    // Store certificate metadata in PKI secrets engine
    return nil
}

func (v *VaultCertStore) Retrieve(ctx context.Context, certID string) (*x509.Certificate, interface{}, error) {
    // Retrieve certificate and key from Vault
    // Decrypt private key using transit engine
    return nil, nil, nil
}
```

### Step 2: Implement TLS Provider

```go
package provider

import (
    "context"
    "crypto/tls"
    "crypto/x509"

    "github.com/freightliner/pkg/security/mtls"
)

type MTLSProvider struct {
    store       mtls.CertificateStore
    trustMgr    mtls.TrustManager
    config      *mtls.TLSConfig
    tlsConfig   *tls.Config
}

func (p *MTLSProvider) GetTLSConfig(ctx context.Context) (*tls.Config, error) {
    if p.tlsConfig != nil {
        return p.tlsConfig, nil
    }

    // Load certificate from store
    cert, key, err := p.store.Retrieve(ctx, p.config.CertificateID)
    if err != nil {
        return nil, err
    }

    // Build TLS config
    tlsCert := tls.Certificate{
        Certificate: [][]byte{cert.Raw},
        PrivateKey:  key,
    }

    // Get CA pool
    caPool, err := p.trustMgr.GetTrustAnchors(ctx)
    if err != nil {
        return nil, err
    }

    p.tlsConfig = &tls.Config{
        Certificates: []tls.Certificate{tlsCert},
        ClientCAs:    caPool,
        ClientAuth:   tls.RequireAndVerifyClientCert,
        MinVersion:   tls.VersionTLS13,
        CipherSuites: []uint16{
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
    }

    return p.tlsConfig, nil
}
```

### Step 3: Implement Authentication

```go
package authenticator

import (
    "context"
    "crypto/x509"

    "github.com/freightliner/pkg/security/mtls"
)

type MTLSAuthenticator struct {
    trustMgr  mtls.TrustManager
    identityExtractor IdentityExtractor
}

func (a *MTLSAuthenticator) AuthenticateClient(ctx context.Context, cert *x509.Certificate) (*mtls.AuthenticatedIdentity, error) {
    // Verify certificate chain
    if err := a.trustMgr.VerifyChain(ctx, cert, nil); err != nil {
        return nil, err
    }

    // Check revocation status
    revoked, err := a.trustMgr.CheckRevocation(ctx, cert)
    if err != nil {
        return nil, err
    }
    if revoked {
        return nil, errors.New("certificate revoked")
    }

    // Extract identity
    identity, err := a.ExtractIdentity(cert)
    if err != nil {
        return nil, err
    }

    // Validate identity against policies
    if err := a.ValidateIdentity(ctx, identity); err != nil {
        return nil, err
    }

    // Build authenticated identity
    authIdentity := &mtls.AuthenticatedIdentity{
        ID:                     identity.Subject.CommonName,
        CommonName:             identity.Subject.CommonName,
        Organization:           identity.Subject.Organization[0],
        CertificateFingerprint: identity.Fingerprint,
        AuthenticatedAt:        time.Now(),
        ExpiresAt:              cert.NotAfter,
    }

    return authIdentity, nil
}
```

### Step 4: Implement Certificate Rotation

```go
package rotation

import (
    "context"
    "time"

    "github.com/freightliner/pkg/security/mtls"
)

type CertRotator struct {
    store     mtls.CertificateStore
    provider  mtls.TLSProvider
    policy    *mtls.RotationPolicy
    callbacks map[string]mtls.RotationCallback
    ticker    *time.Ticker
    stopCh    chan struct{}
}

func (r *CertRotator) StartRotation(ctx context.Context, policy *mtls.RotationPolicy) error {
    r.policy = policy
    r.ticker = time.NewTicker(policy.CheckInterval)
    r.stopCh = make(chan struct{})

    go r.rotationLoop(ctx)

    return nil
}

func (r *CertRotator) rotationLoop(ctx context.Context) {
    for {
        select {
        case <-r.ticker.C:
            r.checkAndRotate(ctx)
        case <-r.stopCh:
            return
        case <-ctx.Done():
            return
        }
    }
}

func (r *CertRotator) checkAndRotate(ctx context.Context) {
    // List all certificates
    certIDs, _ := r.store.List(ctx)

    for _, certID := range certIDs {
        cert, _, err := r.store.Retrieve(ctx, certID)
        if err != nil {
            continue
        }

        // Check if rotation needed
        timeUntilExpiry := time.Until(cert.NotAfter)
        if timeUntilExpiry <= r.policy.RotateBeforeExpiry {
            r.RotateNow(ctx, certID)
        }
    }
}
```

## Integration Points

### HTTP Server with mTLS

```go
package server

import (
    "crypto/tls"
    "net/http"

    "github.com/freightliner/pkg/security/mtls"
)

func NewMTLSServer(addr string, provider mtls.TLSProvider, auth mtls.MutualTLSAuthenticator) (*http.Server, error) {
    tlsConfig, err := provider.GetTLSConfig(context.Background())
    if err != nil {
        return nil, err
    }

    // Add custom verification
    tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        return auth.VerifyPeerCertificate(rawCerts, verifiedChains)
    }

    server := &http.Server{
        Addr:      addr,
        TLSConfig: tlsConfig,
        Handler:   createHandler(auth),
    }

    return server, nil
}

func createHandler(auth mtls.MutualTLSAuthenticator) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        // Extract client certificate
        if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
            http.Error(w, "client certificate required", http.StatusUnauthorized)
            return
        }

        cert := r.TLS.PeerCertificates[0]

        // Authenticate client
        identity, err := auth.AuthenticateClient(r.Context(), cert)
        if err != nil {
            http.Error(w, "authentication failed", http.StatusUnauthorized)
            return
        }

        // Add identity to context
        ctx := context.WithValue(r.Context(), "identity", identity)
        r = r.WithContext(ctx)

        // Handle request
        w.Write([]byte("authenticated"))
    })

    return mux
}
```

### gRPC Client with mTLS

```go
package client

import (
    "context"
    "crypto/tls"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"

    "github.com/freightliner/pkg/security/mtls"
)

func NewMTLSClient(addr string, provider mtls.TLSProvider) (*grpc.ClientConn, error) {
    tlsConfig, err := provider.GetTLSConfig(context.Background())
    if err != nil {
        return nil, err
    }

    creds := credentials.NewTLS(tlsConfig)

    conn, err := grpc.Dial(addr,
        grpc.WithTransportCredentials(creds),
        grpc.WithBlock(),
    )

    return conn, err
}
```

## Best Practices

### Certificate Generation

1. **Key Types**: Prefer ECDSA P-256 or Ed25519 over RSA for better performance
2. **Validity**: Keep certificate validity short (30-90 days) with automated rotation
3. **SANs**: Always include Subject Alternative Names for flexibility
4. **Key Usage**: Specify appropriate key usage extensions

### Security Hardening

1. **TLS Version**: Enforce TLS 1.3 minimum
2. **Cipher Suites**: Use only modern AEAD ciphers
3. **Certificate Pinning**: Pin CA certificates for critical connections
4. **Revocation**: Implement both CRL and OCSP checking
5. **HSM**: Use hardware security modules for CA private keys

### Rotation Strategy

1. **Proactive**: Rotate certificates before expiry (30 days recommended)
2. **Grace Period**: Maintain overlap period where both old and new certs work
3. **Monitoring**: Alert on rotation failures immediately
4. **Rollback**: Keep ability to rollback to previous certificate
5. **Testing**: Test rotation in staging before production

### Monitoring

```go
// Metrics to track
type MTLSMetrics struct {
    CertificatesTotal           int
    CertificatesExpiringSoon    int
    CertificatesExpired         int
    RotationsSuccessful         int64
    RotationsFailed             int64
    AuthenticationAttempts      int64
    AuthenticationFailures      int64
    RevocationChecks            int64
    RevocationCheckFailures     int64
}
```

## Configuration Examples

### Basic Configuration

```yaml
tls:
  certificate_path: /etc/certs/server.crt
  private_key_path: /etc/certs/server.key
  ca_path: /etc/certs/ca.crt
  min_version: "1.3"
  client_auth: require-and-verify
  rotation_enabled: true
  rotation_policy:
    enabled: true
    rotate_before_expiry: 720h  # 30 days
    check_interval: 1h
    auto_renew: true
    renewal_method: vault
    notify_before_expiry: 168h  # 7 days
    notification_channels:
      - slack://ops-channel
      - email://security@example.com
```

### Vault Integration

```yaml
tls:
  vault_enabled: true
  vault_config:
    address: https://vault.example.com
    pki_mount: pki
    role: freightliner-server
    ca_path: /etc/vault/ca.crt
  rotation_policy:
    enabled: true
    rotate_before_expiry: 720h
    renewal_method: vault
```

### HSM Integration

```yaml
tls:
  certificate_path: /etc/certs/server.crt
  use_hsm: true
  hsm_config:
    provider: pkcs11
    library_path: /usr/lib/softhsm/libsofthsm2.so
    pin: ${HSM_PIN}
    slot_id: 0
    key_id: freightliner-server
```

## Testing

### Unit Tests

```go
func TestMTLSAuthentication(t *testing.T) {
    // Create test CA
    ca := createTestCA(t)

    // Issue client certificate
    clientCert := issueClientCert(t, ca)

    // Create authenticator
    auth := &MTLSAuthenticator{
        trustMgr: newTestTrustManager(ca),
    }

    // Test authentication
    identity, err := auth.AuthenticateClient(context.Background(), clientCert)
    assert.NoError(t, err)
    assert.NotNil(t, identity)
    assert.Equal(t, "test-client", identity.CommonName)
}
```

### Integration Tests

```go
func TestMTLSServerClient(t *testing.T) {
    // Setup test server
    server := setupTestMTLSServer(t)
    defer server.Close()

    // Setup test client
    client := setupTestMTLSClient(t)

    // Test request
    resp, err := client.Get(server.URL + "/api")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Troubleshooting

### Common Issues

1. **Certificate Verification Failed**
   - Check certificate chain completeness
   - Verify CA certificate is in trust store
   - Check certificate expiry dates

2. **TLS Handshake Timeout**
   - Verify network connectivity
   - Check firewall rules
   - Verify correct server name (SNI)

3. **Rotation Failures**
   - Check CA connectivity
   - Verify renewal credentials
   - Check disk space for certificate storage

### Debug Commands

```bash
# Verify certificate
openssl x509 -in server.crt -text -noout

# Test TLS connection
openssl s_client -connect server:443 -cert client.crt -key client.key -CAfile ca.crt

# Check certificate expiry
openssl x509 -in server.crt -noout -enddate

# Verify certificate chain
openssl verify -CAfile ca.crt server.crt
```

## Security Considerations

1. **Private Key Protection**: Never log or expose private keys
2. **Certificate Storage**: Use encrypted storage (Vault, HSM)
3. **Rotation Automation**: Automate to prevent manual errors
4. **Revocation Checking**: Always check certificate revocation
5. **Audit Logging**: Log all authentication attempts and failures
6. **Least Privilege**: Issue certificates with minimal permissions
7. **Certificate Transparency**: Consider CT logging for public certificates

## References

- [RFC 8446 - TLS 1.3](https://tools.ietf.org/html/rfc8446)
- [RFC 5280 - X.509 Certificates](https://tools.ietf.org/html/rfc5280)
- [NIST SP 800-52 Rev. 2 - TLS Guidelines](https://csrc.nist.gov/publications/detail/sp/800-52/rev-2/final)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)
