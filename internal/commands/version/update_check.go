package version

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var githubLatestURL = "https://api.github.com/repos/megaport/megaport-cli/releases/latest"

const updateCheckTTL = 24 * time.Hour

type updateCheckCache struct {
	LastChecked   time.Time `json:"last_checked"`
	LatestVersion string    `json:"latest_version"`
}

// checkForUpdate returns the latest version and whether it differs from currentVersion.
// Fails silently — never returns an error. Returns ("", false) for "dev" builds.
func checkForUpdate(currentVersion string, client *http.Client, cacheDir string) (string, bool) {
	if currentVersion == "dev" || currentVersion == "" {
		return "", false
	}
	cacheFile := filepath.Join(cacheDir, "update-check.json")
	latest := readUpdateCache(cacheFile)
	if latest == "" {
		latest = fetchLatestVersion(client)
		if latest == "" {
			return "", false
		}
		writeUpdateCache(cacheFile, latest)
	}
	if latest != currentVersion {
		return latest, true
	}
	return "", false
}

func readUpdateCache(cacheFile string) string {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}
	var cache updateCheckCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return ""
	}
	if time.Since(cache.LastChecked) > updateCheckTTL {
		return ""
	}
	return cache.LatestVersion
}

func writeUpdateCache(cacheFile, ver string) {
	cache := updateCheckCache{LastChecked: time.Now(), LatestVersion: ver}
	data, _ := json.Marshal(cache)
	_ = os.WriteFile(cacheFile, data, 0600)
}

func fetchLatestVersion(client *http.Client) string {
	resp, err := client.Get(githubLatestURL) //nolint:noctx
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var result struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	return result.TagName
}
