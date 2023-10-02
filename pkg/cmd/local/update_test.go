package local

import (
	"testing"
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

			_, err := checkForUpdate(&ExecMock{}, "OWNER/REPO", s.CurrentVersion)
			if err != nil {
				t.Fatal(err)
			}

		})
	}
}
