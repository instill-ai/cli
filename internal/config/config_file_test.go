package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_parseConfig(t *testing.T) {
	defer stubConfig(`---
hosts:
  instill.tech:
    access_token: abcde
    refresh_token: 12345
    id_token: abc123
`, "")()
	config, err := parseConfig("config.yml")
	assert.NoError(t, err)
	accessToken, err := config.Get("instill.tech", "access_token")
	assert.NoError(t, err)
	assert.Equal(t, "abcde", accessToken)
	refreshToken, err := config.Get("instill.tech", "refresh_token")
	assert.NoError(t, err)
	assert.Equal(t, "12345", refreshToken)
	idToken, err := config.Get("instill.tech", "id_token")
	assert.NoError(t, err)
	assert.Equal(t, "abc123", idToken)
}

func Test_parseConfig_multipleHosts(t *testing.T) {
	defer stubConfig(`---
hosts:
  example.com:
    access_token: abcde
    refresh_token: 12345
    id_token: abc123
  instill.tech:
    access_token: abcde
    refresh_token: 12345
    id_token: abc123
`, "")()
	config, err := parseConfig("config.yml")
	assert.NoError(t, err)
	accessToken, err := config.Get("instill.tech", "access_token")
	assert.NoError(t, err)
	assert.Equal(t, "abcde", accessToken)
	refreshToken, err := config.Get("example.com", "refresh_token")
	assert.NoError(t, err)
	assert.Equal(t, "12345", refreshToken)
}

func Test_parseConfig_hostsFile(t *testing.T) {
	defer stubConfig("", `---
instill.tech:
  access_token: abcde
  refresh_token: 12345
`)()
	config, err := parseConfig("config.yml")
	assert.NoError(t, err)
	accessToken, err := config.Get("instill.tech", "access_token")
	assert.NoError(t, err)
	assert.Equal(t, "abcde", accessToken)
	refreshToken, err := config.Get("instill.tech", "refresh_token")
	assert.NoError(t, err)
	assert.Equal(t, "12345", refreshToken)
}

func Test_parseConfig_hostFallback(t *testing.T) {
	defer stubConfig(`---
protocol: grpc
`, `---
instill.tech:
  access_token: abcde
  refresh_token: 12345
example.com:
  access_token: abcde
  refresh_token: 12345
  protocol: https
`)()
	config, err := parseConfig("config.yml")
	assert.NoError(t, err)
	val, err := config.Get("example.com", "protocol")
	assert.NoError(t, err)
	assert.Equal(t, "https", val)
	val, err = config.Get("instill.tech", "protocol")
	assert.NoError(t, err)
	assert.Equal(t, "grpc", val)
	val, err = config.Get("nonexistent.io", "protocol")
	assert.NoError(t, err)
	assert.Equal(t, "grpc", val)
}

func Test_parseConfig_migrateConfig(t *testing.T) {
	defer stubConfig(`---
instill.tech:
  - access_token: abcde
    refresh_token: 12345
`, "")()

	mainBuf := bytes.Buffer{}
	hostsBuf := bytes.Buffer{}
	defer StubWriteConfig(&mainBuf, &hostsBuf)()
	defer StubBackupConfig()()

	_, err := parseConfig("config.yml")
	assert.NoError(t, err)

	expectedHosts := `instill.tech:
    access_token: abcde
    refresh_token: "12345"
`

	assert.Equal(t, expectedHosts, hostsBuf.String())
	assert.NotContains(t, mainBuf.String(), "instill.tech")
	assert.NotContains(t, mainBuf.String(), "access_token")
}

func Test_parseConfigFile(t *testing.T) {
	tests := []struct {
		contents string
		wantsErr bool
	}{
		{
			contents: "",
			wantsErr: true,
		},
		{
			contents: " ",
			wantsErr: false,
		},
		{
			contents: "\n",
			wantsErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("contents: %q", tt.contents), func(t *testing.T) {
			defer stubConfig(tt.contents, "")()
			_, yamlRoot, err := parseConfigFile("config.yml")
			if tt.wantsErr != (err != nil) {
				t.Fatalf("got error: %v", err)
			}
			if tt.wantsErr {
				return
			}
			assert.Equal(t, yaml.MappingNode, yamlRoot.Content[0].Kind)
			assert.Equal(t, 0, len(yamlRoot.Content[0].Content))
		})
	}
}

