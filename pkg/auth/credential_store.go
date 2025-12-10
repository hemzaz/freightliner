// Package auth provides authentication and credential management for container registries.
package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// CredentialStore manages registry credentials compatible with Docker config.json
type CredentialStore struct {
	configPath string
}

// DockerConfig represents the structure of ~/.docker/config.json
type DockerConfig struct {
	Auths       map[string]AuthEntry `json:"auths"`
	CredHelpers map[string]string    `json:"credHelpers,omitempty"`
	CredsStore  string               `json:"credsStore,omitempty"`
}

// AuthEntry represents authentication credentials for a registry
type AuthEntry struct {
	Auth  string `json:"auth,omitempty"`
	Email string `json:"email,omitempty"`
}

var (
	// validHelperName validates that a helper name contains only safe characters
	// Allowed: alphanumeric, hyphens, underscores (no spaces, shell metacharacters, or path separators)
	validHelperName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// knownHelpers is a whitelist of well-known credential helpers
	// This provides defense-in-depth beyond just regex validation
	knownHelpers = map[string]bool{
		"osxkeychain":   true, // macOS Keychain
		"wincred":       true, // Windows Credential Manager
		"secretservice": true, // Linux Secret Service
		"pass":          true, // pass password manager
		"ecr-login":     true, // AWS ECR
		"gcr":           true, // Google GCR
		"acr-env":       true, // Azure ACR
		"gcloud":        true, // Google Cloud SDK
		"desktop":       true, // Docker Desktop
	}
)

// validateHelperName validates a credential helper name to prevent command injection
// Returns error if the helper name contains unsafe characters
func validateHelperName(helper string) error {
	if helper == "" {
		return fmt.Errorf("credential helper name cannot be empty")
	}

	// Check length (reasonable limit to prevent buffer overflow attempts)
	if len(helper) > 64 {
		return fmt.Errorf("credential helper name too long (max 64 characters): %s", helper)
	}

	// Prevent path traversal attempts
	if strings.Contains(helper, "/") || strings.Contains(helper, "\\") {
		return fmt.Errorf("credential helper name cannot contain path separators: %s", helper)
	}

	// Prevent shell metacharacters and command injection
	if strings.ContainsAny(helper, ";|&$`()<>\"'* \t\n") {
		return fmt.Errorf("credential helper name contains invalid characters: %s", helper)
	}

	// Validate against regex pattern (alphanumeric, hyphens, underscores only)
	if !validHelperName.MatchString(helper) {
		return fmt.Errorf("credential helper name contains invalid characters (allowed: alphanumeric, hyphen, underscore): %s", helper)
	}

	// Optional: Warn if helper is not in the known list (but don't block it)
	// This allows for custom helpers while providing visibility
	if !knownHelpers[helper] {
		// Log warning but don't fail - allows custom helpers
		// In production, you might want to log this to security monitoring
		_ = helper // Helper is valid but unknown
	}

	return nil
}

// buildHelperCommand builds a credential helper command name after validation
func buildHelperCommand(helper string) (string, error) {
	// Validate the helper name first
	if err := validateHelperName(helper); err != nil {
		return "", err
	}

	// Use filepath.Base as additional safety layer to prevent path manipulation
	safeHelper := filepath.Base(helper)

	// Double-check after Base() - should be same as input if input was safe
	if safeHelper != helper {
		return "", fmt.Errorf("credential helper name was modified by path cleaning: %s -> %s", helper, safeHelper)
	}

	// Build the command name
	cmdName := "docker-credential-" + safeHelper

	return cmdName, nil
}

// NewCredentialStore creates a new credential store instance
func NewCredentialStore() *CredentialStore {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".docker", "config.json")
	return &CredentialStore{
		configPath: configPath,
	}
}

// NewCredentialStoreWithPath creates a credential store with custom config path
func NewCredentialStoreWithPath(configPath string) *CredentialStore {
	return &CredentialStore{
		configPath: configPath,
	}
}

