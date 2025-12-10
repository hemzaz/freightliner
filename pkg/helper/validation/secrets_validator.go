package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"freightliner/pkg/helper/errors"
)

// SecretsValidator provides comprehensive validation for secrets operations
type SecretsValidator struct {
	// AWS specific validation rules
	awsSecretNameRegex *regexp.Regexp
	// GCP specific validation rules
	gcpSecretNameRegex *regexp.Regexp
	// Maximum secret size limits by provider
	maxSecretSizes map[string]int
}

// SecretValidationConfig contains configuration for secret validation
type SecretValidationConfig struct {
	Provider          string
	MaxSecretSize     int
	AllowEmptyValues  bool
	RequireJSON       bool
	CustomNamePattern string
}

// ValidationError contains detailed validation error information
type ValidationError struct {
	Field      string
	Value      string
	Rule       string
	Message    string
	Suggestion string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s (suggestion: %s)", v.Field, v.Message, v.Suggestion)
}

// NewSecretsValidator creates a new secrets validator with default rules
func NewSecretsValidator() *SecretsValidator {
	return &SecretsValidator{
		// AWS Secrets Manager naming rules: 1-512 characters, alphanumeric, hyphens, underscores, periods, forward slashes
		awsSecretNameRegex: regexp.MustCompile(`^[a-zA-Z0-9/_.-]{1,512}$`),
		// GCP Secret Manager naming rules: 1-255 characters, lowercase letters, numbers, hyphens
		gcpSecretNameRegex: regexp.MustCompile(`^[a-z0-9-]{1,255}$`),
		maxSecretSizes: map[string]int{
			"aws": 65536, // 64KB limit for AWS Secrets Manager
			"gcp": 65536, // 64KB limit for Google Secret Manager
		},
	}
}

// ValidateSecretName validates secret names according to cloud provider naming rules
func (v *SecretsValidator) ValidateSecretName(provider, secretName string) error {
	if secretName == "" {
		return &ValidationError{
			Field:      "secretName",
			Value:      secretName,
			Rule:       "not_empty",
			Message:    "secret name cannot be empty",
			Suggestion: "provide a non-empty secret name",
		}
	}

	// Check for valid UTF-8 encoding
	if !utf8.ValidString(secretName) {
		return &ValidationError{
			Field:      "secretName",
			Value:      secretName,
			Rule:       "utf8_encoding",
			Message:    "secret name contains invalid UTF-8 characters",
			Suggestion: "ensure secret name uses valid UTF-8 encoding",
		}
	}

	switch strings.ToLower(provider) {
	case "aws":
		return v.validateAWSSecretName(secretName)
	case "gcp", "google":
		return v.validateGCPSecretName(secretName)
	default:
		return &ValidationError{
			Field:      "provider",
			Value:      provider,
			Rule:       "supported_provider",
			Message:    fmt.Sprintf("unsupported provider: %s", provider),
			Suggestion: "use 'aws' or 'gcp' as the provider",
		}
	}
}

// validateAWSSecretName validates secret names for AWS Secrets Manager
func (v *SecretsValidator) validateAWSSecretName(secretName string) error {
	if !v.awsSecretNameRegex.MatchString(secretName) {
		return &ValidationError{
			Field:      "secretName",
			Value:      secretName,
			Rule:       "aws_naming",
			Message:    "secret name violates AWS Secrets Manager naming rules",
			Suggestion: "use only alphanumeric characters, hyphens, underscores, periods, and forward slashes (1-512 characters)",
		}
	}

	// Additional AWS-specific validations
	if strings.HasPrefix(secretName, "aws/") {
		return &ValidationError{
			Field:      "secretName",
			Value:      secretName,
			Rule:       "aws_reserved_prefix",
			Message:    "secret name cannot start with 'aws/' (reserved prefix)",
			Suggestion: "choose a name that doesn't start with 'aws/'",
		}
	}

	return nil
}

