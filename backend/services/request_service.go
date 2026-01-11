package services

import (
	"net/url"
	"net/http"
	"strings"
)

type RequestService interface {
	GetClientIP(r *http.Request) string
	BuildBaseURL(r *http.Request) string
	IsTrustedOrigin(r *http.Request, trustedOrigins []string) bool
}

type requestService struct{}

func NewRequestService() RequestService {
	return &requestService{}
}

// GetClientIP extracts the client IP address from the request.
// It checks X-Forwarded-For, X-Real-IP, and falls back to RemoteAddr.
func (s *requestService) GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (first IP in the chain)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
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
		return realIP
	}

	// Fall back to RemoteAddr (remove port if present)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// BuildBaseURL constructs the base URL from the request.
// It uses X-Forwarded-Proto/X-Forwarded-Host if available, otherwise uses request scheme and host.
func (s *requestService) BuildBaseURL(r *http.Request) string {
	var scheme string
	var host string

	// Check for proxy headers
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	} else {
		host = r.Host
	}

	return scheme + "://" + host
}

// IsTrustedOrigin checks if the request origin is in the trusted origins list.
func (s *requestService) IsTrustedOrigin(r *http.Request, trustedOrigins []string) bool {
	if len(trustedOrigins) == 0 {
		return true // If no trusted origins configured, allow all
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		// If no Origin header, check Referer header
		referer := r.Header.Get("Referer")
		if referer == "" {
			return false
		}
		if parsed, err := url.Parse(referer); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			origin = parsed.Scheme + "://" + parsed.Host
		} else {
			return false
		}
	}

	// Remove trailing slash and compare
	origin = strings.TrimSuffix(origin, "/")
	for _, trusted := range trustedOrigins {
		trusted = strings.TrimSpace(strings.TrimSuffix(trusted, "/"))
		if origin == trusted {
			return true
		}
	}

	return false
}