func Test_ConfigDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		onlyWindows bool
		env         map[string]string
		output      string
	}{
		{
			name: "HOME/USERPROFILE specified",
			env: map[string]string{
				"INSTILL_CONFIG_DIR": "",
				"XDG_CONFIG_HOME":    "",
				"AppData":            "",
				"USERPROFILE":        tempDir,
				"HOME":               tempDir,
			},
			output: filepath.Join(tempDir, ".config", "instill"),
		},
		{
			name: "INSTILL_CONFIG_DIR specified",
			env: map[string]string{
				"INSTILL_CONFIG_DIR": filepath.Join(tempDir, "instill_config_dir"),
			},
			output: filepath.Join(tempDir, "instill_config_dir"),
		},
		{
			name: "XDG_CONFIG_HOME specified",
			env: map[string]string{
				"XDG_CONFIG_HOME": tempDir,
			},
			output: filepath.Join(tempDir, "instill"),
		},
		{
			name: "INSTILL_CONFIG_DIR and XDG_CONFIG_HOME specified",
			env: map[string]string{
				"INSTILL_CONFIG_DIR": filepath.Join(tempDir, "instill_config_dir"),
				"XDG_CONFIG_HOME":    tempDir,
			},
			output: filepath.Join(tempDir, "instill_config_dir"),
		},
		{
			name:        "AppData specified",
			onlyWindows: true,
			env: map[string]string{
				"AppData": tempDir,
			},
			output: filepath.Join(tempDir, "Instill CLI"),
		},
		{
			name:        "INSTILL_CONFIG_DIR and AppData specified",
			onlyWindows: true,
			env: map[string]string{
				"INSTILL_CONFIG_DIR": filepath.Join(tempDir, "instill_config_dir"),
				"AppData":            tempDir,
			},
			output: filepath.Join(tempDir, "instill_config_dir"),
		},
		{
			name:        "XDG_CONFIG_HOME and AppData specified",
			onlyWindows: true,
			env: map[string]string{
				"XDG_CONFIG_HOME": tempDir,
				"AppData":         tempDir,
			},
			output: filepath.Join(tempDir, "instill"),
		},
	}

	for _, tt := range tests {
		if tt.onlyWindows && runtime.GOOS != "windows" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					defer os.Setenv(k, old)
				}
			}

			// Create directory to skip auto migration code
			// which gets run when target directory does not exist
			_ = os.MkdirAll(tt.output, 0755)

			assert.Equal(t, tt.output, ConfigDir())
		})
	}
}

func Test_configFile_Write_toDisk(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), ".config", "instill")
	_ = os.MkdirAll(configDir, 0755)
	os.Setenv(INSTILL_CONFIG_DIR, configDir)
	defer os.Unsetenv(INSTILL_CONFIG_DIR)

	cfg := NewFromString(`pager: less`)
	err := cfg.Write()
	if err != nil {
		t.Fatal(err)
	}

	expectedConfig := "pager: less\n"
	if configBytes, err := os.ReadFile(filepath.Join(configDir, "config.yml")); err != nil {
		t.Error(err)
	} else if string(configBytes) != expectedConfig {
		t.Errorf("expected config.yml %q, got %q", expectedConfig, string(configBytes))
	}

	if configBytes, err := os.ReadFile(filepath.Join(configDir, "hosts.yml")); err != nil {
		t.Error(err)
	} else if string(configBytes) != "" {
		t.Errorf("unexpected hosts.yml: %q", string(configBytes))
	}
}

func Test_autoMigrateConfigDir_noMigration_notExist(t *testing.T) {
	homeDir := t.TempDir()
	migrateDir := t.TempDir()

	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}
	old := os.Getenv(homeEnvVar)
	os.Setenv(homeEnvVar, homeDir)
	defer os.Setenv(homeEnvVar, old)

	err := autoMigrateConfigDir(migrateDir)
	assert.Equal(t, errNotExist, err)

	files, err := os.ReadDir(migrateDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))
}

func Test_autoMigrateConfigDir_noMigration_samePath(t *testing.T) {
	homeDir := t.TempDir()
	migrateDir := filepath.Join(homeDir, ".config", "instill")
	err := os.MkdirAll(migrateDir, 0755)
	assert.NoError(t, err)

	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}
	old := os.Getenv(homeEnvVar)
	os.Setenv(homeEnvVar, homeDir)
	defer os.Setenv(homeEnvVar, old)

	err = autoMigrateConfigDir(migrateDir)
	assert.Equal(t, errSamePath, err)

	files, err := os.ReadDir(migrateDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))
}

func Test_autoMigrateConfigDir_migration(t *testing.T) {
	homeDir := t.TempDir()
	migrateDir := t.TempDir()
	homeConfigDir := filepath.Join(homeDir, ".config", "instill")
	migrateConfigDir := filepath.Join(migrateDir, ".config", "instill")

	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}
	old := os.Getenv(homeEnvVar)
	os.Setenv(homeEnvVar, homeDir)
	defer os.Setenv(homeEnvVar, old)

	err := os.MkdirAll(homeConfigDir, 0755)
	assert.NoError(t, err)
	f, err := os.CreateTemp(homeConfigDir, "")
	assert.NoError(t, err)
	f.Close()

	err = autoMigrateConfigDir(migrateConfigDir)
	assert.NoError(t, err)

	_, err = os.ReadDir(homeConfigDir)
	assert.True(t, os.IsNotExist(err))

	files, err := os.ReadDir(migrateConfigDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
}