// validateGCPSecretName validates secret names for Google Secret Manager
func (v *SecretsValidator) validateGCPSecretName(secretName string) error {
	if !v.gcpSecretNameRegex.MatchString(secretName) {
		return &ValidationError{
			Field:      "secretName",
			Value:      secretName,
			Rule:       "gcp_naming",
			Message:    "secret name violates Google Secret Manager naming rules",
			Suggestion: "use only lowercase letters, numbers, and hyphens (1-255 characters)",
		}
	}

	// Additional GCP-specific validations
	if strings.HasPrefix(secretName, "goog") {
		return &ValidationError{
			Field:      "secretName",
			Value:      secretName,
			Rule:       "gcp_reserved_prefix",
			Message:    "secret name cannot start with 'goog' (reserved prefix)",
			Suggestion: "choose a name that doesn't start with 'goog'",
		}
	}

	if strings.HasSuffix(secretName, "-") || strings.HasPrefix(secretName, "-") {
		return &ValidationError{
			Field:      "secretName",
			Value:      secretName,
			Rule:       "gcp_hyphen_position",
			Message:    "secret name cannot start or end with a hyphen",
			Suggestion: "remove leading/trailing hyphens from the secret name",
		}
	}

	return nil
}

// ValidateSecretValue validates secret values with size and format checks
func (v *SecretsValidator) ValidateSecretValue(provider, secretValue string, config *SecretValidationConfig) error {
	if config == nil {
		config = &SecretValidationConfig{
			Provider: provider,
		}
	}

	// Check for empty values
	if secretValue == "" && !config.AllowEmptyValues {
		return &ValidationError{
			Field:      "secretValue",
			Value:      secretValue,
			Rule:       "not_empty",
			Message:    "secret value cannot be empty",
			Suggestion: "provide a non-empty secret value",
		}
	}

	// Check for valid UTF-8 encoding
	if !utf8.ValidString(secretValue) {
		return &ValidationError{
			Field:      "secretValue",
			Value:      "[REDACTED]",
			Rule:       "utf8_encoding",
			Message:    "secret value contains invalid UTF-8 characters",
			Suggestion: "ensure secret value uses valid UTF-8 encoding",
		}
	}

	// Check size limits
	maxSize := config.MaxSecretSize
	if maxSize == 0 {
		if size, exists := v.maxSecretSizes[strings.ToLower(provider)]; exists {
			maxSize = size
		} else {
			maxSize = 65536 // Default 64KB limit
		}
	}

	if len([]byte(secretValue)) > maxSize {
		return &ValidationError{
			Field:      "secretValue",
			Value:      "[REDACTED]",
			Rule:       "size_limit",
			Message:    fmt.Sprintf("secret value exceeds maximum size of %d bytes (got %d bytes)", maxSize, len([]byte(secretValue))),
			Suggestion: fmt.Sprintf("reduce secret value size to under %d bytes", maxSize),
		}
	}

	// Validate JSON format if required
	if config.RequireJSON {
		return v.validateJSONFormat(secretValue)
	}

	return nil
}

// ValidateJSONSecret validates JSON-structured secrets
func (v *SecretsValidator) ValidateJSONSecret(provider, secretName string, data interface{}, config *SecretValidationConfig) error {
	// Validate secret name first
	if err := v.ValidateSecretName(provider, secretName); err != nil {
		return err
	}

	if data == nil {
		return &ValidationError{
			Field:      "data",
			Value:      "nil",
			Rule:       "not_nil",
			Message:    "JSON secret data cannot be nil",
			Suggestion: "provide a valid struct or map to serialize",
		}
	}

	// Marshal to JSON to validate structure and size
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &ValidationError{
			Field:      "data",
			Value:      "[REDACTED]",
			Rule:       "json_marshaling",
			Message:    fmt.Sprintf("failed to marshal data to JSON: %v", err),
			Suggestion: "ensure data structure is JSON-serializable",
		}
	}

	// Validate the resulting JSON string
	jsonConfig := config
	if jsonConfig == nil {
		jsonConfig = &SecretValidationConfig{Provider: provider}
	}
	jsonConfig.RequireJSON = true

	return v.ValidateSecretValue(provider, string(jsonData), jsonConfig)
}

