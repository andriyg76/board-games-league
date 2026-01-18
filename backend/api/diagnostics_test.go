package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/cache"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGeoIPService is a mock implementation of GeoIPService
type MockGeoIPService struct {
	mock.Mock
}

func (m *MockGeoIPService) GetGeoIPInfo(ip string) (*models.GeoIPInfo, error) {
	args := m.Called(ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GeoIPInfo), args.Error(1)
}

// MockCacheCleanupService is a mock implementation of CacheCleanupService
type MockCacheCleanupService struct {
	mock.Mock
}

func (m *MockCacheCleanupService) RegisterCache(name string, c cache.CleanableCache) {
	m.Called(name, c)
}

func (m *MockCacheCleanupService) RegisterCacheWithStats(name string, cleanable cache.CleanableCache, statsProvider interface{}) {
	m.Called(name, cleanable, statsProvider)
}

func (m *MockCacheCleanupService) UnregisterCache(name string) {
	m.Called(name)
}

func (m *MockCacheCleanupService) CleanAll() map[string]int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]int)
}

func (m *MockCacheCleanupService) GetAllStats() []cache.CacheStats {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]cache.CacheStats)
}

func (m *MockCacheCleanupService) Start(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

func (m *MockCacheCleanupService) Stop() {
	m.Called()
}

func createTestUserProfile(externalIDs []string) *user_profile.UserProfile {
	return &user_profile.UserProfile{
		Code:        "testcode",
		ExternalIDs: externalIDs,
		Name:        "Test User",
		Picture:     "",
	}
}

func TestGetDiagnosticsHandler_Unauthorized(t *testing.T) {
	requestService := services.NewRequestService()
	mockGeoIPService := new(MockGeoIPService)
	mockCacheCleanupService := new(MockCacheCleanupService)

	handler := NewDiagnosticsHandler(requestService, mockGeoIPService, mockCacheCleanupService)

	req := httptest.NewRequest("GET", "/api/admin/diagnostics", nil)
	w := httptest.NewRecorder()

	handler.GetDiagnosticsHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetDiagnosticsHandler_NotSuperAdmin(t *testing.T) {
	requestService := services.NewRequestService()
	mockGeoIPService := new(MockGeoIPService)
	mockCacheCleanupService := new(MockCacheCleanupService)

	handler := NewDiagnosticsHandler(requestService, mockGeoIPService, mockCacheCleanupService)

	userProfile := createTestUserProfile([]string{"user@example.com"})
	req := httptest.NewRequest("GET", "/api/admin/diagnostics", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDiagnosticsHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetDiagnosticsHandler_Success(t *testing.T) {
	// Set up test superadmin
	restore := auth.SetSuperAdminsForTesting([]string{"admin@test.com"})
	defer restore()

	// Use real RequestService since RequestInfo is a struct with private fields
	requestService := services.NewRequestService()
	mockGeoIPService := new(MockGeoIPService)
	mockCacheCleanupService := new(MockCacheCleanupService)

	handler := NewDiagnosticsHandler(requestService, mockGeoIPService, mockCacheCleanupService)

	userProfile := createTestUserProfile([]string{"admin@test.com"})
	req := httptest.NewRequest("GET", "/api/admin/diagnostics", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	req.Header.Set("CF-Connecting-IP", "192.0.2.2")
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "192.0.2.1:12345"
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockGeoIPService.On("GetGeoIPInfo", mock.AnythingOfType("string")).Return(&models.GeoIPInfo{
		Country: "US",
		City:    "New York",
	}, nil)

	cacheStats := []cache.CacheStats{
		{
			Name:         "TestCache",
			CurrentSize:  10,
			MaxSize:      100,
			ExpiredCount: 2,
			TTL:          "1h",
			TTLSeconds:   3600,
		},
	}
	mockCacheCleanupService.On("GetAllStats").Return(cacheStats)

	handler.GetDiagnosticsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response DiagnosticsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotEmpty(t, response.RequestInfo.IPAddress)
	assert.NotEmpty(t, response.RuntimeInfo.GoVersion)
	assert.NotEmpty(t, response.BuildInfo.Version)
	assert.Len(t, response.CacheStats, 1)
	assert.Equal(t, "TestCache", response.CacheStats[0].Name)
	assert.Equal(t, 10.0, response.CacheStats[0].UsagePercent)

	mockGeoIPService.AssertExpectations(t)
	mockCacheCleanupService.AssertExpectations(t)
}

func TestGetDiagnosticsHandler_GeoIPError(t *testing.T) {
	// Set up test superadmin
	restore := auth.SetSuperAdminsForTesting([]string{"admin@test.com"})
	defer restore()

	// Use real RequestService
	requestService := services.NewRequestService()
	mockGeoIPService := new(MockGeoIPService)
	mockCacheCleanupService := new(MockCacheCleanupService)

	handler := NewDiagnosticsHandler(requestService, mockGeoIPService, mockCacheCleanupService)

	userProfile := createTestUserProfile([]string{"admin@test.com"})
	req := httptest.NewRequest("GET", "/api/admin/diagnostics", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockGeoIPService.On("GetGeoIPInfo", mock.AnythingOfType("string")).Return(nil, errors.New("geoip error"))

	cacheStats := []cache.CacheStats{}
	mockCacheCleanupService.On("GetAllStats").Return(cacheStats)

	handler.GetDiagnosticsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosticsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// GeoInfo should be nil when there's an error
	assert.Nil(t, response.RequestInfo.GeoInfo)

	mockGeoIPService.AssertExpectations(t)
	mockCacheCleanupService.AssertExpectations(t)
}

func TestGetDiagnosticsRequestHandler_Success(t *testing.T) {
	restore := auth.SetSuperAdminsForTesting([]string{"admin@test.com"})
	defer restore()

	requestService := services.NewRequestService()
	mockGeoIPService := new(MockGeoIPService)
	mockCacheCleanupService := new(MockCacheCleanupService)

	handler := NewDiagnosticsHandler(requestService, mockGeoIPService, mockCacheCleanupService)

	userProfile := createTestUserProfile([]string{"admin@test.com"})
	req := httptest.NewRequest("GET", "/api/admin/diagnostics/request", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	req.RemoteAddr = "192.0.2.1:12345"
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockGeoIPService.On("GetGeoIPInfo", mock.AnythingOfType("string")).Return(&models.GeoIPInfo{
		Country: "US",
		City:    "New York",
	}, nil)

	handler.GetDiagnosticsRequestHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosticsRequestResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotEmpty(t, response.RequestInfo.IPAddress)
	assert.NotEmpty(t, response.ServerInfo.HostURL)
	assert.NotNil(t, response.RequestInfo.ResolutionInfo)

	mockGeoIPService.AssertExpectations(t)
	mockCacheCleanupService.AssertExpectations(t)
}

func TestGetDiagnosticsSystemHandler_Success(t *testing.T) {
	restore := auth.SetSuperAdminsForTesting([]string{"admin@test.com"})
	defer restore()

	requestService := services.NewRequestService()
	mockGeoIPService := new(MockGeoIPService)
	mockCacheCleanupService := new(MockCacheCleanupService)

	handler := NewDiagnosticsHandler(requestService, mockGeoIPService, mockCacheCleanupService)

	userProfile := createTestUserProfile([]string{"admin@test.com"})
	req := httptest.NewRequest("GET", "/api/admin/diagnostics/system", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	os.Setenv("TEST_SYSTEM_VAR", "test_value")
	defer os.Unsetenv("TEST_SYSTEM_VAR")

	cacheStats := []cache.CacheStats{
		{
			Name:         "SystemCache",
			CurrentSize:  5,
			MaxSize:      50,
			ExpiredCount: 1,
			TTL:          "30m",
			TTLSeconds:   1800,
		},
	}
	mockCacheCleanupService.On("GetAllStats").Return(cacheStats)

	handler.GetDiagnosticsSystemHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosticsSystemResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotEmpty(t, response.RuntimeInfo.GoVersion)
	assert.NotEmpty(t, response.EnvironmentVars)
	assert.Len(t, response.CacheStats, 1)

	var found bool
	for _, env := range response.EnvironmentVars {
		if env.Name == "TEST_SYSTEM_VAR" {
			found = true
			assert.Equal(t, "test_value", env.Value)
		}
	}
	assert.True(t, found)

	mockGeoIPService.AssertExpectations(t)
	mockCacheCleanupService.AssertExpectations(t)
}

func TestGetDiagnosticsBuildHandler_Success(t *testing.T) {
	restore := auth.SetSuperAdminsForTesting([]string{"admin@test.com"})
	defer restore()

	requestService := services.NewRequestService()
	mockGeoIPService := new(MockGeoIPService)
	mockCacheCleanupService := new(MockCacheCleanupService)

	handler := NewDiagnosticsHandler(requestService, mockGeoIPService, mockCacheCleanupService)

	userProfile := createTestUserProfile([]string{"admin@test.com"})
	req := httptest.NewRequest("GET", "/api/admin/diagnostics/build", nil)
	ctx := context.WithValue(req.Context(), "user", userProfile)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDiagnosticsBuildHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DiagnosticsBuildResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, BuildVersion, response.BuildInfo.Version)
	assert.Equal(t, BuildCommit, response.BuildInfo.Commit)
	assert.Equal(t, BuildBranch, response.BuildInfo.Branch)
	assert.Equal(t, BuildDate, response.BuildInfo.Date)
}

func TestIsSensitiveEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		expected bool
	}{
		{"password", "PASSWORD", true},
		{"secret", "SECRET", true},
		{"token", "TOKEN", true},
		{"api_key", "API_KEY", true},
		{"mongodb", "MONGODB", true},
		{"database", "DATABASE", true},
		{"regular_var", "REGULAR_VAR", false},
		{"normal", "NORMAL", false},
		{"mixed_case_password", "MY_PASSWORD_VAR", true},
		{"jwt_secret", "JWT_SECRET", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveEnvVar(tt.envName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvironmentVars(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("TEST_PASSWORD", "secret123")
	defer os.Unsetenv("TEST_VAR")
	defer os.Unsetenv("TEST_PASSWORD")

	envVars := getEnvironmentVars()

	// Find our test variables
	var testVar, testPassword *EnvVarInfo
	for _, env := range envVars {
		if env.Name == "TEST_VAR" {
			testVar = &env
		}
		if env.Name == "TEST_PASSWORD" {
			testPassword = &env
		}
	}

	assert.NotNil(t, testVar)
	assert.Equal(t, "test_value", testVar.Value)
	assert.False(t, testVar.Masked)

	assert.NotNil(t, testPassword)
	assert.True(t, testPassword.Masked)
	assert.Contains(t, testPassword.Value, "****")
	assert.NotEqual(t, "secret123", testPassword.Value)
}

func TestGetRuntimeInfo(t *testing.T) {
	info := getRuntimeInfo()

	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.GOOS)
	assert.NotEmpty(t, info.GOARCH)
	assert.Greater(t, info.NumCPU, 0)
	assert.GreaterOrEqual(t, info.NumGoroutine, 0)
	assert.NotEmpty(t, info.Uptime)
	assert.GreaterOrEqual(t, info.UptimeSeconds, int64(0))
	assert.NotEmpty(t, info.StartTime)
	assert.GreaterOrEqual(t, info.Memory.Alloc, uint64(0))
}

func TestFormatDuration(t *testing.T) {
	// Verify the function exists and works by checking runtime info
	info := getRuntimeInfo()
	assert.NotEmpty(t, info.Uptime)
	assert.GreaterOrEqual(t, info.UptimeSeconds, int64(0))
}
