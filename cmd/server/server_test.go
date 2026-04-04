//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedProxyHost(t *testing.T) {
	tests := []struct {
		host    string
		allowed bool
	}{
		{"api.megaport.com", true},
		{"api-staging.megaport.com", true},
		{"api-mpone-dev.megaport.com", true},
		{"custom.megaport.com", true},
		{"API.MEGAPORT.COM", true},           // case insensitive
		{" api.megaport.com ", true},          // trimmed
		{"evil.com", false},
		{"megaport.com.evil.com", false},
		{"api.megaport.com.evil.com", false},
		{"", false},
		{"localhost", false},
		{"192.168.1.1", false},
		{"internal-network.local", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			assert.Equal(t, tt.allowed, isAllowedProxyHost(tt.host))
		})
	}
}

func TestIsAllowedOrigin(t *testing.T) {
	tests := []struct {
		origin  string
		allowed bool
	}{
		{"http://localhost:8080", true},
		{"http://localhost:3000", true},
		{"http://localhost", true},
		{"https://localhost:8443", true},
		{"http://127.0.0.1:8080", true},
		{"http://127.0.0.1", true},
		{"http://[::1]:8080", true},
		{"", false},
		{"http://evil.com", false},
		{"http://megaport.com", false},
		{"http://localhost.evil.com", false},
		{"not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			assert.Equal(t, tt.allowed, isAllowedOrigin(tt.origin))
		})
	}
}

func TestSetCORSHeaders(t *testing.T) {
	t.Run("allowed origin gets CORS headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/test", nil)
		r.Header.Set("Origin", "http://localhost:8080")

		setCORSHeaders(w, r, "Content-Type, Authorization")

		assert.Equal(t, "http://localhost:8080", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("disallowed origin gets no CORS headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/test", nil)
		r.Header.Set("Origin", "http://evil.com")

		setCORSHeaders(w, r, "Content-Type, Authorization")

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("no origin gets no CORS headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/test", nil)

		setCORSHeaders(w, r, "Content-Type")

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestProxyHandler_SSRFProtection(t *testing.T) {
	t.Run("rejects disallowed host", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/proxy/test?base=evil.com", nil)

		proxyHandler(w, r)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "proxy target must be a *.megaport.com host")
	})

	t.Run("rejects missing base parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/proxy/test", nil)

		proxyHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("rejects internal network hosts", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/proxy/test?base=192.168.1.1", nil)

		proxyHandler(w, r)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