// validateJSONFormat validates that a string contains valid JSON
func (v *SecretsValidator) validateJSONFormat(data string) error {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return &ValidationError{
			Field:      "secretValue",
			Value:      "[REDACTED]",
			Rule:       "json_format",
			Message:    fmt.Sprintf("secret value is not valid JSON: %v", err),
			Suggestion: "ensure secret value contains valid JSON data",
		}
	}
	return nil
}

// ValidateSecretOperation validates complete secret operations
func (v *SecretsValidator) ValidateSecretOperation(operation, provider, secretName, secretValue string, config *SecretValidationConfig) error {
	// Validate operation type
	validOperations := map[string]bool{
		"create": true,
		"update": true,
		"read":   true,
		"delete": true,
	}

	if !validOperations[strings.ToLower(operation)] {
		return &ValidationError{
			Field:      "operation",
			Value:      operation,
			Rule:       "valid_operation",
			Message:    fmt.Sprintf("invalid operation: %s", operation),
			Suggestion: "use one of: create, update, read, delete",
		}
	}

	// Validate secret name for all operations
	if err := v.ValidateSecretName(provider, secretName); err != nil {
		return err
	}

	// For delete and read operations, we don't need to validate the value
	if strings.ToLower(operation) == "delete" || strings.ToLower(operation) == "read" {
		return nil
	}

	// For create and update operations, validate the secret value
	return v.ValidateSecretValue(provider, secretValue, config)
}

// SanitizeSecretName sanitizes secret names to comply with provider rules
func (v *SecretsValidator) SanitizeSecretName(provider, secretName string) (string, error) {
	if secretName == "" {
		return "", errors.InvalidInputf("cannot sanitize empty secret name")
	}

	sanitized := secretName

	switch strings.ToLower(provider) {
	case "aws":
		// Replace invalid characters with underscores
		sanitized = regexp.MustCompile(`[^a-zA-Z0-9/_.-]`).ReplaceAllString(sanitized, "_")
		// Ensure length limits
		if len(sanitized) > 512 {
			sanitized = sanitized[:512]
		}

	case "gcp", "google":
		// Convert to lowercase and replace invalid characters with hyphens
		sanitized = strings.ToLower(sanitized)
		sanitized = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(sanitized, "-")
		// Remove leading/trailing hyphens
		sanitized = strings.Trim(sanitized, "-")
		// Ensure length limits
		if len(sanitized) > 255 {
			sanitized = sanitized[:255]
		}
		// Ensure we don't end with hyphen after truncation
		sanitized = strings.TrimSuffix(sanitized, "-")

	default:
		return "", errors.InvalidInputf("unsupported provider for sanitization: %s", provider)
	}

	// Final validation
	if err := v.ValidateSecretName(provider, sanitized); err != nil {
		return "", errors.Wrap(err, "sanitized name still invalid")
	}

	return sanitized, nil
}

// GetValidationSummary returns a summary of validation rules for a provider
func (v *SecretsValidator) GetValidationSummary(provider string) map[string]interface{} {
	summary := map[string]interface{}{
		"provider": provider,
	}

	switch strings.ToLower(provider) {
	case "aws":
		summary["secret_name_rules"] = map[string]interface{}{
			"pattern":       "^[a-zA-Z0-9/_.-]{1,512}$",
			"min_length":    1,
			"max_length":    512,
			"allowed_chars": "alphanumeric, hyphens, underscores, periods, forward slashes",
			"restrictions":  []string{"cannot start with 'aws/'"},
		}
		summary["secret_value_rules"] = map[string]interface{}{
			"max_size_bytes": 65536,
			"encoding":       "UTF-8",
		}

	case "gcp", "google":
		summary["secret_name_rules"] = map[string]interface{}{
			"pattern":       "^[a-z0-9-]{1,255}$",
			"min_length":    1,
			"max_length":    255,
			"allowed_chars": "lowercase letters, numbers, hyphens",
			"restrictions": []string{
				"cannot start with 'goog'",
				"cannot start or end with hyphen",
			},
		}
		summary["secret_value_rules"] = map[string]interface{}{
			"max_size_bytes": 65536,
			"encoding":       "UTF-8",
		}
	}

	return summary
}
