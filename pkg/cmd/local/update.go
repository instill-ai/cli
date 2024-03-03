package local

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

// checkForUpdate checks whether this software has had a newer release on GitHub
func checkForUpdate(execDep ExecDep, stateFilePath, repo, currentVersion string) (*releaseInfo, error) {

	stateEntry, _ := getStateEntry(stateFilePath)
	if stateEntry != nil && time.Since(stateEntry.CheckedForUpdateAt).Hours() < 0 {
		return nil, nil
	}

	releaseInfo, err := getLatestPreReleaseInfo(execDep, repo)
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

func getLatestPreReleaseInfo(execDep ExecDep, repo string) (*releaseInfo, error) {
	var latestPreRelease releaseInfo
	if output, err := execCmd(execDep, "bash", "-c", "curl -s https://api.github.com/repos/%s/releases | jq -r 'map(select(.prerelease)) | first'", repo); err == nil {
		if len(output) > 0 && output[0] == '{' {
			if err := json.Unmarshal([]byte(output), &latestPreRelease); err != nil {
				return nil, err
			}
		}
	}
	return &latestPreRelease, nil
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
