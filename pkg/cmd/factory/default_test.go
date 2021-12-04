package factory

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
)

func Test_ioStreams_pager(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		config    config.Config
		wantPager string
	}{
		{
			name: "INSTILL_PAGER and PAGER set",
			env: map[string]string{
				"INSTILL_PAGER": "INSTILL_PAGER",
				"PAGER":         "PAGER",
			},
			wantPager: "INSTILL_PAGER",
		},
		{
			name: "INSTILL_PAGER and config pager set",
			env: map[string]string{
				"INSTILL_PAGER": "INSTILL_PAGER",
			},
			config:    pagerConfig(),
			wantPager: "INSTILL_PAGER",
		},
		{
			name: "config pager and PAGER set",
			env: map[string]string{
				"PAGER": "PAGER",
			},
			config:    pagerConfig(),
			wantPager: "CONFIG_PAGER",
		},
		{
			name: "only PAGER set",
			env: map[string]string{
				"PAGER": "PAGER",
			},
			wantPager: "PAGER",
		},
		{
			name: "INSTILL_PAGER set to blank string",
			env: map[string]string{
				"INSTILL_PAGER": "",
				"PAGER":         "PAGER",
			},
			wantPager: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					if k == "INSTILL_PAGER" {
						defer os.Unsetenv(k)
					} else {
						defer os.Setenv(k, old)
					}
				}
			}
			f := New("1")
			f.Config = func() (config.Config, error) {
				if tt.config == nil {
					return config.NewBlankConfig(), nil
				} else {
					return tt.config, nil
				}
			}
			io := ioStreams(f)
			assert.Equal(t, tt.wantPager, io.GetPager())
		})
	}
}

func Test_ioStreams_prompt(t *testing.T) {
	tests := []struct {
		name           string
		config         config.Config
		promptDisabled bool
	}{
		{
			name:           "default config",
			promptDisabled: false,
		},
		{
			name:           "config with prompt disabled",
			config:         disablePromptConfig(),
			promptDisabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New("1")
			f.Config = func() (config.Config, error) {
				if tt.config == nil {
					return config.NewBlankConfig(), nil
				} else {
					return tt.config, nil
				}
			}
			io := ioStreams(f)
			assert.Equal(t, tt.promptDisabled, io.GetNeverPrompt())
		})
	}
}

func Test_browserLauncher(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		config      config.Config
		wantBrowser string
	}{
		{
			name: "INSTILL_BROWSER set",
			env: map[string]string{
				"INSTILL_BROWSER": "INSTILL_BROWSER",
			},
			wantBrowser: "INSTILL_BROWSER",
		},
		{
			name:        "config browser set",
			config:      config.NewFromString("browser: CONFIG_BROWSER"),
			wantBrowser: "CONFIG_BROWSER",
		},
		{
			name: "BROWSER set",
			env: map[string]string{
				"BROWSER": "BROWSER",
			},
			wantBrowser: "BROWSER",
		},
		{
			name: "INSTILL_BROWSER and config browser set",
			env: map[string]string{
				"INSTILL_BROWSER": "INSTILL_BROWSER",
			},
			config:      config.NewFromString("browser: CONFIG_BROWSER"),
			wantBrowser: "INSTILL_BROWSER",
		},
		{
			name: "config browser and BROWSER set",
			env: map[string]string{
				"BROWSER": "BROWSER",
			},
			config:      config.NewFromString("browser: CONFIG_BROWSER"),
			wantBrowser: "CONFIG_BROWSER",
		},
		{
			name: "INSTILL_BROWSER and BROWSER set",
			env: map[string]string{
				"BROWSER":         "BROWSER",
				"INSTILL_BROWSER": "INSTILL_BROWSER",
			},
			wantBrowser: "INSTILL_BROWSER",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					defer os.Setenv(k, old)
				}
			}
			f := New("1")
			f.Config = func() (config.Config, error) {
				if tt.config == nil {
					return config.NewBlankConfig(), nil
				} else {
					return tt.config, nil
				}
			}
			browser := browserLauncher(f)
			assert.Equal(t, tt.wantBrowser, browser)
		})
	}
}

func pagerConfig() config.Config {
	return config.NewFromString("pager: CONFIG_PAGER")
}

func disablePromptConfig() config.Config {
	return config.NewFromString("prompt: disabled")
}
