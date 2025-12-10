//go:build cosign
// +build cosign

package cosign

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Policy defines signature verification requirements
type Policy struct {
	// RequireSignature mandates at least one valid signature
	RequireSignature bool `json:"requireSignature" yaml:"requireSignature"`

	// MinSignatures specifies minimum number of valid signatures required
	MinSignatures int `json:"minSignatures" yaml:"minSignatures"`

	// AllowedSigners lists authorized signer identities (email, URI, etc.)
	AllowedSigners []SignerIdentity `json:"allowedSigners,omitempty" yaml:"allowedSigners,omitempty"`

	// DeniedSigners lists blocked signer identities
	DeniedSigners []SignerIdentity `json:"deniedSigners,omitempty" yaml:"deniedSigners,omitempty"`

	// RequireRekor mandates Rekor transparency log verification
	RequireRekor bool `json:"requireRekor" yaml:"requireRekor"`

	// AllowedIssuers lists authorized OIDC issuers for keyless signatures
	AllowedIssuers []string `json:"allowedIssuers,omitempty" yaml:"allowedIssuers,omitempty"`

	// RequireAttestations mandates specific attestation types
	RequireAttestations []AttestationRequirement `json:"requireAttestations,omitempty" yaml:"requireAttestations,omitempty"`

	// KeyRequirements specifies public key requirements
	KeyRequirements *KeyRequirements `json:"keyRequirements,omitempty" yaml:"keyRequirements,omitempty"`

	// EnforcementMode determines policy behavior on failure
	// Options: "enforce" (default), "warn", "audit"
	EnforcementMode string `json:"enforcementMode" yaml:"enforcementMode"`
}

// SignerIdentity defines an authorized or denied signer
type SignerIdentity struct {
	// Email matches against certificate email addresses
	Email string `json:"email,omitempty" yaml:"email,omitempty"`

	// EmailRegex provides pattern matching for emails
	EmailRegex string `json:"emailRegex,omitempty" yaml:"emailRegex,omitempty"`

	// URI matches against certificate URIs (e.g., GitHub workflows)
	URI string `json:"uri,omitempty" yaml:"uri,omitempty"`

	// URIRegex provides pattern matching for URIs
	URIRegex string `json:"uriRegex,omitempty" yaml:"uriRegex,omitempty"`

	// Issuer matches OIDC issuer (e.g., https://token.actions.githubusercontent.com)
	Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty"`

	// Subject matches OIDC subject (combination with issuer)
	Subject string `json:"subject,omitempty" yaml:"subject,omitempty"`

	// PublicKeyFingerprint matches specific public key SHA256 fingerprint
	PublicKeyFingerprint string `json:"publicKeyFingerprint,omitempty" yaml:"publicKeyFingerprint,omitempty"`
}

// AttestationRequirement specifies required attestation types
type AttestationRequirement struct {
	// PredicateType is the attestation predicate type
	// Example: "https://slsa.dev/provenance/v0.2"
	PredicateType string `json:"predicateType" yaml:"predicateType"`

	// MinCount specifies minimum number of attestations required
	MinCount int `json:"minCount" yaml:"minCount"`

	// Policy-specific requirements (e.g., SLSA level)
	Requirements map[string]interface{} `json:"requirements,omitempty" yaml:"requirements,omitempty"`
}

// KeyRequirements specifies public key constraints
type KeyRequirements struct {
	// MinKeySize specifies minimum key size in bits
	MinKeySize int `json:"minKeySize" yaml:"minKeySize"`

	// AllowedAlgorithms lists permitted signature algorithms
	AllowedAlgorithms []string `json:"allowedAlgorithms,omitempty" yaml:"allowedAlgorithms,omitempty"`

	// RequireHardwareKey mandates hardware-backed keys (YubiKey, etc.)
	RequireHardwareKey bool `json:"requireHardwareKey" yaml:"requireHardwareKey"`
}

