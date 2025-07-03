package local

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

// checkForUpdate checks whether this software has had a newer release on GitHub
func checkForUpdate(stateFilePath, repo, currentVersion string) (*releaseInfo, error) {

	stateEntry, _ := getStateEntry(stateFilePath)
	if stateEntry != nil && time.Since(stateEntry.CheckedForUpdateAt).Hours() < 0 {
		return nil, nil
	}

	releaseInfo, err := getLatestReleaseInfo(repo)
	if err != nil {
		return nil, err
	}

	err = setStateEntry(stateFilePath, time.Now(), *releaseInfo)
	if err != nil {
		return nil, err
	}

	if semver.Compare(
		strings.Replace(releaseInfo.Version, semver.Prerelease(releaseInfo.Version), "", 1),
		strings.Replace(currentVersion, semver.Prerelease(currentVersion), "", 1)) == 1 {
		return releaseInfo, nil
	}

	return nil, nil
}

// getLatestReleaseInfo fetches the latest release info from GitHub API
func getLatestReleaseInfo(repo string) (*releaseInfo, error) {
	return getLatestReleaseInfoWithClient(nil, repo)
}

// getLatestReleaseInfoWithClient fetches the latest release info using a custom HTTP client
func getLatestReleaseInfoWithClient(client *http.Client, repo string) (*releaseInfo, error) {
	// Use default client if none provided
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	// Make the API request
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent header (GitHub API requires this)
	req.Header.Set("User-Agent", "instill-cli")

	// Add GitHub token authentication if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the JSON array
	var releases []releaseInfo
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases JSON: %w", err)
	}

	// Check if we have any releases
	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found for repository %s", repo)
	}

	// Return the first (latest) release
	return &releases[0], nil
}

func getStateEntry(stateFilePath string) (*stateEntry, error) {
	content, err := os.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	var stateEntry stateEntry
	err = yaml.Unmarshal(content, &stateEntry)
	if err != nil {
		return nil, err
	}

	return &stateEntry, nil
}

func setStateEntry(stateFilePath string, t time.Time, r releaseInfo) error {
	data := stateEntry{CheckedForUpdateAt: t, LatestRelease: r}
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(stateFilePath), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(stateFilePath, content, 0600)
	return err
}
