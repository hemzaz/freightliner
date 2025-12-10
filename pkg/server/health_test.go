package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleHealth tests the basic health endpoint
func TestHandleHealth(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var health HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(t, err)
	assert.Equal(t, "healthy", health.Status)
	assert.NotEmpty(t, health.Uptime)
}

// TestHandleReadiness tests the readiness probe endpoint
func TestHandleReadiness(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest("GET", "/readiness", nil)
	w := httptest.NewRecorder()

	server.handleReadiness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var health HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(t, err)
	assert.Equal(t, "ready", health.Status)
	assert.NotNil(t, health.Checks)
	assert.NotEmpty(t, health.Uptime)
}

// TestHandleLiveness tests the liveness probe endpoint
func TestHandleLiveness(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest("GET", "/liveness", nil)
	w := httptest.NewRecorder()

	server.handleLiveness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var health HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(t, err)
	assert.Equal(t, "alive", health.Status)
	assert.NotNil(t, health.Checks)
	assert.NotNil(t, health.System)
}

// TestHandleSystemInfo tests the system info endpoint
func TestHandleSystemInfo(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest("GET", "/system/info", nil)
	w := httptest.NewRecorder()

	server.handleSystemInfo(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var info map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &info)
	require.NoError(t, err)
	assert.Equal(t, "freightliner", info["service"])
	assert.NotNil(t, info["uptime"])
	assert.NotNil(t, info["system"])
	assert.NotNil(t, info["configuration"])
}

// TestHandleVersion tests the version endpoint
func TestHandleVersion(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()

	server.handleVersion(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var versionInfo map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &versionInfo)
	require.NoError(t, err)
	assert.NotEmpty(t, versionInfo["version"])
	assert.NotEmpty(t, versionInfo["go_version"])
	assert.NotEmpty(t, versionInfo["os_arch"])
}

// TestCheckServiceHealth tests service health checking
func TestCheckServiceHealth(t *testing.T) {
	server := createTestServer(t)

	tests := []struct {
		name             string
		serviceName      string
		serviceAvailable bool
		expectedStatus   string
		expectHealthy    bool
	}{
		{
			name:             "available service",
			serviceName:      "test-service",
			serviceAvailable: true,
			expectedStatus:   "healthy",
			expectHealthy:    true,
		},
		{
			name:             "unavailable service",
			serviceName:      "test-service",
			serviceAvailable: false,
			expectedStatus:   "unhealthy",
			expectHealthy:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.checkServiceHealth(tt.serviceName, tt.serviceAvailable)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.NotEmpty(t, result.Message)
			assert.NotZero(t, result.Duration)
		})
	}
}

// TestGetSystemInfo tests system information collection
func TestGetSystemInfo(t *testing.T) {
	server := createTestServer(t)

	info := server.getSystemInfo()

	assert.NotNil(t, info)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
	assert.Greater(t, info.NumCPU, 0)
	assert.Greater(t, info.NumGoroutine, 0)
	assert.NotNil(t, info.MemoryUsage)
	assert.Greater(t, info.MemoryUsage.Alloc, uint64(0))
}

// TestHealthStatusStructures tests health status data structures
func TestHealthStatusStructures(t *testing.T) {
	// Test HealthStatus
	health := HealthStatus{
		Status:  "healthy",
		Version: "1.0.0",
		Uptime:  "10h",
		Checks:  make(map[string]CheckResult),
		System:  &SystemInfo{},
	}
	assert.Equal(t, "healthy", health.Status)

	// Test CheckResult
	check := CheckResult{
		Status:   "healthy",
		Message:  "OK",
		Duration: 100,
		Error:    "",
	}
	assert.Equal(t, "healthy", check.Status)

	// Test SystemInfo
	sysInfo := SystemInfo{
		GoVersion:    "go1.21",
		OS:           "linux",
		Arch:         "amd64",
		NumCPU:       8,
		NumGoroutine: 10,
		MemoryUsage: &MemoryInfo{
			Alloc:      1024,
			TotalAlloc: 2048,
			Sys:        4096,
			NumGC:      5,
		},
	}
	assert.Equal(t, 8, sysInfo.NumCPU)
	assert.NotNil(t, sysInfo.MemoryUsage)

	// Test MemoryInfo
	memInfo := MemoryInfo{
		Alloc:      1024,
		TotalAlloc: 2048,
		Sys:        4096,
		NumGC:      10,
	}
	assert.Equal(t, uint64(1024), memInfo.Alloc)
	assert.Equal(t, uint32(10), memInfo.NumGC)
}
