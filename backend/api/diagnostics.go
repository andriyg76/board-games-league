package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
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

const (
	defaultLogLines = 200
	maxLogLines     = 5000
)

// sensitiveEnvPatterns contains patterns for environment variable names that should be masked
var sensitiveEnvPatterns = []string{
	"PASSWORD", "SECRET", "TOKEN", "KEY", "CREDENTIAL", "AUTH",
	"PRIVATE", "API_KEY", "APIKEY", "ACCESS", "MONGO", "DATABASE",
	"DB_", "REDIS", "AWS_", "AZURE_", "GCP_", "OAUTH", "JWT",
}

type DiagnosticsHandler struct {
	requestService      services.RequestService
	geoIPService        services.GeoIPService
	cacheCleanupService services.CacheCleanupService
}

func NewDiagnosticsHandler(requestService services.RequestService, geoIPService services.GeoIPService, cacheCleanupService services.CacheCleanupService) *DiagnosticsHandler {
	return &DiagnosticsHandler{
		requestService:      requestService,
		geoIPService:        geoIPService,
		cacheCleanupService: cacheCleanupService,
	}
}

type ServerInfo struct {
	HostURL        string   `json:"host_url"`
	TrustedOrigins []string `json:"trusted_origins"`
}

type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Branch  string `json:"branch"`
	Date    string `json:"date"`
}

type RequestInfo struct {
	IPAddress      string            `json:"ip_address"`
	BaseURL        string            `json:"base_url"`
	UserAgent      string            `json:"user_agent"`
	Origin         string            `json:"origin"`
	IsTrusted      bool              `json:"is_trusted"`
	GeoInfo        *models.GeoIPInfo `json:"geo_info,omitempty"`
	ResolutionInfo map[string]string `json:"resolution_info"`
}

type RuntimeInfo struct {
	GoVersion     string `json:"go_version"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	NumCPU        int    `json:"num_cpu"`
	NumGoroutine  int    `json:"num_goroutine"`
	Uptime        string `json:"uptime"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	StartTime     string `json:"start_time"`
	Memory        struct {
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
	Name   string `json:"name"`
	Value  string `json:"value"`
	Masked bool   `json:"masked"`
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

type LogsInfo struct {
	Lines     []string `json:"lines"`
	Requested int      `json:"requested"`
	Returned  int      `json:"returned"`
	Error     string   `json:"error,omitempty"`
}

type DiagnosticsResponse struct {
	ServerInfo      *ServerInfo      `json:"server_info,omitempty"`
	BuildInfo       *BuildInfo       `json:"build_info,omitempty"`
	RequestInfo     *RequestInfo     `json:"request_info,omitempty"`
	RuntimeInfo     *RuntimeInfo     `json:"runtime_info,omitempty"`
	EnvironmentVars []EnvVarInfo     `json:"environment_vars,omitempty"`
	CacheStats      []CacheStatsInfo `json:"cache_stats,omitempty"`
	Logs            *LogsInfo        `json:"logs,omitempty"`
}

func (h *DiagnosticsHandler) GetDiagnosticsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.ensureSuperAdmin(w, r) {
		return
	}

	sections := parseDiagnosticsSections(r.URL.Query()["sections"])
	includeAll := len(sections) == 0
	includeRequest := includeAll || sections["request"]
	includeSystem := includeAll || sections["system"]
	includeBuild := includeAll || sections["build"]
	includeLogs := includeAll || sections["logs"]

	response := DiagnosticsResponse{}

	if includeRequest {
		reqInfo := h.requestService.ParseRequest(r)
		serverInfo := h.buildServerInfo(reqInfo)
		requestInfo := h.buildRequestInfo(r, reqInfo)
		response.ServerInfo = &serverInfo
		response.RequestInfo = &requestInfo
	}

	if includeBuild {
		buildInfo := h.buildBuildInfo()
		response.BuildInfo = &buildInfo
	}

	if includeSystem {
		runtimeInfo := getRuntimeInfo()
		response.RuntimeInfo = &runtimeInfo
		response.EnvironmentVars = getEnvironmentVars()
		response.CacheStats = h.buildCacheStats()
	}

	if includeLogs {
		logsInfo := h.buildLogsInfo(r)
		response.Logs = &logsInfo
	}

	h.writeJSONResponse(w, r, response)
}

func (h *DiagnosticsHandler) ensureSuperAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := r.Context().Value("user").(*user_profile.UserProfile)
	if !ok || claims == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusUnauthorized, fmt.Errorf("unauthorized"), "unauthorized")
		return false
	}

	if !auth.IsSuperAdminByExternalIDs(claims.ExternalIDs) {
		utils.LogAndWriteHTTPError(r, w, http.StatusForbidden, fmt.Errorf("forbidden"), "admin access required")
		return false
	}

	return true
}

