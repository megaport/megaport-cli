//go:build !js && !wasm
// +build !js,!wasm

package server

import (
	"log"
	"net"
	"net/http"
	"time"
)

// Server represents the WASM HTTP server with session management
type Server struct {
	sessionManager *SessionManager
	httpClient     *http.Client
	logger         *log.Logger
}

// NewServer creates a new server instance
func NewServer(sessionDuration time.Duration, logger *log.Logger) *Server {
	return &Server{
		sessionManager: NewSessionManager(sessionDuration),
		httpClient:     createOptimizedHTTPClient(),
		logger:         logger,
	}
}

// createOptimizedHTTPClient creates an HTTP client with connection pooling
func createOptimizedHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       20,
			IdleConnTimeout:       90 * time.Second,
			DisableCompression:    false,
			DisableKeepAlives:     false,
			ForceAttemptHTTP2:     false,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   15 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
}

// GetSessionManager returns the session manager
func (s *Server) GetSessionManager() *SessionManager {
	return s.sessionManager
}