func Test_StateDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		onlyWindows bool
		env         map[string]string
		output      string
	}{
		{
			name: "HOME/USERPROFILE specified",
			env: map[string]string{
				"XDG_STATE_HOME":     "",
				"INSTILL_CONFIG_DIR": "",
				"XDG_CONFIG_HOME":    "",
				"LocalAppData":       "",
				"USERPROFILE":        tempDir,
				"HOME":               tempDir,
			},
			output: filepath.Join(tempDir, ".local", "instill", "state"),
		},
		{
			name: "XDG_STATE_HOME specified",
			env: map[string]string{
				"XDG_STATE_HOME": tempDir,
			},
			output: filepath.Join(tempDir, "instill"),
		},
		{
			name:        "LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"LocalAppData": tempDir,
			},
			output: filepath.Join(tempDir, "Instill CLI"),
		},
		{
			name:        "XDG_STATE_HOME and LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"XDG_STATE_HOME": tempDir,
				"LocalAppData":   tempDir,
			},
			output: filepath.Join(tempDir, "instill"),
		},
	}

	for _, tt := range tests {
		if tt.onlyWindows && runtime.GOOS != "windows" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					defer os.Setenv(k, old)
				}
			}

			// Create directory to skip auto migration code
			// which gets run when target directory does not exist
			_ = os.MkdirAll(tt.output, 0755)

			assert.Equal(t, tt.output, StateDir())
		})
	}
}

func Test_autoMigrateStateDir_noMigration_notExist(t *testing.T) {
	homeDir := t.TempDir()
	migrateDir := t.TempDir()

	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}
	old := os.Getenv(homeEnvVar)
	os.Setenv(homeEnvVar, homeDir)
	defer os.Setenv(homeEnvVar, old)

	err := autoMigrateStateDir(migrateDir)
	assert.Equal(t, errNotExist, err)

	files, err := os.ReadDir(migrateDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))
}

func Test_autoMigrateStateDir_noMigration_samePath(t *testing.T) {
	homeDir := t.TempDir()
	migrateDir := filepath.Join(homeDir, ".config", "instill")
	err := os.MkdirAll(migrateDir, 0755)
	assert.NoError(t, err)

	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}
	old := os.Getenv(homeEnvVar)
	os.Setenv(homeEnvVar, homeDir)
	defer os.Setenv(homeEnvVar, old)

	err = autoMigrateStateDir(migrateDir)
	assert.Equal(t, errSamePath, err)

	files, err := os.ReadDir(migrateDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))
}

func Test_autoMigrateStateDir_migration(t *testing.T) {
	homeDir := t.TempDir()
	migrateDir := t.TempDir()
	homeConfigDir := filepath.Join(homeDir, ".config", "instill")
	migrateStateDir := filepath.Join(migrateDir, ".local", "instill")

	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}
	old := os.Getenv(homeEnvVar)
	os.Setenv(homeEnvVar, homeDir)
	defer os.Setenv(homeEnvVar, old)

	err := os.MkdirAll(homeConfigDir, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(homeConfigDir, "state.yml"), nil, 0755)
	assert.NoError(t, err)

	err = autoMigrateStateDir(migrateStateDir)
	assert.NoError(t, err)

	files, err := os.ReadDir(homeConfigDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))

	files, err = os.ReadDir(migrateStateDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "state.yml", files[0].Name())
}

func Test_DataDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		onlyWindows bool
		env         map[string]string
		output      string
	}{
		{
			name: "HOME/USERPROFILE specified",
			env: map[string]string{
				"XDG_DATA_HOME":      "",
				"INSTILL_CONFIG_DIR": "",
				"XDG_CONFIG_HOME":    "",
				"LocalAppData":       "",
				"USERPROFILE":        tempDir,
				"HOME":               tempDir,
			},
			output: filepath.Join(tempDir, ".local", "share", "instill"),
		},
		{
			name: "XDG_DATA_HOME specified",
			env: map[string]string{
				"XDG_DATA_HOME": tempDir,
			},
			output: filepath.Join(tempDir, "instill"),
		},
		{
			name:        "LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"LocalAppData": tempDir,
			},
			output: filepath.Join(tempDir, "Instill CLI"),
		},
		{
			name:        "XDG_DATA_HOME and LocalAppData specified",
			onlyWindows: true,
			env: map[string]string{
				"XDG_DATA_HOME": tempDir,
				"LocalAppData":  tempDir,
			},
			output: filepath.Join(tempDir, "instill"),
		},
	}

	for _, tt := range tests {
		if tt.onlyWindows && runtime.GOOS != "windows" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					old := os.Getenv(k)
					os.Setenv(k, v)
					defer os.Setenv(k, old)
				}
			}

			assert.Equal(t, tt.output, DataDir())
		})
	}
}
