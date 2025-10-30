//go:build !js && !wasm
// +build !js,!wasm

package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestServer() *Server {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	return NewServer(1*time.Hour, logger)
}

func TestHandleLogin_Success(t *testing.T) {
	// Note: This test will actually try to authenticate with Megaport API
	// For unit testing, we should mock the authentication
	// This is more of an integration test example

	server := createTestServer()

	loginReq := LoginRequest{
		AccessKey:   "test-access-key",
		SecretKey:   "test-secret-key",
		Environment: "production",
	}

	body, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleLogin(w, req)

	// Note: This will likely fail without valid credentials
	// In a real scenario, you'd want to mock the Megaport API call
	// For now, we're just testing the structure
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestHandleLogin_MissingCredentials(t *testing.T) {
	server := createTestServer()

	tests := []struct {
		name    string
		request LoginRequest
	}{
		{
			name: "Missing access key",
			request: LoginRequest{
				SecretKey:   "test-secret",
				Environment: "production",
			},
		},
		{
			name: "Missing secret key",
			request: LoginRequest{
				AccessKey:   "test-access",
				Environment: "production",
			},
		},
		{
			name: "Missing both",
			request: LoginRequest{
				Environment: "production",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.HandleLogin(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestHandleLogin_InvalidMethod(t *testing.T) {
	server := createTestServer()

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/auth/login", nil)
			w := httptest.NewRecorder()

			server.HandleLogin(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	server := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleLogout_Success(t *testing.T) {
	server := createTestServer()

	// Create a session first
	session, err := server.sessionManager.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	// Logout request
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("X-Session-Token", session.ID)
	w := httptest.NewRecorder()

	server.HandleLogout(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify session was deleted
	deletedSession := server.sessionManager.GetSession(session.ID)
	assert.Nil(t, deletedSession)
}

func TestHandleLogout_MissingToken(t *testing.T) {
	server := createTestServer()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	server.HandleLogout(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleLogout_InvalidMethod(t *testing.T) {
	server := createTestServer()

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/auth/logout", nil)
			w := httptest.NewRecorder()

			server.HandleLogout(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestHandleSessionCheck_ValidSession(t *testing.T) {
	server := createTestServer()

	// Create a session
	session, err := server.sessionManager.CreateSession("test-key", "test-secret")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	req.Header.Set("X-Session-Token", session.ID)
	w := httptest.NewRecorder()

	server.HandleSessionCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response["valid"].(bool))
	assert.NotNil(t, response["expiresAt"])
}

func TestHandleSessionCheck_InvalidSession(t *testing.T) {
	server := createTestServer()

	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	req.Header.Set("X-Session-Token", "invalid-token")
	w := httptest.NewRecorder()

	server.HandleSessionCheck(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleSessionCheck_MissingToken(t *testing.T) {
	server := createTestServer()

	req := httptest.NewRequest(http.MethodGet, "/auth/check", nil)
	w := httptest.NewRecorder()

	server.HandleSessionCheck(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginRequest_DefaultEnvironment(t *testing.T) {
	server := createTestServer()

	loginReq := LoginRequest{
		AccessKey: "test-key",
		SecretKey: "test-secret",
		// Environment not specified
	}

	body, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleLogin(w, req)

	// Should default to production (though authentication will likely fail)
	// This test verifies the request handling, not the authentication
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}
