package local

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/instill-ai/cli/pkg/httpmock"
	"golang.org/x/mod/semver"
)

func TestCheckForUpdate(t *testing.T) {
	scenarios := []struct {
		Name           string
		CurrentVersion string
		LatestVersion  string
		LatestURL      string
		ExpectsResult  bool
	}{
		{
			Name:           "latest is newer",
			CurrentVersion: "v0.0.1",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
	}

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			// Create HTTP mock registry
			mockRegistry := &httpmock.Registry{}

			// Register mock response for GitHub API
			mockRegistry.Register(
				httpmock.REST("GET", "repos/OWNER/REPO/releases"),
				httpmock.StringResponse(fmt.Sprintf(`[{
					"tag_name": "%s",
					"html_url": "%s",
					"published_at": "2023-01-01T00:00:00Z"
				}]`, s.LatestVersion, s.LatestURL)),
			)

			// Create HTTP client with mocked transport
			client := &http.Client{
				Transport: mockRegistry,
			}

			// Test the function with mocked client
			_, err := checkForUpdateWithClient(client, tempFilePath(), "OWNER/REPO", s.CurrentVersion)
			if err != nil {
				t.Fatal(err)
			}

			// Verify HTTP requests were made
			if len(mockRegistry.Requests) != 1 {
				t.Fatalf("expected 1 HTTP request, got %d", len(mockRegistry.Requests))
			}
			requestPath := mockRegistry.Requests[0].URL.Path
			if requestPath != "/repos/OWNER/REPO/releases" {
				t.Errorf("HTTP path: %q", requestPath)
			}
		})
	}
}

// Helper function for testing that accepts a custom HTTP client
func checkForUpdateWithClient(client *http.Client, stateFilePath, repo, currentVersion string) (*releaseInfo, error) {
	stateEntry, _ := getStateEntry(stateFilePath)
	if stateEntry != nil && time.Since(stateEntry.CheckedForUpdateAt).Hours() < 0 {
		return nil, nil
	}

	releaseInfo, err := getLatestReleaseInfoWithClient(client, repo)
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

func tempFilePath() string {
	file, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	os.Remove(file.Name())
	return file.Name()
}
