package services

import (
	"net/http/httptest"
	"testing"
)

func TestRequestInfo_IsTrustedOrigin_NoTrustedOriginsAllowsAll(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/diagnostics", nil)
	r.Header.Set("Origin", "https://evil.example")

	reqInfo := svc.ParseRequest(r)
	// When no trusted origins configured, all origins are allowed
	if !reqInfo.IsTrustedOrigin() {
		t.Fatalf("expected allow-all when trusted origins not configured")
	}
}

func TestRequestInfo_ClientIP_CFConnectingIP(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)
	r.Header.Set("CF-Connecting-IP", "1.2.3.4")
	r.Header.Set("X-Forwarded-For", "5.6.7.8")
	r.RemoteAddr = "9.10.11.12:1234"

	reqInfo := svc.ParseRequest(r)
	if reqInfo.ClientIP() != "1.2.3.4" {
		t.Fatalf("expected CF-Connecting-IP to take priority, got %s", reqInfo.ClientIP())
	}
}

func TestRequestInfo_ClientIP_XForwardedFor(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)
	r.Header.Set("X-Forwarded-For", "5.6.7.8, 1.2.3.4")
	r.RemoteAddr = "9.10.11.12:1234"

	reqInfo := svc.ParseRequest(r)
	if reqInfo.ClientIP() != "5.6.7.8" {
		t.Fatalf("expected first IP from X-Forwarded-For, got %s", reqInfo.ClientIP())
	}
}

func TestRequestInfo_ClientIP_RemoteAddr(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)
	r.RemoteAddr = "9.10.11.12:1234"

	reqInfo := svc.ParseRequest(r)
	if reqInfo.ClientIP() != "9.10.11.12" {
		t.Fatalf("expected RemoteAddr without port, got %s", reqInfo.ClientIP())
	}
}

func TestRequestInfo_Protocol_CFVisitor(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)
	r.Header.Set("CF-Visitor", `{"scheme":"https"}`)

	reqInfo := svc.ParseRequest(r)
	if reqInfo.Protocol() != "https" {
		t.Fatalf("expected https from CF-Visitor, got %s", reqInfo.Protocol())
	}
}

func TestRequestInfo_Protocol_XForwardedProto(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)
	r.Header.Set("X-Forwarded-Proto", "https")

	reqInfo := svc.ParseRequest(r)
	if reqInfo.Protocol() != "https" {
		t.Fatalf("expected https from X-Forwarded-Proto, got %s", reqInfo.Protocol())
	}
}

func TestRequestInfo_IsHTTPS(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)
	r.Header.Set("X-Forwarded-Proto", "https")

	reqInfo := svc.ParseRequest(r)
	if !reqInfo.IsHTTPS() {
		t.Fatalf("expected IsHTTPS to be true")
	}
}

func TestRequestInfo_NewCookie(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)

	reqInfo := svc.ParseRequest(r)
	cookie := reqInfo.NewCookie("test", "value", 3600)

	if cookie.Name != "test" {
		t.Fatalf("expected cookie name 'test', got %s", cookie.Name)
	}
	if cookie.Value != "value" {
		t.Fatalf("expected cookie value 'value', got %s", cookie.Value)
	}
	if cookie.MaxAge != 3600 {
		t.Fatalf("expected cookie MaxAge 3600, got %d", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Fatalf("expected cookie HttpOnly to be true")
	}
}

func TestRequestInfo_ClearCookie(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)

	reqInfo := svc.ParseRequest(r)
	cookie := reqInfo.ClearCookie("test")

	if cookie.Name != "test" {
		t.Fatalf("expected cookie name 'test', got %s", cookie.Name)
	}
	if cookie.Value != "" {
		t.Fatalf("expected empty cookie value, got %s", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Fatalf("expected cookie MaxAge -1, got %d", cookie.MaxAge)
	}
}

func TestRequestInfo_CookieSecure_HTTPS(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)
	r.Header.Set("X-Forwarded-Proto", "https")

	reqInfo := svc.ParseRequest(r)
	if !reqInfo.CookieSecure() {
		t.Fatalf("expected CookieSecure to be true for HTTPS")
	}
}

func TestRequestInfo_CookieSecure_HTTP(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/test", nil)

	reqInfo := svc.ParseRequest(r)
	if reqInfo.CookieSecure() {
		t.Fatalf("expected CookieSecure to be false for HTTP")
	}
}

func TestRequestInfo_CookieDomain(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local:8080/test", nil)

	reqInfo := svc.ParseRequest(r)
	if reqInfo.CookieDomain() != "server.local" {
		t.Fatalf("expected CookieDomain 'server.local', got %s", reqInfo.CookieDomain())
	}
}
