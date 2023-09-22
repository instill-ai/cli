package instances

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func TestInstancesAddCmd(t *testing.T) {
	tests := []struct {
		name     string
		stdin    string
		stdinTTY bool
		input    string
		output   AddOptions
		isErr    bool
	}{
		{
			name:   "no arguments",
			input:  "",
			output: AddOptions{},
			isErr:  true,
		},
		{
			name:  "instances add foo --default",
			input: "foo --default",
			output: AddOptions{
				InstanceOptions: InstanceOptions{
					APIHostname: "foo",
					Default:     true,
				},
			},
			isErr: false,
		},
		{
			name:  "instances add foo --oauth2 bar",
			input: "foo --oauth2 bar",
			output: AddOptions{
				InstanceOptions: InstanceOptions{
					APIHostname: "foo",
					Oauth2:      "bar",
				}},
			isErr: false,
		},
		{
			name:  "wrong hostname",
			input: "foo|bar",
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, stdin, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				Config: func() (config.Config, error) {
					return config.ConfigStub{}, nil
				},
				IOStreams:  io,
				Executable: func() string { return "/path/to/instill" },
			}

			io.SetStdoutTTY(true)
			io.SetStdinTTY(tt.stdinTTY)
			if tt.stdin != "" {
				stdin.WriteString(tt.stdin)
			}

			argv, err := shlex.Split(tt.input)
			assert.NoError(t, err)

			var gotOpts *AddOptions
			cmd := NewAddCmd(f, func(opts *AddOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.Flags().BoolP("help", "x", false, "")

			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.isErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.output.APIHostname, gotOpts.APIHostname)
			assert.Equal(t, tt.output.Oauth2, gotOpts.Oauth2)
		})
	}
}

func TestInstancesAddCmdRun(t *testing.T) {
	tests := []struct {
		name     string
		input    *AddOptions
		stdout   string
		stderr   string
		isErr    bool
		expectFn func(*testing.T, config.Config)
	}{
		{
			name: "instances add foo --default",
			input: &AddOptions{
				InstanceOptions: InstanceOptions{
					APIHostname: "foo",
					Default:     true,
				},
				Config: config.ConfigStub{},
			},
			stdout: "Instance 'foo' has been added\n",
			isErr:  false,
		},
		{
			name: "instances add foo --oauth2 bar1 --secret bar2 --client-id bar3",
			input: &AddOptions{
				Config: config.ConfigStub{},
				InstanceOptions: InstanceOptions{
					APIHostname: "foo",
					Oauth2:      "bar1",
					Secret:      "bar2",
					ClientID:    "bar3",
				},
			},
			expectFn: func(t *testing.T, cfg config.Config) {
				v, err := cfg.Get("foo", "oauth2_hostname")
				assert.NoError(t, err)
				assert.Equal(t, "bar1", v)
				v, err = cfg.Get("foo", "oauth2_client_secret")
				assert.NoError(t, err)
				assert.Equal(t, "bar2", v)
				v, err = cfg.Get("foo", "oauth2_client_id")
				assert.NoError(t, err)
				assert.Equal(t, "bar3", v)
			},
			stdout: "Instance 'foo' has been added\n",
			isErr:  false,
		},
		{
			name: "missing oauth2 secret and client id",
			input: &AddOptions{
				Config: config.ConfigStub{},
				InstanceOptions: InstanceOptions{
					APIHostname: "foo",
					Oauth2:      "bar1",
				},
			},
			isErr: true,
		},
	}

	for _, tt := range tests {
		io, _, stdout, stderr := iostreams.Test()
		io.SetStdoutTTY(true)
		tt.input.IO = io

		t.Run(tt.name, func(t *testing.T) {
			err := runAdd(tt.input)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.stdout, stdout.String())
			assert.Equal(t, tt.stderr, stderr.String())
			if tt.expectFn != nil {
				tt.expectFn(t, tt.input.Config)
			}
		})
	}
}
