package version

import (
	"testing"
)

func TestFormat(t *testing.T) {
	expects := "instill version 1.4.0 (2020-12-15)\nhttps://github.com/instill-ai/cli/releases/tag/v1.4.0\n"
	if got := Format("1.4.0", "2020-12-15"); got != expects {
		t.Errorf("Format() = %q, wants %q", got, expects)
	}
}

func TestChangelogURL(t *testing.T) {
	tag := "0.3.2"
	url := "https://github.com/instill-ai/cli/releases/tag/v0.3.2"
	result := changelogURL(tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "v0.3.2"
	url = "https://github.com/instill-ai/cli/releases/tag/v0.3.2"
	result = changelogURL(tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "0.3.2-pre.1"
	url = "https://github.com/instill-ai/cli/releases/tag/v0.3.2-pre.1"
	result = changelogURL(tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "0.3.5-90-gdd3f0e0"
	url = "https://github.com/instill-ai/cli/releases/latest"
	result = changelogURL(tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "deadbeef"
	url = "https://github.com/instill-ai/cli/releases/latest"
	result = changelogURL(tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}
}