func (h *DiagnosticsHandler) buildServerInfo(reqInfo services.RequestInfo) ServerInfo {
	config := h.requestService.GetConfig()
	hostURL := config.HostURL
	if hostURL == "" {
		hostURL = reqInfo.BaseURL()
	}

	return ServerInfo{
		HostURL:        hostURL,
		TrustedOrigins: config.TrustedOrigins,
	}
}

func (h *DiagnosticsHandler) buildBuildInfo() BuildInfo {
	return BuildInfo{
		Version: BuildVersion,
		Commit:  BuildCommit,
		Branch:  BuildBranch,
		Date:    BuildDate,
	}
}

func (h *DiagnosticsHandler) buildRequestInfo(r *http.Request, reqInfo services.RequestInfo) RequestInfo {
	requestInfo := RequestInfo{
		IPAddress: reqInfo.ClientIP(),
		BaseURL:   reqInfo.BaseURL(),
		UserAgent: reqInfo.UserAgent(),
		Origin:    reqInfo.Origin(),
		IsTrusted: reqInfo.IsTrustedOrigin(),
	}

	resolutionInfo := make(map[string]string)
	resolutionInfo["RemoteAddr"] = r.RemoteAddr
	resolutionInfo["Host"] = r.Host
	resolutionInfo["Protocol"] = reqInfo.Protocol()

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
	requestInfo.ResolutionInfo = resolutionInfo

	if reqInfo.ClientIP() != "" {
		if geoInfo, err := h.geoIPService.GetGeoIPInfo(reqInfo.ClientIP()); err == nil {
			requestInfo.GeoInfo = geoInfo
		} else {
			glog.Warn("Failed to get geo info: %v", err)
		}
	}

	return requestInfo
}

func (h *DiagnosticsHandler) buildCacheStats() []CacheStatsInfo {
	if h.cacheCleanupService == nil {
		return nil
	}

	cacheStats := h.cacheCleanupService.GetAllStats()
	responseStats := make([]CacheStatsInfo, 0, len(cacheStats))

	for _, stats := range cacheStats {
		usagePercent := 0.0
		if stats.MaxSize > 0 {
			usagePercent = float64(stats.CurrentSize) / float64(stats.MaxSize) * 100
		}

		responseStats = append(responseStats, CacheStatsInfo{
			Name:         stats.Name,
			CurrentSize:  stats.CurrentSize,
			MaxSize:      stats.MaxSize,
			ExpiredCount: stats.ExpiredCount,
			TTL:          stats.TTL,
			TTLSeconds:   stats.TTLSeconds,
			UsagePercent: usagePercent,
		})
	}

	return responseStats
}

func (h *DiagnosticsHandler) writeJSONResponse(w http.ResponseWriter, r *http.Request, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to encode diagnostics response")
	}
}

func (h *DiagnosticsHandler) buildLogsInfo(r *http.Request) LogsInfo {
	requested := parseLogLines(r.URL.Query().Get("log_lines"))
	logPath, err := getServerLogPath()
	if err != nil {
		return LogsInfo{
			Requested: requested,
			Returned:  0,
			Error:     err.Error(),
		}
	}

	lines, err := readLastLines(logPath, requested)
	if err != nil {
		glog.Warn("Failed to read server logs: %v", err)
		return LogsInfo{
			Requested: requested,
			Returned:  0,
			Error:     "Failed to read server logs",
		}
	}

	return LogsInfo{
		Lines:     lines,
		Requested: requested,
		Returned:  len(lines),
	}
}

func parseDiagnosticsSections(values []string) map[string]bool {
	sections := make(map[string]bool)
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			section := strings.TrimSpace(strings.ToLower(part))
			if section != "" {
				sections[section] = true
			}
		}
	}
	return sections
}

func parseLogLines(value string) int {
	if value == "" {
		return defaultLogLines
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultLogLines
	}
	if parsed > maxLogLines {
		return maxLogLines
	}
	return parsed
}

func getServerLogPath() (string, error) {
	logDir := strings.TrimSpace(os.Getenv("LOG_DIR"))
	if logDir == "" {
		return "", errors.New("LOG_DIR is not configured")
	}
	return filepath.Join(logDir, "server.log"), nil
}

func readLastLines(path string, maxLines int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("server.log not found")
		}
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() == 0 {
		return []string{}, nil
	}

	const chunkSize int64 = 8192
	var (
		offset     = stat.Size()
		linesFound int
		chunks     [][]byte
	)

	for offset > 0 && linesFound <= maxLines {
		readSize := chunkSize
		if offset < readSize {
			readSize = offset
		}
		offset -= readSize

		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}

		chunk := make([]byte, readSize)
		if _, err := io.ReadFull(file, chunk); err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
		linesFound += bytes.Count(chunk, []byte{'\n'})
	}

	var data []byte
	for i := len(chunks) - 1; i >= 0; i-- {
		data = append(data, chunks[i]...)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	return lines, nil
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
