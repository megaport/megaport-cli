package version

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func githubHandler(tagName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": tagName})
	}
}

func TestCheckForUpdate_UpdateAvailable(t *testing.T) {
	srv := httptest.NewServer(githubHandler("v2.0.0"))
	defer srv.Close()

	orig := githubLatestURL
	githubLatestURL = srv.URL
	defer func() { githubLatestURL = orig }()

	dir := t.TempDir()
	latest, hasUpdate := checkForUpdate("v1.0.0", srv.Client(), dir)
	assert.True(t, hasUpdate)
	assert.Equal(t, "v2.0.0", latest)
}

func TestCheckForUpdate_AlreadyCurrent(t *testing.T) {
	srv := httptest.NewServer(githubHandler("v1.0.0"))
	defer srv.Close()

	orig := githubLatestURL
	githubLatestURL = srv.URL
	defer func() { githubLatestURL = orig }()

	dir := t.TempDir()
	latest, hasUpdate := checkForUpdate("v1.0.0", srv.Client(), dir)
	assert.False(t, hasUpdate)
	assert.Empty(t, latest)
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	// Neither "dev" nor "" should trigger a network call.
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
	}))
	defer srv.Close()

	orig := githubLatestURL
	githubLatestURL = srv.URL
	defer func() { githubLatestURL = orig }()

	dir := t.TempDir()
	_, has1 := checkForUpdate("dev", srv.Client(), dir)
	_, has2 := checkForUpdate("", srv.Client(), dir)
	assert.False(t, has1)
	assert.False(t, has2)
	assert.Equal(t, 0, callCount, "should not hit network for dev/empty version")
}

func TestCheckForUpdate_NetworkError(t *testing.T) {
	orig := githubLatestURL
	githubLatestURL = "http://127.0.0.1:0/nowhere"
	defer func() { githubLatestURL = orig }()

	dir := t.TempDir()
	latest, hasUpdate := checkForUpdate("v1.0.0", &http.Client{Timeout: time.Second}, dir)
	assert.False(t, hasUpdate)
	assert.Empty(t, latest)
}

func TestCheckForUpdate_CacheHit(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v9.9.9"})
	}))
	defer srv.Close()

	orig := githubLatestURL
	githubLatestURL = srv.URL
	defer func() { githubLatestURL = orig }()

	dir := t.TempDir()

	// Write a fresh cache entry manually.
	cache := updateCheckCache{
		LastChecked:   time.Now(),
		LatestVersion: "v2.0.0",
	}
	data, err := json.Marshal(cache)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "update-check.json"), data, 0600))

	latest, hasUpdate := checkForUpdate("v1.0.0", srv.Client(), dir)
	assert.True(t, hasUpdate)
	assert.Equal(t, "v2.0.0", latest)
	assert.Equal(t, 0, callCount, "should use cache, not hit network")
}

func TestCheckForUpdate_CacheExpired(t *testing.T) {
	srv := httptest.NewServer(githubHandler("v3.0.0"))
	defer srv.Close()

	orig := githubLatestURL
	githubLatestURL = srv.URL
	defer func() { githubLatestURL = orig }()

	dir := t.TempDir()

	// Write a stale cache entry (25 hours old).
	cache := updateCheckCache{
		LastChecked:   time.Now().Add(-25 * time.Hour),
		LatestVersion: "v2.0.0",
	}
	data, err := json.Marshal(cache)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "update-check.json"), data, 0600))

	latest, hasUpdate := checkForUpdate("v1.0.0", srv.Client(), dir)
	assert.True(t, hasUpdate)
	assert.Equal(t, "v3.0.0", latest, "should fetch fresh version, ignoring stale cache")
}

func TestFetchLatestVersion_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	orig := githubLatestURL
	githubLatestURL = srv.URL
	defer func() { githubLatestURL = orig }()

	result := fetchLatestVersion(srv.Client())
	assert.Empty(t, result)
}

func TestFetchLatestVersion_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	orig := githubLatestURL
	githubLatestURL = srv.URL
	defer func() { githubLatestURL = orig }()

	result := fetchLatestVersion(srv.Client())
	assert.Empty(t, result)
}

func TestReadUpdateCache_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "update-check.json")
	require.NoError(t, os.WriteFile(cacheFile, []byte("not-valid-json"), 0600))

	result := readUpdateCache(cacheFile)
	assert.Empty(t, result)
}

func TestVersionCommand_ShowsUpdateMessage(t *testing.T) {
	srv := httptest.NewServer(githubHandler("v99.0.0"))
	defer srv.Close()

	origURL := githubLatestURL
	origVersion := version
	origExec := execCommand
	defer func() {
		githubLatestURL = origURL
		version = origVersion
		execCommand = origExec
	}()

	githubLatestURL = srv.URL
	version = "v1.0.0"
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	t.Setenv("MEGAPORT_CONFIG_DIR", t.TempDir())

	rootCmd := &cobra.Command{Use: "test-cli"}
	AddCommandsTo(rootCmd)

	versionCmd, _, err := rootCmd.Find([]string{"version"})
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)

	require.NoError(t, versionCmd.RunE(versionCmd, []string{}))
	assert.Contains(t, buf.String(), "Megaport CLI Version: v1.0.0")
}
