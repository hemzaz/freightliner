package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentialStore(t *testing.T) {
	store := NewCredentialStore()
	assert.NotNil(t, store)
	assert.Contains(t, store.configPath, ".docker/config.json")
}

func TestCredentialStore_Store(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := NewCredentialStoreWithPath(configPath)

	err := store.Store("registry.io", "user", "pass")
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err)
}

func TestCredentialStore_Get(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := NewCredentialStoreWithPath(configPath)

	registry := "registry.io"
	username := "testuser"
	password := "testpass"

	// Store first
	err := store.Store(registry, username, password)
	require.NoError(t, err)

	// Get
	gotUser, gotPass, err := store.Get(registry)
	require.NoError(t, err)
	assert.Equal(t, username, gotUser)
	assert.Equal(t, password, gotPass)
}

func TestCredentialStore_GetNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := NewCredentialStoreWithPath(configPath)

	_, _, err := store.Get("nonexistent.io")
	assert.Error(t, err)
}

func TestCredentialStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := NewCredentialStoreWithPath(configPath)

	registry := "registry.io"

	// Store first
	err := store.Store(registry, "user", "pass")
	require.NoError(t, err)

	// Delete
	err = store.Delete(registry)
	require.NoError(t, err)

	// Verify deleted
	_, _, err = store.Get(registry)
	assert.Error(t, err)
}

func TestCredentialStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	store := NewCredentialStoreWithPath(configPath)

	// Empty list
	list, err := store.List()
	require.NoError(t, err)
	assert.Empty(t, list)

	// Add registries
	registries := []string{"reg1.io", "reg2.io", "reg3.io"}
	for _, reg := range registries {
		err := store.Store(reg, "user", "pass")
		require.NoError(t, err)
	}

	// List
	list, err = store.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)
	for _, reg := range registries {
		assert.Contains(t, list, reg)
	}
}
