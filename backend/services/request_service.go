package services

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/andriyg76/glog"
)

// RequestInfo contains parsed request information with helper methods
type RequestInfo struct {
	clientIP       string
	protocol       string
	host           string
	baseURL        string
	origin         string
	referer        string
	userAgent      string
	trustedOrigins []string
	configuredURL  *url.URL
	cookieDomain   string
	cookieSecure   bool
}

// RequestService provides methods to parse request information
type RequestService interface {
	// ParseRequest parses the request and returns a RequestInfo with all extracted data
	ParseRequest(r *http.Request, trustedOrigins []string) RequestInfo
	// GetConfig returns the service configuration (HOST_URL based)
	GetConfig() *RequestConfig
}

// RequestConfig holds the service configuration from environment
type RequestConfig struct {
	HostURL      string
	CookieDomain string
	CookieSecure bool
	ConfiguredURL *url.URL
}

type requestService struct {
	config *RequestConfig
}

var (
	globalService     *requestService
	globalServiceOnce sync.Once
)

func NewRequestService() RequestService {
	globalServiceOnce.Do(func() {
		globalService = &requestService{
			config: initConfig(),
		}
	})
	return globalService
}

func initConfig() *RequestConfig {
	cfg := &RequestConfig{}
	
	hostURLEnv := os.Getenv("HOST_URL")
	if hostURLEnv != "" {
		hostURLEnv = strings.TrimSuffix(hostURLEnv, "/")
		cfg.HostURL = hostURLEnv

		parsed, err := url.Parse(hostURLEnv)
		if err == nil && parsed.Host != "" {
			cfg.ConfiguredURL = parsed
			cfg.CookieSecure = parsed.Scheme == "https"

			// Extract domain for cookie (remove port if present)
			host := parsed.Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
			cfg.CookieDomain = host

			glog.Info("RequestService configured with HOST_URL: %s, domain: %s, secure: %v",
				hostURLEnv, cfg.CookieDomain, cfg.CookieSecure)
		} else {
			glog.Warn("Failed to parse HOST_URL: %s, error: %v", hostURLEnv, err)
		}
	} else {
		glog.Info("HOST_URL not configured, will use request-based detection")
	}
	
	return cfg
}

func (s *requestService) GetConfig() *RequestConfig {
	return s.config
}

// ParseRequest extracts all relevant information from the HTTP request
func (s *requestService) ParseRequest(r *http.Request, trustedOrigins []string) RequestInfo {
	info := RequestInfo{
		trustedOrigins: trustedOrigins,
		configuredURL:  s.config.ConfiguredURL,
		cookieDomain:   s.config.CookieDomain,
		cookieSecure:   s.config.CookieSecure,
		userAgent:      r.Header.Get("User-Agent"),
		origin:         r.Header.Get("Origin"),
		referer:        r.Header.Get("Referer"),
	}

	// Parse client IP
	info.clientIP = s.extractClientIP(r)

	// Parse protocol
	info.protocol = s.extractProtocol(r)

	// Parse host
	info.host = s.extractHost(r)

	// Build base URL
	if s.config.HostURL != "" {
		info.baseURL = s.config.HostURL
	} else {
		info.baseURL = info.protocol + "://" + info.host
	}

	return info
}

// extractClientIP extracts the client IP address from the request
// Priority: CF-Connecting-IP > True-Client-IP > X-Forwarded-For > X-Real-IP > RemoteAddr
func (s *requestService) extractClientIP(r *http.Request) string {
	// Check Cloudflare CF-Connecting-IP header (most reliable when behind CF)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// Check True-Client-IP (Cloudflare Enterprise / Akamai)
	if trueClientIP := r.Header.Get("True-Client-IP"); trueClientIP != "" {
		return strings.TrimSpace(trueClientIP)
	}

	// Check X-Forwarded-For header (first IP in the chain)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fall back to RemoteAddr (remove port if present)
	remote := strings.TrimSpace(r.RemoteAddr)
	if host, _, err := net.SplitHostPort(remote); err == nil && host != "" {
		return host
	}
	return remote
}