// PolicyEvaluationResult contains policy evaluation details
type PolicyEvaluationResult struct {
	Passed    bool
	Errors    []string
	Warnings  []string
	Details   map[string]interface{}
	Evaluated int // Number of signatures evaluated
	Valid     int // Number of valid signatures
}

// NewPolicy creates a default policy
func NewPolicy() *Policy {
	return &Policy{
		RequireSignature: true,
		MinSignatures:    1,
		RequireRekor:     false,
		EnforcementMode:  "enforce",
	}
}

// LoadPolicyFromFile loads policy from YAML or JSON file
func LoadPolicyFromFile(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	policy := &Policy{}

	// Try YAML first
	if err := yaml.Unmarshal(data, policy); err != nil {
		// Try JSON
		if jsonErr := json.Unmarshal(data, policy); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse policy (YAML: %v, JSON: %v)", err, jsonErr)
		}
	}

	// Validate policy
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid policy: %w", err)
	}

	return policy, nil
}

// Validate checks policy configuration validity
func (p *Policy) Validate() error {
	if p.MinSignatures < 0 {
		return fmt.Errorf("minSignatures cannot be negative")
	}

	if p.RequireSignature && p.MinSignatures == 0 {
		p.MinSignatures = 1
	}

	// Validate enforcement mode
	switch p.EnforcementMode {
	case "", "enforce", "warn", "audit":
		// Valid modes
		if p.EnforcementMode == "" {
			p.EnforcementMode = "enforce"
		}
	default:
		return fmt.Errorf("invalid enforcement mode: %s (must be enforce, warn, or audit)", p.EnforcementMode)
	}

	// Validate signer identities
	for i, signer := range p.AllowedSigners {
		if err := signer.Validate(); err != nil {
			return fmt.Errorf("invalid allowed signer at index %d: %w", i, err)
		}
	}
	for i, signer := range p.DeniedSigners {
		if err := signer.Validate(); err != nil {
			return fmt.Errorf("invalid denied signer at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate checks signer identity configuration
func (s *SignerIdentity) Validate() error {
	// At least one identity field must be specified
	if s.Email == "" && s.EmailRegex == "" &&
		s.URI == "" && s.URIRegex == "" &&
		s.Issuer == "" && s.Subject == "" &&
		s.PublicKeyFingerprint == "" {
		return fmt.Errorf("signer identity must specify at least one matching criterion")
	}

	// Validate regex patterns
	if s.EmailRegex != "" {
		if _, err := regexp.Compile(s.EmailRegex); err != nil {
			return fmt.Errorf("invalid email regex: %w", err)
		}
	}
	if s.URIRegex != "" {
		if _, err := regexp.Compile(s.URIRegex); err != nil {
			return fmt.Errorf("invalid URI regex: %w", err)
		}
	}

	return nil
}

// Evaluate checks if signatures satisfy the policy
func (p *Policy) Evaluate(ctx context.Context, signatures []Signature) error {
	result := &PolicyEvaluationResult{
		Passed:    true,
		Details:   make(map[string]interface{}),
		Evaluated: len(signatures),
	}

	// Check minimum signature count
	if p.RequireSignature && len(signatures) < p.MinSignatures {
		result.Passed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("insufficient signatures: found %d, required %d", len(signatures), p.MinSignatures))
	}

	// Validate each signature
	validCount := 0
	for i, sig := range signatures {
		if err := p.validateSignature(ctx, &sig, result); err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("signature %d validation warning: %v", i, err))
		} else {
			validCount++
		}
	}

	result.Valid = validCount
	result.Details["validSignatures"] = validCount
	result.Details["totalSignatures"] = len(signatures)

	// Check if enough signatures are valid
	if validCount < p.MinSignatures {
		result.Passed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("insufficient valid signatures: found %d, required %d", validCount, p.MinSignatures))
	}

	// Handle enforcement mode
	return p.handleEvaluationResult(result)
}

