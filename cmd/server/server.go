//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/server"
)

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
	webDir := flag.String("dir", "web", "Directory to serve files from")
	sessionDuration := flag.Duration("session-duration", 1*time.Hour, "Session duration")
	flag.Parse()

	// Create server with session management
	srv := server.NewServer(*sessionDuration, log.Default())

	// Authentication endpoints
	http.HandleFunc("/auth/login", srv.HandleLogin)
	http.HandleFunc("/auth/logout", srv.HandleLogout)
	http.HandleFunc("/auth/check", srv.HandleSessionCheck)

	// Authenticated API proxy
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		authenticatedProxyHandler(w, r, srv)
	})

	// Legacy proxy handler (backward compatible)
	http.HandleFunc("/proxy/", proxyHandler)

	// Static file server for everything else
	fs := http.FileServer(http.Dir(*webDir))
	http.Handle("/", addCorsHeaders(fs))

	log.Printf("Starting Megaport CLI WASM Server on http://localhost:%s", *port)
	log.Printf("Serving files from: %s", *webDir)
	log.Printf("Session duration: %v", *sessionDuration)
	log.Printf("\nEndpoints:")
	log.Printf("  - POST /auth/login    - Customer login")
	log.Printf("  - POST /auth/logout   - Customer logout")
	log.Printf("  - GET  /auth/check    - Check session validity")
	log.Printf("  - *    /api/*         - Authenticated API proxy")
	log.Printf("  - *    /proxy/*       - Legacy direct proxy")
	log.Fatal(http.ListenAndServe(":"+*port, nil))
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
		http.Error(w, "Failed to create proxy request: "+err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Proxy request failed: "+err.Error(), http.StatusBadGateway)
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

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")

	// Write response
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the target host from query parameter
	targetHost := r.URL.Query().Get("base")
	if targetHost == "" {
		http.Error(w, "Missing 'base' query parameter", http.StatusBadRequest)
		return
	}

	// Extract the path after /proxy/
	path := strings.TrimPrefix(r.URL.Path, "/proxy/")

	// Build the target URL with query parameters (excluding 'base')
	targetURL := "https://" + targetHost + "/" + path

	// Forward all query parameters except 'base'
	query := r.URL.Query()
	query.Del("base")
	if len(query) > 0 {
		targetURL += "?" + query.Encode()
	}

	log.Printf("Proxying %s request to: %s", r.Method, targetURL)

	// Create new request
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers from original request (except Host)
	for key, values := range r.Header {
		if key != "Host" {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	// Set proper Content-Type for OAuth requests
	if proxyReq.Header.Get("Content-Type") == "" && r.Method == "POST" {
		proxyReq.Header.Set("Content-Type", "application/json")
	}

	// Debug logging
	authHeader := proxyReq.Header.Get("Authorization")
	authPreview := authHeader
	if len(authHeader) > 20 {
		authPreview = authHeader[:20] + "..." + authHeader[len(authHeader)-8:]
	}
	log.Printf("Request headers: Content-Type=%s, Authorization=%s",
		proxyReq.Header.Get("Content-Type"),
		authPreview)

	// Log all headers for debugging
	log.Printf("All request headers:")
	for key, values := range proxyReq.Header {
		for _, value := range values {
			valuePreview := value
			if key == "Authorization" && len(value) > 20 {
				valuePreview = value[:20] + "..." + value[len(value)-8:]
			}
			log.Printf("  %s: %s", key, valuePreview)
		}
	}

	// Make the request using optimized client with connection pooling
	resp, err := optimizedHTTPClient.Do(proxyReq)
	if err != nil {
		log.Printf("Proxy request failed: %v", err)
		http.Error(w, "Proxy request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response body for debugging
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Proxy response status: %d, body: %s", resp.StatusCode, string(respBody))

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Write response body (already read for logging)
	_, err = w.Write(respBody)
	if err != nil {
		log.Printf("Error writing response body: %v", err)
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Serve file
		fs.ServeHTTP(w, r)
	}
}
