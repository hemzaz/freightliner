//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"freightliner/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCredentialStore_StoreAndGet tests storing and retrieving credentials
func TestCredentialStore_StoreAndGet(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := auth.NewCredentialStoreWithPath(configPath)

	registry := "test-registry.io"
	username := "testuser"
	password := "testpass"

	// Store credentials
	err := store.Store(registry, username, password)
	require.NoError(t, err)

	// Verify config file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Retrieve credentials
	gotUsername, gotPassword, err := store.Get(registry)
	require.NoError(t, err)
	assert.Equal(t, username, gotUsername)
	assert.Equal(t, password, gotPassword)
}

// TestCredentialStore_Delete tests deleting credentials
func TestCredentialStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := auth.NewCredentialStoreWithPath(configPath)

	registry := "test-registry.io"
	username := "testuser"
	password := "testpass"

	// Store credentials
	err := store.Store(registry, username, password)
	require.NoError(t, err)

	// Delete credentials
	err = store.Delete(registry)
	require.NoError(t, err)

	// Verify credentials are deleted
	_, _, err = store.Get(registry)
	assert.Error(t, err)
}

// TestCredentialStore_List tests listing stored registries
func TestCredentialStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := auth.NewCredentialStoreWithPath(configPath)

	// Store multiple credentials
	registries := []string{
		"registry1.io",
		"registry2.io",
		"registry3.io",
	}

	for _, registry := range registries {
		err := store.Store(registry, "user", "pass")
		require.NoError(t, err)
	}

	// List registries
	list, err := store.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)

	// Verify all registries are in the list
	for _, registry := range registries {
		assert.Contains(t, list, registry)
	}
}

// TestCredentialStore_MultipleRegistries tests storing credentials for multiple registries
func TestCredentialStore_MultipleRegistries(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := auth.NewCredentialStoreWithPath(configPath)

	type creds struct {
		registry string
		username string
		password string
	}

	testCreds := []creds{
		{"docker.io", "user1", "pass1"},
		{"ghcr.io", "user2", "pass2"},
		{"registry.io", "user3", "pass3"},
	}

	// Store all credentials
	for _, c := range testCreds {
		err := store.Store(c.registry, c.username, c.password)
		require.NoError(t, err)
	}

	// Verify all credentials
	for _, c := range testCreds {
		username, password, err := store.Get(c.registry)
		require.NoError(t, err)
		assert.Equal(t, c.username, username)
		assert.Equal(t, c.password, password)
	}
}

// TestCredentialStore_Update tests updating existing credentials
func TestCredentialStore_Update(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := auth.NewCredentialStoreWithPath(configPath)

	registry := "test-registry.io"

	// Store initial credentials
	err := store.Store(registry, "user1", "pass1")
	require.NoError(t, err)

	// Update credentials
	err = store.Store(registry, "user2", "pass2")
	require.NoError(t, err)

	// Verify updated credentials
	username, password, err := store.Get(registry)
	require.NoError(t, err)
	assert.Equal(t, "user2", username)
	assert.Equal(t, "pass2", password)
}

// TestCredentialStore_EmptyStore tests operations on empty store
func TestCredentialStore_EmptyStore(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := auth.NewCredentialStoreWithPath(configPath)

	// List should return empty
	list, err := store.List()
	require.NoError(t, err)
	assert.Empty(t, list)

	// Get should fail
	_, _, err = store.Get("nonexistent.io")
	assert.Error(t, err)

	// Delete should fail
	err = store.Delete("nonexistent.io")
	assert.Error(t, err)
}

// TestCredentialStore_SpecialCharacters tests handling of special characters
func TestCredentialStore_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := auth.NewCredentialStoreWithPath(configPath)

	registry := "test-registry.io"
	username := "user@company.com"
	password := "p@ss:w0rd!#$%"

	// Store credentials with special characters
	err := store.Store(registry, username, password)
	require.NoError(t, err)

	// Retrieve and verify
	gotUsername, gotPassword, err := store.Get(registry)
	require.NoError(t, err)
	assert.Equal(t, username, gotUsername)
	assert.Equal(t, password, gotPassword)
}