// validateSignature validates a single signature against policy
func (p *Policy) validateSignature(ctx context.Context, sig *Signature, result *PolicyEvaluationResult) error {
	// Check Rekor requirement
	if p.RequireRekor && sig.Bundle == nil {
		return fmt.Errorf("signature missing required Rekor bundle")
	}

	// Check allowed issuers for keyless signatures
	if len(p.AllowedIssuers) > 0 && sig.Issuer != "" {
		if !contains(p.AllowedIssuers, sig.Issuer) {
			return fmt.Errorf("issuer %s not in allowed list", sig.Issuer)
		}
	}

	// Check signer allowlist
	if len(p.AllowedSigners) > 0 {
		if !p.matchesSignerList(sig, p.AllowedSigners) {
			return fmt.Errorf("signature does not match any allowed signer")
		}
	}

	// Check signer denylist
	if len(p.DeniedSigners) > 0 {
		if p.matchesSignerList(sig, p.DeniedSigners) {
			return fmt.Errorf("signature matches denied signer")
		}
	}

	// Validate key requirements
	if p.KeyRequirements != nil && sig.Certificate != nil {
		if err := p.validateKeyRequirements(sig.Certificate); err != nil {
			return fmt.Errorf("key requirements not met: %w", err)
		}
	}

	return nil
}

// matchesSignerList checks if signature matches any identity in the list
func (p *Policy) matchesSignerList(sig *Signature, signers []SignerIdentity) bool {
	for _, signer := range signers {
		if signer.Matches(sig) {
			return true
		}
	}
	return false
}

// Matches checks if signature matches this signer identity
func (s *SignerIdentity) Matches(sig *Signature) bool {
	// Check OIDC identity
	if s.Issuer != "" && sig.Issuer != s.Issuer {
		return false
	}
	if s.Subject != "" && sig.Subject != s.Subject {
		return false
	}

	// Check email
	if s.Email != "" || s.EmailRegex != "" {
		if sig.Certificate == nil {
			return false
		}
		if !s.matchesEmail(sig.Certificate.EmailAddresses) {
			return false
		}
	}

	// Check URI
	if s.URI != "" || s.URIRegex != "" {
		if sig.Certificate == nil {
			return false
		}
		var uris []string
		for _, uri := range sig.Certificate.URIs {
			uris = append(uris, uri.String())
		}
		if !s.matchesURI(uris) {
			return false
		}
	}

	return true
}

// matchesEmail checks if any email matches identity
func (s *SignerIdentity) matchesEmail(emails []string) bool {
	for _, email := range emails {
		if s.Email != "" && email == s.Email {
			return true
		}
		if s.EmailRegex != "" {
			if matched, _ := regexp.MatchString(s.EmailRegex, email); matched {
				return true
			}
		}
	}
	return false
}

// matchesURI checks if any URI matches identity
func (s *SignerIdentity) matchesURI(uris []string) bool {
	for _, uri := range uris {
		if s.URI != "" && uri == s.URI {
			return true
		}
		if s.URIRegex != "" {
			if matched, _ := regexp.MatchString(s.URIRegex, uri); matched {
				return true
			}
		}
	}
	return false
}

// validateKeyRequirements checks certificate key requirements
func (p *Policy) validateKeyRequirements(cert *x509.Certificate) error {
	// This is a placeholder - full implementation would check key algorithms,
	// sizes, and hardware backing
	return nil
}

// handleEvaluationResult processes policy evaluation based on enforcement mode
func (p *Policy) handleEvaluationResult(result *PolicyEvaluationResult) error {
	if result.Passed {
		return nil
	}

	errorMsg := fmt.Sprintf("policy evaluation failed: %d error(s), %d warning(s)",
		len(result.Errors), len(result.Warnings))

	for _, err := range result.Errors {
		errorMsg += fmt.Sprintf("\n  - %s", err)
	}

	switch p.EnforcementMode {
	case "enforce":
		return fmt.Errorf(errorMsg)
	case "warn":
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", errorMsg)
		return nil
	case "audit":
		// Log but don't fail
		fmt.Fprintf(os.Stderr, "AUDIT: %s\n", errorMsg)
		return nil
	default:
		return fmt.Errorf(errorMsg)
	}
}

// Helper functions

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
