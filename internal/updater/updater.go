package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	githubAPI     = "https://api.github.com/repos/Eljakani/ward/releases/latest"
	checkInterval = 24 * time.Hour
	httpTimeout   = 2 * time.Second
	cacheFile     = "last_update_check"
)

// githubRelease is the subset of the GitHub release API we need.
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// updateCache stores the last check result so we don't hit GitHub every run.
type updateCache struct {
	CheckedAt     int64  `json:"checked_at"`
	LatestVersion string `json:"latest_version"`
}

// CheckForUpdate queries GitHub for the latest release and returns an update
// notice if a newer version is available. Returns "" if up to date or on error.
// Results are cached to wardDir for checkInterval to avoid rate limits.
func CheckForUpdate(currentVersion string, wardDir string) string {
	// Skip if running a dev build
	if currentVersion == "dev" || currentVersion == "" {
		return ""
	}

	// Check cache first
	cachePath := filepath.Join(wardDir, cacheFile)
	if notice := checkCache(cachePath, currentVersion); notice != "" {
		return notice
	}

	// Fetch from GitHub
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(githubAPI)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	// Save cache
	saveCache(cachePath, release.TagName)

	// Compare
	if isNewer(release.TagName, currentVersion) {
		return formatNotice(currentVersion, release.TagName, release.HTMLURL)
	}

	return ""
}

func checkCache(path string, currentVersion string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var cache updateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return ""
	}

	// Cache expired?
	if time.Since(time.Unix(cache.CheckedAt, 0)) > checkInterval {
		return ""
	}

	// Cache valid — check version
	if isNewer(cache.LatestVersion, currentVersion) {
		return formatNotice(currentVersion, cache.LatestVersion, "https://github.com/Eljakani/ward/releases/latest")
	}

	// Return a special sentinel to indicate "checked, up to date"
	return ""
}

func saveCache(path string, latestVersion string) {
	cache := updateCache{
		CheckedAt:     time.Now().Unix(),
		LatestVersion: latestVersion,
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, data, 0644)
}

// isNewer returns true if latest > current using simple semver comparison.
func isNewer(latest, current string) bool {
	latestParts := parseSemver(latest)
	currentParts := parseSemver(current)
	if latestParts == nil || currentParts == nil {
		return false
	}

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// parseSemver extracts [major, minor, patch] from a version string like "v1.2.3" or "1.2.3".
func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil
	}

	result := make([]int, 3)
	for i, p := range parts {
		// Strip any pre-release suffix (e.g., "3-beta")
		if idx := strings.IndexAny(p, "-+"); idx >= 0 {
			p = p[:idx]
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		result[i] = n
	}
	return result
}

func formatNotice(current, latest, url string) string {
	return fmt.Sprintf(
		"A new version of Ward is available: %s → %s\nUpdate: go install github.com/eljakani/ward@%s\nRelease: %s",
		current, latest, latest, url,
	)
}