// Store saves credentials for a registry
func (cs *CredentialStore) Store(registry, username, password string) error {
	// Ensure .docker directory exists
	dir := filepath.Dir(cs.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create new
	config, err := cs.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Encode credentials as base64 (Docker format)
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	// Store credentials
	if config.Auths == nil {
		config.Auths = make(map[string]AuthEntry)
	}

	config.Auths[registry] = AuthEntry{
		Auth: auth,
	}

	// Save config
	return cs.saveConfig(config)
}

// Get retrieves credentials for a registry
func (cs *CredentialStore) Get(registry string) (username, password string, err error) {
	config, err := cs.loadConfig()
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}

	// Check if credential helper is configured
	if config.CredsStore != "" {
		return cs.getFromHelper(config.CredsStore, registry)
	}

	// Check for registry-specific helper
	if helper, ok := config.CredHelpers[registry]; ok {
		return cs.getFromHelper(helper, registry)
	}

	// Get from auths
	authEntry, ok := config.Auths[registry]
	if !ok {
		return "", "", fmt.Errorf("credentials not found for registry: %s", registry)
	}

	// Decode base64 auth
	decoded, err := base64.StdEncoding.DecodeString(authEntry.Auth)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode credentials: %w", err)
	}

	// Split username:password
	parts := []byte(decoded)
	colonIndex := -1
	for i, b := range parts {
		if b == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return "", "", fmt.Errorf("invalid credential format")
	}

	username = string(parts[:colonIndex])
	password = string(parts[colonIndex+1:])

	return username, password, nil
}

// Delete removes credentials for a registry
func (cs *CredentialStore) Delete(registry string) error {
	config, err := cs.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if credential helper is configured
	if config.CredsStore != "" {
		return cs.deleteFromHelper(config.CredsStore, registry)
	}

	// Check for registry-specific helper
	if helper, ok := config.CredHelpers[registry]; ok {
		return cs.deleteFromHelper(helper, registry)
	}

	// Delete from auths
	if config.Auths == nil {
		return fmt.Errorf("credentials not found for registry: %s", registry)
	}

	delete(config.Auths, registry)

	return cs.saveConfig(config)
}

// List returns all registries with stored credentials
func (cs *CredentialStore) List() ([]string, error) {
	config, err := cs.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	registries := make([]string, 0, len(config.Auths))
	for registry := range config.Auths {
		registries = append(registries, registry)
	}

	return registries, nil
}

