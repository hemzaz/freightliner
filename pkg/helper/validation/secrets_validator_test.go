package validation

import (
	"strings"
	"testing"
)

func TestSecretsValidator_ValidateSecretName(t *testing.T) {
	validator := NewSecretsValidator()

	tests := []struct {
		name        string
		provider    string
		secretName  string
		expectError bool
		errorRule   string
	}{
		// AWS Tests
		{
			name:        "AWS valid name",
			provider:    "aws",
			secretName:  "my-secret_name.test/path",
			expectError: false,
		},
		{
			name:        "AWS empty name",
			provider:    "aws",
			secretName:  "",
			expectError: true,
			errorRule:   "not_empty",
		},
		{
			name:        "AWS reserved prefix",
			provider:    "aws",
			secretName:  "aws/my-secret",
			expectError: true,
			errorRule:   "aws_reserved_prefix",
		},
		{
			name:        "AWS invalid characters",
			provider:    "aws",
			secretName:  "my-secret@invalid",
			expectError: true,
			errorRule:   "aws_naming",
		},
		{
			name:        "AWS too long",
			provider:    "aws",
			secretName:  strings.Repeat("a", 513),
			expectError: true,
			errorRule:   "aws_naming",
		},
		// GCP Tests
		{
			name:        "GCP valid name",
			provider:    "gcp",
			secretName:  "my-secret-name",
			expectError: false,
		},
		{
			name:        "GCP empty name",
			provider:    "gcp",
			secretName:  "",
			expectError: true,
			errorRule:   "not_empty",
		},
		{
			name:        "GCP reserved prefix",
			provider:    "gcp",
			secretName:  "goog-secret",
			expectError: true,
			errorRule:   "gcp_reserved_prefix",
		},
		{
			name:        "GCP uppercase",
			provider:    "gcp",
			secretName:  "My-Secret",
			expectError: true,
			errorRule:   "gcp_naming",
		},
		{
			name:        "GCP leading hyphen",
			provider:    "gcp",
			secretName:  "-my-secret",
			expectError: true,
			errorRule:   "gcp_hyphen_position",
		},
		{
			name:        "GCP trailing hyphen",
			provider:    "gcp",
			secretName:  "my-secret-",
			expectError: true,
			errorRule:   "gcp_hyphen_position",
		},
		{
			name:        "GCP too long",
			provider:    "gcp",
			secretName:  strings.Repeat("a", 256),
			expectError: true,
			errorRule:   "gcp_naming",
		},
		// Unsupported provider
		{
			name:        "Unsupported provider",
			provider:    "azure",
			secretName:  "my-secret",
			expectError: true,
			errorRule:   "supported_provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSecretName(tt.provider, tt.secretName)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("expected ValidationError but got %T", err)
					return
				}

				if tt.errorRule != "" && validationErr.Rule != tt.errorRule {
					t.Errorf("expected rule %s but got %s", tt.errorRule, validationErr.Rule)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestSecretsValidator_ValidateSecretValue(t *testing.T) {
	validator := NewSecretsValidator()

	tests := []struct {
		name        string
		provider    string
		secretValue string
		config      *SecretValidationConfig
		expectError bool
		errorRule   string
	}{
		{
			name:        "Valid secret value",
			provider:    "aws",
			secretValue: "my-secret-value",
			expectError: false,
		},
		{
			name:        "Empty secret value - not allowed by default",
			provider:    "aws",
			secretValue: "",
			expectError: true,
			errorRule:   "not_empty",
		},
		{
			name:        "Empty secret value - allowed with config",
			provider:    "aws",
			secretValue: "",
			config:      &SecretValidationConfig{AllowEmptyValues: true},
			expectError: false,
		},
		{
			name:        "Secret value too large",
			provider:    "aws",
			secretValue: strings.Repeat("a", 65537),
			expectError: true,
			errorRule:   "size_limit",
		},
		{
			name:        "Valid JSON when required",
			provider:    "aws",
			secretValue: `{"key": "value"}`,
			config:      &SecretValidationConfig{RequireJSON: true},
			expectError: false,
		},
		{
			name:        "Invalid JSON when required",
			provider:    "aws",
			secretValue: `{invalid json}`,
			config:      &SecretValidationConfig{RequireJSON: true},
			expectError: true,
			errorRule:   "json_format",
		},
		{
			name:        "Custom size limit",
			provider:    "aws",
			secretValue: strings.Repeat("a", 1001),
			config:      &SecretValidationConfig{MaxSecretSize: 1000},
			expectError: true,
			errorRule:   "size_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSecretValue(tt.provider, tt.secretValue, tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("expected ValidationError but got %T", err)
					return
				}

				if tt.errorRule != "" && validationErr.Rule != tt.errorRule {
					t.Errorf("expected rule %s but got %s", tt.errorRule, validationErr.Rule)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestSecretsValidator_ValidateJSONSecret(t *testing.T) {
	validator := NewSecretsValidator()

	tests := []struct {
		name        string
		provider    string
		secretName  string
		data        interface{}
		expectError bool
		errorRule   string
	}{
		{
			name:       "Valid JSON secret",
			provider:   "aws",
			secretName: "my-json-secret",
			data: map[string]interface{}{
				"username": "user",
				"password": "pass",
			},
			expectError: false,
		},
		{
			name:        "Nil data",
			provider:    "aws",
			secretName:  "my-json-secret",
			data:        nil,
			expectError: true,
			errorRule:   "not_nil",
		},
		{
			name:       "Invalid secret name",
			provider:   "gcp",
			secretName: "Invalid-Name",
			data: map[string]interface{}{
				"key": "value",
			},
			expectError: true,
			errorRule:   "gcp_naming",
		},
		{
			name:        "Non-serializable data",
			provider:    "aws",
			secretName:  "my-secret",
			data:        make(chan int), // channels can't be JSON marshaled
			expectError: true,
			errorRule:   "json_marshaling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateJSONSecret(tt.provider, tt.secretName, tt.data, nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("expected ValidationError but got %T", err)
					return
				}

				if tt.errorRule != "" && validationErr.Rule != tt.errorRule {
					t.Errorf("expected rule %s but got %s", tt.errorRule, validationErr.Rule)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestSecretsValidator_SanitizeSecretName(t *testing.T) {
	validator := NewSecretsValidator()

	tests := []struct {
		name           string
		provider       string
		secretName     string
		expectedResult string
		expectError    bool
	}{
		{
			name:           "AWS sanitization - replace invalid chars",
			provider:       "aws",
			secretName:     "my@secret#name!",
			expectedResult: "my_secret_name_",
			expectError:    false,
		},
		{
			name:           "AWS sanitization - truncate long name",
			provider:       "aws",
			secretName:     strings.Repeat("a", 600),
			expectedResult: strings.Repeat("a", 512),
			expectError:    false,
		},
		{
			name:           "GCP sanitization - convert to lowercase and replace chars",
			provider:       "gcp",
			secretName:     "My@Secret#Name!",
			expectedResult: "my-secret-name",
			expectError:    false,
		},
		{
			name:           "GCP sanitization - remove leading/trailing hyphens",
			provider:       "gcp",
			secretName:     "@my-secret@",
			expectedResult: "my-secret",
			expectError:    false,
		},
		{
			name:           "GCP sanitization - truncate and fix hyphen",
			provider:       "gcp",
			secretName:     strings.Repeat("a", 300) + "-",
			expectedResult: strings.Repeat("a", 255),
			expectError:    false,
		},
		{
			name:        "Empty name",
			provider:    "aws",
			secretName:  "",
			expectError: true,
		},
		{
			name:        "Unsupported provider",
			provider:    "azure",
			secretName:  "my-secret",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.SanitizeSecretName(tt.provider, tt.secretName)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}

				if result != tt.expectedResult {
					t.Errorf("expected result %s but got %s", tt.expectedResult, result)
				}

				// Verify the sanitized name passes validation
				validateErr := validator.ValidateSecretName(tt.provider, result)
				if validateErr != nil {
					t.Errorf("sanitized name failed validation: %v", validateErr)
				}
			}
		})
	}
}

func TestSecretsValidator_ValidateSecretOperation(t *testing.T) {
	validator := NewSecretsValidator()

	tests := []struct {
		name        string
		operation   string
		provider    string
		secretName  string
		secretValue string
		expectError bool
		errorRule   string
	}{
		{
			name:        "Valid create operation",
			operation:   "create",
			provider:    "aws",
			secretName:  "my-secret",
			secretValue: "secret-value",
			expectError: false,
		},
		{
			name:        "Valid read operation - no value needed",
			operation:   "read",
			provider:    "aws",
			secretName:  "my-secret",
			secretValue: "",
			expectError: false,
		},
		{
			name:        "Valid delete operation - no value needed",
			operation:   "delete",
			provider:    "gcp",
			secretName:  "my-secret",
			secretValue: "",
			expectError: false,
		},
		{
			name:        "Invalid operation",
			operation:   "invalid",
			provider:    "aws",
			secretName:  "my-secret",
			secretValue: "value",
			expectError: true,
			errorRule:   "valid_operation",
		},
		{
			name:        "Invalid secret name",
			operation:   "create",
			provider:    "gcp",
			secretName:  "Invalid-Name",
			secretValue: "value",
			expectError: true,
			errorRule:   "gcp_naming",
		},
		{
			name:        "Create with empty value",
			operation:   "create",
			provider:    "aws",
			secretName:  "my-secret",
			secretValue: "",
			expectError: true,
			errorRule:   "not_empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSecretOperation(tt.operation, tt.provider, tt.secretName, tt.secretValue, nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("expected ValidationError but got %T", err)
					return
				}

				if tt.errorRule != "" && validationErr.Rule != tt.errorRule {
					t.Errorf("expected rule %s but got %s", tt.errorRule, validationErr.Rule)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestSecretsValidator_GetValidationSummary(t *testing.T) {
	validator := NewSecretsValidator()

	tests := []struct {
		name     string
		provider string
	}{
		{
			name:     "AWS validation summary",
			provider: "aws",
		},
		{
			name:     "GCP validation summary",
			provider: "gcp",
		},
		{
			name:     "Google validation summary",
			provider: "google",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := validator.GetValidationSummary(tt.provider)

			if summary == nil {
				t.Errorf("expected summary but got nil")
				return
			}

			if provider, exists := summary["provider"]; !exists || provider != tt.provider {
				t.Errorf("expected provider %s in summary", tt.provider)
			}

			if _, exists := summary["secret_name_rules"]; !exists {
				t.Errorf("expected secret_name_rules in summary")
			}

			if _, exists := summary["secret_value_rules"]; !exists {
				t.Errorf("expected secret_value_rules in summary")
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:      "secretName",
		Value:      "invalid@name",
		Rule:       "aws_naming",
		Message:    "contains invalid characters",
		Suggestion: "use only alphanumeric characters",
	}

	expected := "validation failed for secretName: contains invalid characters (suggestion: use only alphanumeric characters)"
	if err.Error() != expected {
		t.Errorf("expected error message %s but got %s", expected, err.Error())
	}
}
