//go:build !js && !wasm
// +build !js,!wasm

package server

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Session represents a customer's authenticated session
type Session struct {
	ID            string
	AccessKey     string
	SecretKey     string
	MegaportToken string
	TokenExpiry   time.Time
	CreatedAt     time.Time
	ExpiresAt     time.Time
	LastActivity  time.Time
}

// SessionManager manages customer sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	duration time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionDuration time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		duration: sessionDuration,
	}

	// Start cleanup goroutine
	go sm.cleanupExpiredSessions()

	return sm
}

// CreateSession creates a new session for a customer
func (sm *SessionManager) CreateSession(accessKey, secretKey string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:           sessionID,
		AccessKey:    accessKey,
		SecretKey:    secretKey,
		CreatedAt:    now,
		ExpiresAt:    now.Add(sm.duration),
		LastActivity: now,
	}

	sm.sessions[sessionID] = session
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return nil
	}

	return session
}

// UpdateActivity updates the last activity time for a session
func (sm *SessionManager) UpdateActivity(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
	}
}

// UpdateMegaportToken updates the Megaport API token for a session
func (sm *SessionManager) UpdateMegaportToken(sessionID, token string, expiry time.Time) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.MegaportToken = token
		session.TokenExpiry = expiry
	}
}

// DeleteSession removes a session (logout)
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sessionID)
}

// cleanupExpiredSessions periodically removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for id, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

// GetSessionCount returns the number of active sessions
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// generateSessionID generates a cryptographically secure random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