// extractProtocol determines the protocol (http/https) from request headers
// Priority: HOST_URL config > CF-Visitor > X-Forwarded-Proto > X-Forwarded-Scheme > TLS state
func (s *requestService) extractProtocol(r *http.Request) string {
	// If HOST_URL is configured, use its scheme
	if s.config.ConfiguredURL != nil {
		return s.config.ConfiguredURL.Scheme
	}

	// Check Cloudflare CF-Visitor header (contains JSON like {"scheme":"https"})
	if cfVisitor := r.Header.Get("CF-Visitor"); cfVisitor != "" {
		var visitor struct {
			Scheme string `json:"scheme"`
		}
		if err := json.Unmarshal([]byte(cfVisitor), &visitor); err == nil && visitor.Scheme != "" {
			return visitor.Scheme
		}
	}

	// Check X-Forwarded-Proto header
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.ToLower(strings.TrimSpace(proto))
	}

	// Check X-Forwarded-Scheme header (alternative)
	if scheme := r.Header.Get("X-Forwarded-Scheme"); scheme != "" {
		return strings.ToLower(strings.TrimSpace(scheme))
	}

	// Check X-Scheme header
	if scheme := r.Header.Get("X-Scheme"); scheme != "" {
		return strings.ToLower(strings.TrimSpace(scheme))
	}

	// Fall back to TLS state
	if r.TLS != nil {
		return "https"
	}

	return "http"
}

// extractHost extracts the host from request headers
func (s *requestService) extractHost(r *http.Request) string {
	// Check X-Forwarded-Host header
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		return strings.TrimSpace(forwardedHost)
	}

	// Check X-Original-Host header
	if originalHost := r.Header.Get("X-Original-Host"); originalHost != "" {
		return strings.TrimSpace(originalHost)
	}

	return r.Host
}

// ---- RequestInfo methods ----

// ClientIP returns the client's IP address
func (i RequestInfo) ClientIP() string {
	return i.clientIP
}

// Protocol returns the request protocol (http or https)
func (i RequestInfo) Protocol() string {
	return i.protocol
}

// IsHTTPS returns true if the request uses HTTPS
func (i RequestInfo) IsHTTPS() bool {
	return i.protocol == "https"
}

// Host returns the request host
func (i RequestInfo) Host() string {
	return i.host
}

// BaseURL returns the full base URL (scheme://host)
func (i RequestInfo) BaseURL() string {
	return i.baseURL
}

// HostURL is an alias for BaseURL
func (i RequestInfo) HostURL() string {
	return i.baseURL
}

// Origin returns the Origin header value
func (i RequestInfo) Origin() string {
	return i.origin
}

// Referer returns the Referer header value
func (i RequestInfo) Referer() string {
	return i.referer
}

// UserAgent returns the User-Agent header value
func (i RequestInfo) UserAgent() string {
	return i.userAgent
}

// CookieDomain returns the domain to use for cookies
// Returns empty string if cookies should be set for the current domain only
func (i RequestInfo) CookieDomain() string {
	return i.cookieDomain
}

// CookieSecure returns whether cookies should have the Secure flag
func (i RequestInfo) CookieSecure() bool {
	return i.cookieSecure
}

// IsTrustedOrigin checks if the request origin is trusted
func (i RequestInfo) IsTrustedOrigin() bool {
	// Build effective trusted origins list
	effectiveTrusted := make([]string, len(i.trustedOrigins))
	copy(effectiveTrusted, i.trustedOrigins)

	// Add configured base URL to trusted origins if available
	if i.configuredURL != nil {
		configuredOrigin := i.configuredURL.Scheme + "://" + i.configuredURL.Host
		effectiveTrusted = append(effectiveTrusted, configuredOrigin)
	}

	if len(effectiveTrusted) == 0 {
		return true // If no trusted origins configured, allow all
	}

	origin := i.origin
	if origin == "" {
		// If no Origin header, check Referer header
		if i.referer == "" {
			return false
		}
		if parsed, err := url.Parse(i.referer); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			origin = parsed.Scheme + "://" + parsed.Host
		} else {
			return false
		}
	}

	// Remove trailing slash and compare
	origin = strings.TrimSuffix(origin, "/")
	for _, trusted := range effectiveTrusted {
		trusted = strings.TrimSpace(strings.TrimSuffix(trusted, "/"))
		if trusted != "" && origin == trusted {
			return true
		}
	}

	return false
}

// NewCookie creates an http.Cookie with proper domain and secure settings
func (i RequestInfo) NewCookie(name, value string, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   i.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
	// Only set domain if explicitly configured and not localhost
	if i.cookieDomain != "" && i.cookieDomain != "localhost" {
		cookie.Domain = i.cookieDomain
	}
	return cookie
}

// ClearCookie creates a cookie that clears the specified cookie name
func (i RequestInfo) ClearCookie(name string) *http.Cookie {
	return i.NewCookie(name, "", -1)
}
