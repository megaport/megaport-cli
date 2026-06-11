//go:build !js && !wasm

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/megaport/megaport-cli/internal/server"
)

// isAllowedOrigin checks whether the CORS origin is from localhost.
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	hostname := u.Hostname()
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}

// setCORSHeaders sets CORS headers only for allowed origins.
func setCORSHeaders(w http.ResponseWriter, r *http.Request, allowedHeaders string) {
	origin := r.Header.Get("Origin")
	if isAllowedOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
	}
}

// setSecurityHeaders sets defense-in-depth security headers on all responses.
func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' https://*.megaport.com")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

// withSecurityHeaders wraps an http.HandlerFunc to add security headers.
func withSecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)
		next(w, r)
	}
}

// rateLimiter implements a simple sliding window rate limiter.
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

// newRateLimiter creates a rate limiter that allows limit requests per window.
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow returns true if the key is within the rate limit.
func (rl *rateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter to only recent attempts
	var recent []time.Time
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit {
		rl.attempts[key] = recent
		return false
	}

	rl.attempts[key] = append(recent, now)
	return true
}

// cleanup removes stale entries from the rate limiter map.
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window)
	for key, attempts := range rl.attempts {
		var recent []time.Time
		for _, t := range attempts {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(rl.attempts, key)
		} else {
			rl.attempts[key] = recent
		}
	}
}

// startCleanup runs periodic cleanup of stale rate limiter entries.
func (rl *rateLimiter) startCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
}

// Optimized HTTP client with connection pooling
var optimizedHTTPClient = &http.Client{
	Timeout: 60 * time.Second, // Increased timeout for large responses
	Transport: &http.Transport{
		// Connection pooling settings
		MaxIdleConns:        100,              // Maximum idle connections across all hosts
		MaxIdleConnsPerHost: 10,               // Maximum idle connections per host
		MaxConnsPerHost:     20,               // Maximum total connections per host
		IdleConnTimeout:     90 * time.Second, // How long idle connections stay open

		// Performance optimizations
		DisableCompression: false, // Enable gzip compression
		DisableKeepAlives:  false, // Enable HTTP keep-alive
		ForceAttemptHTTP2:  false, // Disable HTTP/2 to avoid timeout issues

		// Timeout settings - increased for large responses
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second, // Increased for slow APIs
		ExpectContinueTimeout: 1 * time.Second,

		// Connection settings
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
}

func main() {
	port := flag.String("port", "8080", "Port to serve on")
	bind := flag.String("bind", "127.0.0.1", "Address to bind to (use 0.0.0.0 to expose on all interfaces)")
	webDir := flag.String("dir", "web", "Directory to serve files from")
	sessionDuration := flag.Duration("session-duration", 1*time.Hour, "Session duration")
	flag.Parse()

	// Create server with session management
	srv := server.NewServer(*sessionDuration, log.Default())

	// Rate limiter for login endpoint: 10 attempts per minute per IP
	loginLimiter := newRateLimiter(10, 1*time.Minute)
	loginLimiter.startCleanup(5 * time.Minute)

	// Authentication endpoints
	http.HandleFunc("/auth/login", withSecurityHeaders(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip == "" {
			ip = r.RemoteAddr
		}
		if !loginLimiter.Allow(ip) {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(loginLimiter.window.Seconds())))
			http.Error(w, "Too many login attempts, please try again later", http.StatusTooManyRequests)
			return
		}
		srv.HandleLogin(w, r)
	}))
	http.HandleFunc("/auth/logout", withSecurityHeaders(srv.HandleLogout))
	http.HandleFunc("/auth/check", withSecurityHeaders(srv.HandleSessionCheck))

	// Authenticated API proxy
	http.HandleFunc("/api/", withSecurityHeaders(func(w http.ResponseWriter, r *http.Request) {
		authenticatedProxyHandler(w, r, srv)
	}))

	// Static file server for everything else
	fs := http.FileServer(http.Dir(*webDir))
	http.Handle("/", withSecurityHeaders(addCorsHeaders(fs)))

	addr := net.JoinHostPort(*bind, *port)
	log.Printf("Starting Megaport CLI WASM Server on http://%s", addr)
	log.Printf("Serving files from: %s", *webDir)
	log.Printf("Session duration: %v", *sessionDuration)
	log.Printf("\nEndpoints:")
	log.Printf("  - POST /auth/login    - Customer login")
	log.Printf("  - POST /auth/logout   - Customer logout")
	log.Printf("  - GET  /auth/check    - Check session validity")
	log.Printf("  - *    /api/*         - Authenticated API proxy")
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func authenticatedProxyHandler(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	// Get session token from header
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		http.Error(w, "Unauthorized: Missing session token", http.StatusUnauthorized)
		return
	}

	// Validate session
	session := srv.GetSessionManager().GetSession(sessionToken)
	if session == nil {
		http.Error(w, "Unauthorized: Invalid or expired session", http.StatusUnauthorized)
		return
	}

	// Update session activity
	srv.GetSessionManager().UpdateActivity(sessionToken)

	// Check if Megaport token needs refresh
	if time.Now().After(session.TokenExpiry.Add(-5 * time.Minute)) {
		log.Printf("Token expiring soon, refreshing...")
		// Refresh token logic would go here
		// For now, we'll let it use the existing token
	}

	// Extract the path after /api/
	path := strings.TrimPrefix(r.URL.Path, "/api/")

	// Determine target host based on environment
	targetHost := "api.megaport.com" // Default to production
	if strings.Contains(session.AccessKey, "staging") {
		targetHost = "api-staging.megaport.com"
	}

	// Build the target URL
	targetURL := "https://" + targetHost + "/" + path

	// Forward query parameters
	if len(r.URL.Query()) > 0 {
		targetURL += "?" + r.URL.RawQuery
	}

	log.Printf("Authenticated proxy: %s %s (session: %s)", r.Method, targetURL, sessionToken[:8]+"...")

	// Create proxy request
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		log.Printf("Failed to create proxy request: %v", err)
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers from original request
	for key, values := range r.Header {
		if key != "Host" && key != "X-Session-Token" {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	// Use the session's Megaport token for authentication
	proxyReq.Header.Set("Authorization", "Bearer "+session.MegaportToken)

	// Set proper Content-Type
	if proxyReq.Header.Get("Content-Type") == "" && r.Method == "POST" {
		proxyReq.Header.Set("Content-Type", "application/json")
	}

	// Make the request
	resp, err := optimizedHTTPClient.Do(proxyReq)
	if err != nil {
		log.Printf("Proxy request failed: %v", err)
		http.Error(w, "Proxy request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Proxy response status: %d, body size: %d bytes", resp.StatusCode, len(respBody))

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Add CORS headers and re-apply security headers after upstream header copy
	// to ensure upstream cannot override our security policy
	setCORSHeaders(w, r, "Content-Type, Authorization, X-Session-Token")
	setSecurityHeaders(w)

	// Write response
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(respBody); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func addCorsHeaders(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set proper MIME types
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}

		// CORS headers
		setCORSHeaders(w, r, "Content-Type, Authorization, X-Session-Token")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Serve file
		fs.ServeHTTP(w, r)
	}
}
