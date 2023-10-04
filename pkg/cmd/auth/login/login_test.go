package login

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func Test_NewCmdLogin(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		stdin    string
		stdinTTY bool
		wants    LoginOptions
		wantsErr bool
	}{
		{
			name:     "tty, hostname",
			stdinTTY: true,
			cli:      "--hostname barry.burton",
			wants: LoginOptions{
				Hostname:    "barry.burton",
				Interactive: true,
			},
		},
		{
			name:     "tty",
			stdinTTY: true,
			cli:      "",
			wants: LoginOptions{
				Hostname:    instance.FallbackHostname(),
				Interactive: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, stdin, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams:  io,
				Executable: func() string { return "/path/to/inst" },
				Config:     config.ConfigStubFactory,
			}

			io.SetStdoutTTY(true)
			io.SetStdinTTY(tt.stdinTTY)
			if tt.stdin != "" {
				stdin.WriteString(tt.stdin)
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *LoginOptions
			cmd := NewCmdLogin(f, func(opts *LoginOptions) error {
				gotOpts = opts
				return nil
			})
			// TODO cobra hack-around
			cmd.Flags().BoolP("help", "x", false, "")

			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.Hostname, gotOpts.Hostname)
			assert.Equal(t, tt.wants.Interactive, gotOpts.Interactive)
		})
	}
}

type mockRoundTripper struct {
	response *http.Response
}

func (rt *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return rt.response, nil
}

func TestLocalLogin(t *testing.T) {
	// fixtures
	token1 := "foobar"
	json := `{ "access_token": "` + token1 + `" }`
	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "application/json")
	_, err := recorder.WriteString(json)
	assert.NoError(t, err)
	expectedResponse := recorder.Result()
	transport := &mockRoundTripper{expectedResponse}
	// test
	token2, err := loginLocal(transport, "baz", "baz")
	// assert
	assert.NoError(t, err)
	assert.Equal(t, token1, token2)

}
