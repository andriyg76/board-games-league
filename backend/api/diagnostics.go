package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
)

// BuildInfo holds build-time information injected via ldflags
var (
	BuildVersion = "unknown"
	BuildCommit  = "unknown"
	BuildBranch  = "unknown"
	BuildDate    = "unknown"
)

// StartTime tracks when the application started (for uptime calculation)
var StartTime = time.Now()

// sensitiveEnvPatterns contains patterns for environment variable names that should be masked
var sensitiveEnvPatterns = []string{
	"PASSWORD", "SECRET", "TOKEN", "KEY", "CREDENTIAL", "AUTH",
	"PRIVATE", "API_KEY", "APIKEY", "ACCESS", "MONGO", "DATABASE",
	"DB_", "REDIS", "AWS_", "AZURE_", "GCP_", "OAUTH", "JWT",
}

type DiagnosticsHandler struct {
	requestService     services.RequestService
	geoIPService       services.GeoIPService
	cacheCleanupService services.CacheCleanupService
}

func NewDiagnosticsHandler(requestService services.RequestService, geoIPService services.GeoIPService, cacheCleanupService services.CacheCleanupService) *DiagnosticsHandler {
	return &DiagnosticsHandler{
		requestService:     requestService,
		geoIPService:       geoIPService,
		cacheCleanupService: cacheCleanupService,
	}
}

