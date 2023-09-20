package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Hosts(t *testing.T) {
	c := NewBlankConfig()
	hosts, err := c.Hosts()
	require.NoError(t, err)
	assert.Equal(t, []string{}, hosts)
}

func Test_HostsTyped(t *testing.T) {
	c := NewBlankConfig()
	hosts, err := c.HostsTyped()
	require.NoError(t, err)
	assert.Equal(t, []HostConfigTyped(nil), hosts)
}

func Test_fileConfig_Typed(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), ".config", "instill")
	_ = os.MkdirAll(configDir, 0755)
	os.Setenv(INSTILL_CONFIG_DIR, configDir)
	defer os.Unsetenv(INSTILL_CONFIG_DIR)

	fc := NewBlankConfig()
	hc := fc.MakeConfigForHost("foo")
	_ = hc.SetStringValue("oauth2_hostname", "bar1")
	_ = hc.SetStringValue("oauth2_secret", "bar2")
	_ = hc.SetStringValue("api_version", "bar3")
	hct, err := hostConfigToTyped(hc)
	assert.NoError(t, err)

	assert.Equal(t, "bar1", hct.Oauth2Hostname)
	assert.Equal(t, "bar2", hct.Oauth2Secret)
	assert.Equal(t, "bar3", hct.APIVersion)

	hct.APIVersion = "bar4"

	err = fc.SaveTyped(hct)
	require.NoError(t, err)
}
