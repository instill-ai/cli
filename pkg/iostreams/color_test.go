package iostreams

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvColorDisabled(t *testing.T) {
	noColorBackup := os.Getenv("NO_COLOR")
	cliColorBackup := os.Getenv("CLICOLOR")
	cliColorForceBackup := os.Getenv("CLICOLOR_FORCE")
	t.Cleanup(func() {
		os.Setenv("NO_COLOR", noColorBackup)
		os.Setenv("CLICOLOR", cliColorBackup)
		os.Setenv("CLICOLOR_FORCE", cliColorForceBackup)
	})

	tests := []struct {
		name          string
		noColor       string
		cliColor      string
		cliColorForce string
		want          bool
	}{
		{
			name: "pristine env",
			want: false,
		},
		{
			name:    "NO_COLOR enabled",
			noColor: "1",
			want:    true,
		},
		{
			name:     "CLICOLOR disabled",
			cliColor: "0",
			want:     true,
		},
		{
			name:     "CLICOLOR enabled",
			cliColor: "1",
			want:     false,
		},
		{
			name:          "CLICOLOR_FORCE has no effect",
			cliColorForce: "1",
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)
			os.Setenv("CLICOLOR", tt.cliColor)
			os.Setenv("CLICOLOR_FORCE", tt.cliColorForce)

			if got := EnvColorDisabled(); got != tt.want {
				t.Errorf("EnvColorDisabled(): want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestEnvColorForced(t *testing.T) {
	noColorBackup := os.Getenv("NO_COLOR")
	cliColorBackup := os.Getenv("CLICOLOR")
	cliColorForceBackup := os.Getenv("CLICOLOR_FORCE")
	t.Cleanup(func() {
		os.Setenv("NO_COLOR", noColorBackup)
		os.Setenv("CLICOLOR", cliColorBackup)
		os.Setenv("CLICOLOR_FORCE", cliColorForceBackup)
	})

	tests := []struct {
		name          string
		noColor       string
		cliColor      string
		cliColorForce string
		want          bool
	}{
		{
			name: "pristine env",
			want: false,
		},
		{
			name:    "NO_COLOR enabled",
			noColor: "1",
			want:    false,
		},
		{
			name:     "CLICOLOR disabled",
			cliColor: "0",
			want:     false,
		},
		{
			name:     "CLICOLOR enabled",
			cliColor: "1",
			want:     false,
		},
		{
			name:          "CLICOLOR_FORCE enabled",
			cliColorForce: "1",
			want:          true,
		},
		{
			name:          "CLICOLOR_FORCE disabled",
			cliColorForce: "0",
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)
			os.Setenv("CLICOLOR", tt.cliColor)
			os.Setenv("CLICOLOR_FORCE", tt.cliColorForce)

			if got := EnvColorForced(); got != tt.want {
				t.Errorf("EnvColorForced(): want %v, got %v", tt.want, got)
			}
		})
	}
}

func Test_HextoRGB(t *testing.T) {
	tests := []struct {
		name  string
		hex   string
		text  string
		wants string
		cs    *ColorScheme
	}{
		{
			name:  "truecolor",
			hex:   "fc0303",
			text:  "red",
			wants: "\033[38;2;252;3;3mred\033[0m",
			cs:    NewColorScheme(true, true, true),
		},
		{
			name:  "no truecolor",
			hex:   "fc0303",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(true, true, false),
		},
		{
			name:  "no color",
			hex:   "fc0303",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(false, false, false),
		},
	}

	for _, tt := range tests {
		output := tt.cs.HexToRGB(tt.hex, tt.text)
		assert.Equal(t, tt.wants, output)
	}
}