// loadConfig loads the Docker config from disk
func (cs *CredentialStore) loadConfig() (*DockerConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(cs.configPath); os.IsNotExist(err) {
		// Return empty config
		return &DockerConfig{
			Auths: make(map[string]AuthEntry),
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config DockerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.Auths == nil {
		config.Auths = make(map[string]AuthEntry)
	}

	return &config, nil
}

// saveConfig saves the Docker config to disk
func (cs *CredentialStore) saveConfig(config *DockerConfig) error {
	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(cs.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// getFromHelper retrieves credentials from a credential helper
// Implements the Docker credential helper protocol:
// https://docs.docker.com/engine/reference/commandline/login/#credential-helpers
func (cs *CredentialStore) getFromHelper(helper, registry string) (string, string, error) {
	// Validate and build the credential helper command name (prevents command injection)
	cmdName, err := buildHelperCommand(helper)
	if err != nil {
		return "", "", fmt.Errorf("invalid credential helper name: %w", err)
	}

	// Check if the helper exists
	_, err = exec.LookPath(cmdName)
	if err != nil {
		return "", "", fmt.Errorf("credential helper '%s' not found in PATH: %w", cmdName, err)
	}

	// Execute the helper with 'get' action
	cmd := exec.Command(cmdName, "get")
	cmd.Stdin = strings.NewReader(registry)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("credential helper '%s' failed: %w, stderr: %s", cmdName, err, stderr.String())
	}

	// Parse the response
	// Format: {"ServerURL":"...","Username":"...","Secret":"..."}
	type HelperResponse struct {
		ServerURL string `json:"ServerURL"`
		Username  string `json:"Username"`
		Secret    string `json:"Secret"`
	}

	var response HelperResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return "", "", fmt.Errorf("failed to parse credential helper response: %w", err)
	}

	return response.Username, response.Secret, nil
}

// storeWithHelper stores credentials using a credential helper
func (cs *CredentialStore) storeWithHelper(helper, registry, username, password string) error {
	// Validate and build the credential helper command name (prevents command injection)
	cmdName, err := buildHelperCommand(helper)
	if err != nil {
		return fmt.Errorf("invalid credential helper name: %w", err)
	}

	// Check if the helper exists
	_, err = exec.LookPath(cmdName)
	if err != nil {
		return fmt.Errorf("credential helper '%s' not found in PATH: %w", cmdName, err)
	}

	// Prepare the input for the helper
	// Format: {"ServerURL":"...","Username":"...","Secret":"..."}
	input := map[string]string{
		"ServerURL": registry,
		"Username":  username,
		"Secret":    password,
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Execute the helper with 'store' action
	cmd := exec.Command(cmdName, "store")
	cmd.Stdin = bytes.NewReader(inputJSON)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("credential helper '%s' store failed: %w, stderr: %s", cmdName, err, stderr.String())
	}

	return nil
}

// deleteFromHelper deletes credentials from a credential helper
func (cs *CredentialStore) deleteFromHelper(helper, registry string) error {
	// Validate and build the credential helper command name (prevents command injection)
	cmdName, err := buildHelperCommand(helper)
	if err != nil {
		return fmt.Errorf("invalid credential helper name: %w", err)
	}

	// Check if the helper exists
	_, err = exec.LookPath(cmdName)
	if err != nil {
		return fmt.Errorf("credential helper '%s' not found in PATH: %w", cmdName, err)
	}

	// Execute the helper with 'erase' action
	cmd := exec.Command(cmdName, "erase")
	cmd.Stdin = strings.NewReader(registry)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("credential helper '%s' erase failed: %w, stderr: %s", cmdName, err, stderr.String())
	}

	return nil
}

// listFromHelper lists credentials from a credential helper
func (cs *CredentialStore) listFromHelper(helper string) (map[string]string, error) {
	// Validate and build the credential helper command name (prevents command injection)
	cmdName, err := buildHelperCommand(helper)
	if err != nil {
		return nil, fmt.Errorf("invalid credential helper name: %w", err)
	}

	// Check if the helper exists
	_, err = exec.LookPath(cmdName)
	if err != nil {
		return nil, fmt.Errorf("credential helper '%s' not found in PATH: %w", cmdName, err)
	}

	// Execute the helper with 'list' action
	cmd := exec.Command(cmdName, "list")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("credential helper '%s' list failed: %w, stderr: %s", cmdName, err, stderr.String())
	}

	// Parse the response
	// Format: {"https://index.docker.io/v1/":"username",...}
	var credentials map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse credential helper list response: %w", err)
	}

	return credentials, nil
}

// IsHelperAvailable checks if a credential helper is available in the system
func IsHelperAvailable(helper string) bool {
	// Validate and build the credential helper command name (prevents command injection)
	cmdName, err := buildHelperCommand(helper)
	if err != nil {
		return false // Invalid helper name
	}

	_, err = exec.LookPath(cmdName)
	return err == nil
}

// GetAvailableHelpers returns a list of available credential helpers
func GetAvailableHelpers() []string {
	commonHelpers := []string{
		"osxkeychain",   // macOS Keychain
		"wincred",       // Windows Credential Manager
		"secretservice", // Linux Secret Service (GNOME Keyring, KWallet)
		"pass",          // pass - the standard unix password manager
		"ecr-login",     // AWS ECR credential helper
		"gcr",           // Google GCR credential helper
		"acr-env",       // Azure ACR credential helper
	}

	var available []string
	for _, helper := range commonHelpers {
		if IsHelperAvailable(helper) {
			available = append(available, helper)
		}
	}

	return available
}
