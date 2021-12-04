package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
)

func Test_CheckAuth(t *testing.T) {
	tests := []struct {
		name     string
		cfg      func(config.Config)
		expected bool
	}{
		{
			name:     "no hosts",
			cfg:      func(c config.Config) {},
			expected: false,
		},
		{
			name: "host, no access token",
			cfg: func(c config.Config) {
				_ = c.Set("instill.tech", "access_token", "")
			},
			expected: false,
		},
		{
			name: "host, access token",
			cfg: func(c config.Config) {
				_ = c.Set("instill.tech", "access_token", "a token")
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
