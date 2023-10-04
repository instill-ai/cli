package cmdutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
)

func Test_CheckAuth(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), ".config", "inst")
	_ = os.MkdirAll(configDir, 0755)
	os.Setenv(config.INSTILL_CONFIG_DIR, configDir)
	defer os.Unsetenv(config.INSTILL_CONFIG_DIR)

	tests := []struct {
		name     string
		cfg      func(config.Config)
		expected bool
	}{
		{
			name:     "no instances",
			cfg:      func(c config.Config) {},
			expected: false,
		},
		{
			name: "no oauth2 hostname, no access token",
			cfg: func(c config.Config) {
				_ = c.Set("api.instill.tech", "access_token", "")
			},
			expected: true,
		},
		{
			name: "no oauth2 hostname, no access token",
			cfg: func(c config.Config) {
				_ = c.Set("instill.tech", "access_token", "")
				_ = c.Set("instill.tech", "oauth2_hostname", "auth.instill.tech")
			},
			expected: false,
		},
		{
			name: "oauth2 hostname, access token",
			cfg: func(c config.Config) {
				_ = c.Set("instill.tech", "access_token", "a token")
				_ = c.Set("instill.tech", "oauth2_hostname", "auth.instill.tech")
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewBlankConfig()
			tt.cfg(cfg)
			result := CheckAuth(cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}
