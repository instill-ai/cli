package config

import (
	"bytes"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
)

func Test_fileConfig_Set(t *testing.T) {
	mainBuf := bytes.Buffer{}
	hostsBuf := bytes.Buffer{}
	defer StubWriteConfig(&mainBuf, &hostsBuf)()

	c := NewBlankConfig()
	assert.NoError(t, c.Set("", "editor", "nano"))
	assert.NoError(t, c.Set("github.com", "protocol", "ssh"))
	assert.NoError(t, c.Set("example.com", "editor", "vim"))
	assert.NoError(t, c.Set("github.com", "user", "hubot"))
	assert.NoError(t, c.Write())

	assert.Contains(t, mainBuf.String(), "editor: nano")
	assert.Contains(t, mainBuf.String(), "protocol: https")
	assert.Equal(t, `github.com:
    protocol: ssh
    user: hubot
example.com:
    editor: vim
`, hostsBuf.String())
}

func Test_defaultConfig(t *testing.T) {
	mainBuf := bytes.Buffer{}
	hostsBuf := bytes.Buffer{}
	defer StubWriteConfig(&mainBuf, &hostsBuf)()

	cfg := NewBlankConfig()
	assert.NoError(t, cfg.Write())

	expected := heredoc.Doc(`
		# What protocol to use when performing git operations. Supported values: https
		protocol: https
		# What editor inst should run. If blank, will refer to environment.
		editor:
		# When to interactively prompt. This is a global config that cannot be overridden by hostname. Supported values: enabled, disabled
		prompt: enabled
		# A pager program to send command output to, e.g. "less". Set the value to "cat" to disable the pager.
		pager:
		# The path to a unix socket through which send HTTP connections. If blank, HTTP traffic will be handled by net/http.DefaultTransport.
		http_unix_socket:
		# What web browser inst should use when opening URLs. If blank, will refer to environment.
		browser:
		# The default hostname to use for commands that require a hostname, e.g. inst instance list.
		default_hostname:
	`)
	assert.Equal(t, expected, mainBuf.String())
	assert.Equal(t, "", hostsBuf.String())

	proto, err := cfg.Get("", "protocol")
	assert.NoError(t, err)
	assert.Equal(t, "https", proto)

	editor, err := cfg.Get("", "editor")
	assert.NoError(t, err)
	assert.Equal(t, "", editor)

	browser, err := cfg.Get("", "browser")
	assert.NoError(t, err)
	assert.Equal(t, "", browser)
}

func Test_ValidateValue(t *testing.T) {
	err := ValidateValue("protocol", "sshpps")
	assert.EqualError(t, err, "invalid value")

	err = ValidateValue("protocol", "ssh")
	assert.NoError(t, err)

	err = ValidateValue("editor", "vim")
	assert.NoError(t, err)

	err = ValidateValue("got", "123")
	assert.NoError(t, err)

	err = ValidateValue("http_unix_socket", "really_anything/is/allowed/and/net.Dial\\(...\\)/will/ultimately/validate")
	assert.NoError(t, err)
}

func Test_ValidateKey(t *testing.T) {
	err := ValidateKey("invalid")
	assert.EqualError(t, err, "invalid key")

	err = ValidateKey("protocol")
	assert.NoError(t, err)

	err = ValidateKey("editor")
	assert.NoError(t, err)

	err = ValidateKey("prompt")
	assert.NoError(t, err)

	err = ValidateKey("pager")
	assert.NoError(t, err)

	err = ValidateKey("http_unix_socket")
	assert.NoError(t, err)

	err = ValidateKey("browser")
	assert.NoError(t, err)
}
