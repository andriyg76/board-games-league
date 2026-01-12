package api

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
	"net/http"
	"os"
	"strings"
)

// BuildInfo holds build-time information injected via ldflags
var (
	BuildVersion = "unknown"
	BuildCommit  = "unknown"
	BuildBranch  = "unknown"
	BuildDate    = "unknown"
)

type DiagnosticsHandler struct {
	requestService services.RequestService
	geoIPService   services.GeoIPService
}

func NewDiagnosticsHandler(requestService services.RequestService, geoIPService services.GeoIPService) *DiagnosticsHandler {
	return &DiagnosticsHandler{
		requestService: requestService,
		geoIPService:   geoIPService,
	}
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
		IPAddress string            `json:"ip_address"`
		BaseURL   string            `json:"base_url"`
		UserAgent string            `json:"user_agent"`
		Origin    string            `json:"origin"`
		IsTrusted bool              `json:"is_trusted"`
		GeoInfo   *models.GeoIPInfo `json:"geo_info,omitempty"`
	} `json:"request_info"`
}

func (h *DiagnosticsHandler) GetDiagnosticsHandler(w http.ResponseWriter, r *http.Request) {
	// Check admin status
	claims, ok := r.Context().Value("user").(*user_profile.UserProfile)
	if !ok || claims == nil {
		utils.LogAndWriteHTTPError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"), "unauthorized")
		return
	}

	if !auth.IsSuperAdminByExternalIDs(claims.ExternalIDs) {
		utils.LogAndWriteHTTPError(w, http.StatusForbidden, fmt.Errorf("forbidden"), "admin access required")
		return
	}

	// Get trusted origins from env
	trustedOriginsEnv := os.Getenv("TRUSTED_ORIGINS")
	trustedOrigins := []string{}
	if trustedOriginsEnv != "" {
		trustedOrigins = strings.Split(trustedOriginsEnv, ",")
		for i := range trustedOrigins {
			trustedOrigins[i] = strings.TrimSpace(trustedOrigins[i])
		}
	}

	// Get client IP
	clientIP := h.requestService.GetClientIP(r)
	baseURL := h.requestService.BuildBaseURL(r)
	isTrusted := h.requestService.IsTrustedOrigin(r, trustedOrigins)

	response := DiagnosticsResponse{}
	response.ServerInfo.HostURL = utils.GetHostUrl(r)
	response.ServerInfo.TrustedOrigins = trustedOrigins
	response.BuildInfo.Version = BuildVersion
	response.BuildInfo.Commit = BuildCommit
	response.BuildInfo.Branch = BuildBranch
	response.BuildInfo.Date = BuildDate
	response.RequestInfo.IPAddress = clientIP
	response.RequestInfo.BaseURL = baseURL
	response.RequestInfo.UserAgent = r.Header.Get("User-Agent")
	response.RequestInfo.Origin = r.Header.Get("Origin")
	response.RequestInfo.IsTrusted = isTrusted

	// Get geo info (non-blocking)
	if clientIP != "" {
		if geoInfo, err := h.geoIPService.GetGeoIPInfo(clientIP); err == nil {
			response.RequestInfo.GeoInfo = geoInfo
		} else {
			glog.Warn("Failed to get geo info: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = glog.Error("Failed to encode diagnostics response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
