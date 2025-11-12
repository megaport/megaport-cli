//go:build !js && !wasm
// +build !js,!wasm

package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionManager_CreateSession(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	accessKey := "test-access-key"
	secretKey := "test-secret-key"

	session, err := sm.CreateSession(accessKey, secretKey)
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, accessKey, session.AccessKey)
	assert.Equal(t, secretKey, session.SecretKey)
	assert.True(t, session.ExpiresAt.After(time.Now()))
	assert.False(t, session.CreatedAt.IsZero())
	assert.False(t, session.LastActivity.IsZero())
}

func TestSessionManager_GetSession(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	// Create a session
	session, err := sm.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	// Get the session
	retrieved := sm.GetSession(session.ID)
	require.NotNil(t, retrieved)
	assert.Equal(t, session.ID, retrieved.ID)
	assert.Equal(t, session.AccessKey, retrieved.AccessKey)

	// Get non-existent session
	nonExistent := sm.GetSession("non-existent-id")
	assert.Nil(t, nonExistent)
}

func TestSessionManager_DeleteSession(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	// Create a session
	session, err := sm.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	// Verify it exists
	retrieved := sm.GetSession(session.ID)
	assert.NotNil(t, retrieved)

	// Delete it
	sm.DeleteSession(session.ID)

	// Verify it's gone
	deleted := sm.GetSession(session.ID)
	assert.Nil(t, deleted)
}

func TestSessionManager_UpdateActivity(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	// Create a session
	session, err := sm.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	// Wait a moment
	time.Sleep(10 * time.Millisecond)
	originalActivity := session.LastActivity

	// Update activity
	sm.UpdateActivity(session.ID)

	// Get updated session
	updated := sm.GetSession(session.ID)
	require.NotNil(t, updated)
	assert.True(t, updated.LastActivity.After(originalActivity))
}

func TestSessionManager_CleanupExpiredSessions(t *testing.T) {
	// Create manager with short duration
	sm := NewSessionManager(100 * time.Millisecond)

	// Create a session
	session, err := sm.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	// Verify it exists
	retrieved := sm.GetSession(session.ID)
	assert.NotNil(t, retrieved)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Manual cleanup by deleting expired sessions
	sm.mu.Lock()
	now := time.Now()
	for id, s := range sm.sessions {
		if now.After(s.ExpiresAt) {
			delete(sm.sessions, id)
		}
	}
	sm.mu.Unlock()

	// Verify it's been cleaned up
	expired := sm.GetSession(session.ID)
	assert.Nil(t, expired)
}

func TestSessionManager_UpdateMegaportToken(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	// Create a session
	session, err := sm.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	// Update token
	token := "test-megaport-token"
	expiry := time.Now().Add(1 * time.Hour)
	sm.UpdateMegaportToken(session.ID, token, expiry)

	// Verify token was updated
	updated := sm.GetSession(session.ID)
	require.NotNil(t, updated)
	assert.Equal(t, token, updated.MegaportToken)
	assert.Equal(t, expiry.Unix(), updated.TokenExpiry.Unix())
}

func TestSessionManager_MegaportToken(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	// Create a session
	session, err := sm.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	// No token initially
	retrieved := sm.GetSession(session.ID)
	assert.Empty(t, retrieved.MegaportToken)

	// Add a token
	expectedToken := "test-megaport-token"
	expiry := time.Now().Add(1 * time.Hour)
	sm.UpdateMegaportToken(session.ID, expectedToken, expiry)

	// Get the session and verify token
	updated := sm.GetSession(session.ID)
	require.NotNil(t, updated)
	assert.Equal(t, expectedToken, updated.MegaportToken)
	assert.Equal(t, expiry.Unix(), updated.TokenExpiry.Unix())
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	// Create multiple sessions concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			session, err := sm.CreateSession("key", "secret")
			assert.NoError(t, err)
			assert.NotNil(t, session)

			// Read session
			retrieved := sm.GetSession(session.ID)
			assert.NotNil(t, retrieved)

			// Update activity
			sm.UpdateActivity(session.ID)

			// Delete session
			sm.DeleteSession(session.ID)

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSession_Expiration(t *testing.T) {
	sm := NewSessionManager(1 * time.Hour)

	tests := []struct {
		name     string
		expiry   time.Time
		expected bool
	}{
		{
			name:     "Not expired",
			expiry:   time.Now().Add(1 * time.Hour),
			expected: true, // session should exist
		},
		{
			name:     "Expired",
			expiry:   time.Now().Add(-1 * time.Hour),
			expected: false, // session should be nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := sm.CreateSession("test-key", "test-secret")
			require.NoError(t, err)

			// Manually set expiry for testing
			sm.mu.Lock()
			session.ExpiresAt = tt.expiry
			sm.mu.Unlock()

			// Try to get the session
			retrieved := sm.GetSession(session.ID)
			if tt.expected {
				assert.NotNil(t, retrieved)
			} else {
				assert.Nil(t, retrieved)
			}
		})
	}
}