type RuntimeInfo struct {
	GoVersion    string `json:"go_version"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	Uptime       string `json:"uptime"`
	UptimeSeconds int64 `json:"uptime_seconds"`
	StartTime    string `json:"start_time"`
	Memory       struct {
		Alloc      uint64 `json:"alloc_bytes"`
		TotalAlloc uint64 `json:"total_alloc_bytes"`
		Sys        uint64 `json:"sys_bytes"`
		HeapAlloc  uint64 `json:"heap_alloc_bytes"`
		HeapSys    uint64 `json:"heap_sys_bytes"`
		HeapInuse  uint64 `json:"heap_inuse_bytes"`
		NumGC      uint32 `json:"num_gc"`
	} `json:"memory"`
}

type EnvVarInfo struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Masked  bool   `json:"masked"`
}

type CacheStatsInfo struct {
	Name         string  `json:"name"`
	CurrentSize  int     `json:"current_size"`
	MaxSize      int     `json:"max_size"`
	ExpiredCount int     `json:"expired_count"`
	TTL          string  `json:"ttl"`
	TTLSeconds   int64   `json:"ttl_seconds"`
	UsagePercent float64 `json:"usage_percent"` // (current_size / max_size) * 100
}

type DiagnosticsResponse struct {
	ServerInfo struct {
		HostURL        string   `json:"host_url"`
		TrustedOrigins []string `json:"trusted_origins"`
	} `json:"server_info"`
	BuildInfo struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Branch  string `json:"branch"`
		Date    string `json:"date"`
	} `json:"build_info"`
	RequestInfo struct {
		IPAddress      string            `json:"ip_address"`
		BaseURL        string            `json:"base_url"`
		UserAgent      string            `json:"user_agent"`
		Origin         string            `json:"origin"`
		IsTrusted      bool              `json:"is_trusted"`
		GeoInfo        *models.GeoIPInfo `json:"geo_info,omitempty"`
		ResolutionInfo map[string]string `json:"resolution_info"`
	} `json:"request_info"`
	RuntimeInfo     RuntimeInfo      `json:"runtime_info"`
	EnvironmentVars []EnvVarInfo     `json:"environment_vars"`
	CacheStats      []CacheStatsInfo `json:"cache_stats,omitempty"`
}

func (h *DiagnosticsHandler) GetDiagnosticsHandler(w http.ResponseWriter, r *http.Request) {
	// Check admin status
	claims, ok := r.Context().Value("user").(*user_profile.UserProfile)
	if !ok || claims == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusUnauthorized, fmt.Errorf("unauthorized"), "unauthorized")
		return
	}

	if !auth.IsSuperAdminByExternalIDs(claims.ExternalIDs) {
		utils.LogAndWriteHTTPError(r, w, http.StatusForbidden, fmt.Errorf("forbidden"), "admin access required")
		return
	}

	// Parse request info (trusted origins loaded from TRUSTED_ORIGINS env in config)
	reqInfo := h.requestService.ParseRequest(r)
	config := h.requestService.GetConfig()

	response := DiagnosticsResponse{}
	response.ServerInfo.HostURL = config.HostURL
	if response.ServerInfo.HostURL == "" {
		response.ServerInfo.HostURL = reqInfo.BaseURL()
	}
	response.ServerInfo.TrustedOrigins = config.TrustedOrigins
	response.BuildInfo.Version = BuildVersion
	response.BuildInfo.Commit = BuildCommit
	response.BuildInfo.Branch = BuildBranch
	response.BuildInfo.Date = BuildDate
	response.RequestInfo.IPAddress = reqInfo.ClientIP()
	response.RequestInfo.BaseURL = reqInfo.BaseURL()
	response.RequestInfo.UserAgent = reqInfo.UserAgent()
	response.RequestInfo.Origin = reqInfo.Origin()
	response.RequestInfo.IsTrusted = reqInfo.IsTrustedOrigin()

	// Collect resolution info - headers and request properties used for detection
	resolutionInfo := make(map[string]string)
	resolutionInfo["RemoteAddr"] = r.RemoteAddr
	resolutionInfo["Host"] = r.Host
	resolutionInfo["Protocol"] = reqInfo.Protocol()
	
	// Add relevant headers if present
	headerNames := []string{
		"CF-Connecting-IP", "CF-Visitor", "CF-Ray", "CF-IPCountry",
		"X-Forwarded-For", "X-Forwarded-Proto", "X-Forwarded-Host",
		"X-Forwarded-Scheme", "X-Real-IP", "X-Scheme", "X-Original-Host",
		"True-Client-IP", "Origin", "Referer",
	}
	for _, name := range headerNames {
		if value := r.Header.Get(name); value != "" {
			resolutionInfo[name] = value
		}
	}
	response.RequestInfo.ResolutionInfo = resolutionInfo

	// Get geo info (non-blocking)
	if reqInfo.ClientIP() != "" {
		if geoInfo, err := h.geoIPService.GetGeoIPInfo(reqInfo.ClientIP()); err == nil {
			response.RequestInfo.GeoInfo = geoInfo
		} else {
			glog.Warn("Failed to get geo info: %v", err)
		}
	}

	// Populate Go runtime info
	response.RuntimeInfo = getRuntimeInfo()

	// Populate environment variables (with sensitive values masked)
	response.EnvironmentVars = getEnvironmentVars()

	// Get cache statistics
	if h.cacheCleanupService != nil {
		cacheStats := h.cacheCleanupService.GetAllStats()
		response.CacheStats = make([]CacheStatsInfo, 0, len(cacheStats))

		for _, stats := range cacheStats {
			usagePercent := 0.0
			if stats.MaxSize > 0 {
				usagePercent = float64(stats.CurrentSize) / float64(stats.MaxSize) * 100
			}

			response.CacheStats = append(response.CacheStats, CacheStatsInfo{
				Name:         stats.Name,
				CurrentSize:  stats.CurrentSize,
				MaxSize:      stats.MaxSize,
				ExpiredCount: stats.ExpiredCount,
				TTL:          stats.TTL,
				TTLSeconds:   stats.TTLSeconds,
				UsagePercent: usagePercent,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = glog.Error("Failed to encode diagnostics response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// isSensitiveEnvVar checks if an environment variable name matches sensitive patterns
func isSensitiveEnvVar(name string) bool {
	upperName := strings.ToUpper(name)
	for _, pattern := range sensitiveEnvPatterns {
		if strings.Contains(upperName, pattern) {
			return true
		}
	}
	return false
}

// getRuntimeInfo collects Go runtime information
func getRuntimeInfo() RuntimeInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(StartTime)

	info := RuntimeInfo{
		GoVersion:     runtime.Version(),
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
		NumCPU:        runtime.NumCPU(),
		NumGoroutine:  runtime.NumGoroutine(),
		Uptime:        formatDuration(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
		StartTime:     StartTime.Format(time.RFC3339),
	}
	info.Memory.Alloc = memStats.Alloc
	info.Memory.TotalAlloc = memStats.TotalAlloc
	info.Memory.Sys = memStats.Sys
	info.Memory.HeapAlloc = memStats.HeapAlloc
	info.Memory.HeapSys = memStats.HeapSys
	info.Memory.HeapInuse = memStats.HeapInuse
	info.Memory.NumGC = memStats.NumGC

	return info
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// getEnvironmentVars returns all environment variables with sensitive values masked
func getEnvironmentVars() []EnvVarInfo {
	envVars := os.Environ()
	result := make([]EnvVarInfo, 0, len(envVars))

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		value := parts[1]
		masked := false

		if isSensitiveEnvVar(name) {
			if len(value) > 4 {
				value = value[:2] + "****" + value[len(value)-2:]
			} else {
				value = "****"
			}
			masked = true
		}

		result = append(result, EnvVarInfo{
			Name:   name,
			Value:  value,
			Masked: masked,
		})
	}

	// Sort by name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}
