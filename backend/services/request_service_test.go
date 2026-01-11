package services

import (
	"net/http/httptest"
	"testing"
)

func TestRequestService_IsTrustedOrigin_RefererFallbackParsesOrigin(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/diagnostics", nil)
	r.Header.Del("Origin")
	r.Header.Set("Referer", "https://example.com/page?id=1")

	if !svc.IsTrustedOrigin(r, []string{"https://example.com"}) {
		t.Fatalf("expected referer fallback to match trusted origin")
	}
}

func TestRequestService_IsTrustedOrigin_RefererFallbackInvalidURL(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/diagnostics", nil)
	r.Header.Del("Origin")
	r.Header.Set("Referer", "not a url")

	if svc.IsTrustedOrigin(r, []string{"https://example.com"}) {
		t.Fatalf("expected invalid referer to be untrusted")
	}
}

func TestRequestService_IsTrustedOrigin_OriginHeaderStillWorks(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/diagnostics", nil)
	r.Header.Set("Origin", "https://example.com")

	if !svc.IsTrustedOrigin(r, []string{"https://example.com"}) {
		t.Fatalf("expected origin header to match trusted origin")
	}
}

func TestRequestService_IsTrustedOrigin_NoTrustedOriginsAllowsAll(t *testing.T) {
	svc := NewRequestService()

	r := httptest.NewRequest("GET", "http://server.local/diagnostics", nil)
	r.Header.Set("Origin", "https://evil.example")

	if !svc.IsTrustedOrigin(r, nil) {
		t.Fatalf("expected allow-all when trusted origins not configured")
	}
}
