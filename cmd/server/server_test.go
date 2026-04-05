//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripFunc allows using a function as an http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// redirectClientTo returns an HTTP client that rewrites all requests to the
// given test backend, preserving path and query. This lets us test handlers
// that hardcode external hostnames (e.g. api.megaport.com).
func redirectClientTo(backend *httptest.Server) *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = backend.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(req)
		}),
	}
}

func TestIsAllowedProxyHost(t *testing.T) {
	tests := []struct {
		host    string
		allowed bool
	}{
		{"api.megaport.com", true},
		{"api-staging.megaport.com", true},
		{"api-mpone-dev.megaport.com", true},
		{"custom.megaport.com", true},
		{"API.MEGAPORT.COM", true},   // case insensitive
		{" api.megaport.com ", true}, // trimmed
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
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
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

// newTestServer creates a server.Server for testing with a 1-hour session duration.
func newTestServer() *server.Server {
	return server.NewServer(1*time.Hour, log.New(io.Discard, "", 0))
}

func TestAuthenticatedProxyHandler_MissingToken(t *testing.T) {
	srv := newTestServer()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/products", nil)
	// No X-Session-Token header

	authenticatedProxyHandler(w, r, srv)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing session token")
}

func TestAuthenticatedProxyHandler_InvalidToken(t *testing.T) {
	srv := newTestServer()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/products", nil)
	r.Header.Set("X-Session-Token", "invalid-token-that-does-not-exist")

	authenticatedProxyHandler(w, r, srv)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired session")
}

func TestAuthenticatedProxyHandler_ValidSession(t *testing.T) {
	// Create a mock backend that the proxy will forward to
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the Authorization header was set
		assert.Equal(t, "Bearer test-megaport-token", r.Header.Get("Authorization"))
		// Verify X-Session-Token was NOT forwarded
		assert.Empty(t, r.Header.Get("X-Session-Token"))
		// Verify path was forwarded correctly
		assert.Equal(t, "/v2/products", r.URL.Path)
		w.Header().Set("X-Custom-Header", "from-backend")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer backend.Close()

	srv := newTestServer()

	// Create a session
	session, err := srv.GetSessionManager().CreateSession("test-access-key", "test-secret-key")
	require.NoError(t, err)

	// Set the megaport token on the session
	srv.GetSessionManager().UpdateMegaportToken(session.ID, "test-megaport-token", time.Now().Add(1*time.Hour))

	// Override the optimized HTTP client to redirect to test backend
	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/products", nil)
	r.Header.Set("X-Session-Token", session.ID)
	r.Header.Set("Origin", "http://localhost:8080")

	authenticatedProxyHandler(w, r, srv)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "from-backend", w.Header().Get("X-Custom-Header"))
	assert.Contains(t, w.Body.String(), `{"status":"ok"}`)
	// CORS headers should be set
	assert.Equal(t, "http://localhost:8080", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestAuthenticatedProxyHandler_ValidSessionWithQueryParams(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/products", r.URL.Path)
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer backend.Close()

	srv := newTestServer()
	session, err := srv.GetSessionManager().CreateSession("test-key", "test-secret")
	require.NoError(t, err)
	srv.GetSessionManager().UpdateMegaportToken(session.ID, "tok", time.Now().Add(1*time.Hour))

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/products?limit=10", nil)
	r.Header.Set("X-Session-Token", session.ID)

	authenticatedProxyHandler(w, r, srv)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthenticatedProxyHandler_StagingHost(t *testing.T) {
	var receivedHost string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	srv := newTestServer()
	// Access key containing "staging" triggers staging host selection
	session, err := srv.GetSessionManager().CreateSession("staging-key", "secret")
	require.NoError(t, err)
	srv.GetSessionManager().UpdateMegaportToken(session.ID, "tok", time.Now().Add(1*time.Hour))

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	// Capture the original URL host before redirect rewrites it
	optimizedHTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			receivedHost = req.URL.Host
			req.URL.Scheme = "http"
			req.URL.Host = backend.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/products", nil)
	r.Header.Set("X-Session-Token", session.ID)

	authenticatedProxyHandler(w, r, srv)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "api-staging.megaport.com", receivedHost)
}

func TestAuthenticatedProxyHandler_POSTSetsContentType(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	srv := newTestServer()
	session, err := srv.GetSessionManager().CreateSession("key", "secret")
	require.NoError(t, err)
	srv.GetSessionManager().UpdateMegaportToken(session.ID, "tok", time.Now().Add(1*time.Hour))

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v2/products", strings.NewReader(`{"name":"test"}`))
	r.Header.Set("X-Session-Token", session.ID)

	authenticatedProxyHandler(w, r, srv)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthenticatedProxyHandler_TokenNearExpiry(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	srv := newTestServer()
	session, err := srv.GetSessionManager().CreateSession("key", "secret")
	require.NoError(t, err)
	// Token expires in 2 minutes — within the 5-minute refresh window
	srv.GetSessionManager().UpdateMegaportToken(session.ID, "expiring-tok", time.Now().Add(2*time.Minute))

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/products", nil)
	r.Header.Set("X-Session-Token", session.ID)

	authenticatedProxyHandler(w, r, srv)

	// Should still succeed — the token refresh path logs but doesn't block
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthenticatedProxyHandler_ExpiredSession(t *testing.T) {
	// Create server with a short-but-stable session duration
	srv := server.NewServer(100*time.Millisecond, log.New(io.Discard, "", 0))

	session, err := srv.GetSessionManager().CreateSession("key", "secret")
	require.NoError(t, err)

	// Wait long enough for the session to expire without relying on millisecond-level scheduling
	time.Sleep(500 * time.Millisecond)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/products", nil)
	r.Header.Set("X-Session-Token", session.ID)

	authenticatedProxyHandler(w, r, srv)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired session")
}

func TestProxyHandler_SuccessfulProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/products", r.URL.Path)
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		// 'base' param should have been stripped
		assert.Empty(t, r.URL.Query().Get("base"))
		w.Header().Set("X-Backend-Header", "present")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer backend.Close()

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/proxy/v2/products?base=api.megaport.com&limit=10", nil)
	r.Header.Set("Origin", "http://localhost:8080")

	proxyHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "present", w.Header().Get("X-Backend-Header"))
	assert.Contains(t, w.Body.String(), `{"data":[]}`)
	// CORS headers should be set
	assert.Equal(t, "http://localhost:8080", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestProxyHandler_POSTSetsContentType(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/proxy/v2/products?base=api.megaport.com", strings.NewReader(`{}`))

	proxyHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProxyHandler_CopiesRequestHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
		// Host header should NOT be forwarded
		assert.NotEqual(t, "evil-host.com", r.Host)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/proxy/v2/products?base=api.megaport.com", nil)
	r.Header.Set("Authorization", "Bearer mytoken")
	r.Host = "evil-host.com"

	proxyHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProxyHandler_NoQueryParamsExceptBase(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No extra query params — query string should be empty
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/proxy/v2/test?base=api.megaport.com", nil)

	proxyHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProxyHandler_ForwardsQueryParams(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Empty(t, r.URL.Query().Get("base"))
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	origClient := optimizedHTTPClient
	defer func() { optimizedHTTPClient = origClient }()
	optimizedHTTPClient = redirectClientTo(backend)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/proxy/v2/products?base=api.megaport.com&limit=10", nil)

	proxyHandler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddCorsHeaders_OptionsPreflightReturns200(t *testing.T) {
	// Create a dummy file server handler
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called for OPTIONS preflight")
	})

	handler := addCorsHeaders(inner)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/index.html", nil)
	r.Header.Set("Origin", "http://localhost:8080")

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:8080", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestAddCorsHeaders_SetsMIMETypes(t *testing.T) {
	tests := []struct {
		path        string
		contentType string
	}{
		{"/app.wasm", "application/wasm"},
		{"/main.js", "application/javascript"},
		{"/index.html", "text/html; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			innerCalled := false
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				innerCalled = true
				// Verify content-type was set before inner handler runs
				assert.Equal(t, tt.contentType, w.Header().Get("Content-Type"))
			})

			handler := addCorsHeaders(inner)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.path, nil)
			r.Header.Set("Origin", "http://localhost:8080")

			handler.ServeHTTP(w, r)

			assert.True(t, innerCalled, "inner handler should be called for GET %s", tt.path)
		})
	}
}

func TestAddCorsHeaders_NoMIMETypeOverrideForOtherFiles(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content-Type should NOT be pre-set for .css files
		assert.Empty(t, w.Header().Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	})

	handler := addCorsHeaders(inner)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/style.css", nil)
	r.Header.Set("Origin", "http://localhost:8080")

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddCorsHeaders_DisallowedOriginNoCORS(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := addCorsHeaders(inner)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/index.html", nil)
	r.Header.Set("Origin", "http://evil.com")

	handler.ServeHTTP(w, r)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestAddCorsHeaders_GETPassesThroughToInner(t *testing.T) {
	innerCalled := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := addCorsHeaders(inner)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/index.html", nil)

	handler.ServeHTTP(w, r)

	assert.True(t, innerCalled, "inner handler should be called for GET requests")
}
