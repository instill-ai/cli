package instance

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func TestInstanceEditCmd(t *testing.T) {
	tests := []struct {
		name     string
		stdin    string
		stdinTTY bool
		input    string
		output   EditOptions
		isErr    bool
	}{
		{
			name:   "no arguments",
			input:  "",
			output: EditOptions{},
			isErr:  true,
		},
		{
			name:  "instance edit api.instill.tech --default",
			input: "api.instill.tech --default",
			output: EditOptions{
				InstanceOptions: InstanceOptions{
					APIHostname: "api.instill.tech",
					Default:     true,
				},
			},
			isErr: false,
		},
		{
			name:  "instance edit api.instill.tech --oauth2 bar",
			input: "api.instill.tech --oauth2 bar",
			output: EditOptions{
				InstanceOptions: InstanceOptions{
					APIHostname: "api.instill.tech",
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
				Executable: func() string { return "/path/to/inst" },
			}

			io.SetStdoutTTY(true)
			io.SetStdinTTY(tt.stdinTTY)
			if tt.stdin != "" {
				stdin.WriteString(tt.stdin)
			}

			argv, err := shlex.Split(tt.input)
			assert.NoError(t, err)

			var gotOpts *EditOptions
			cmd := NewEditCmd(f, func(opts *EditOptions) error {
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

func TestInstanceEditCmdRun(t *testing.T) {
	tests := []struct {
		name     string
		input    *EditOptions
		stdout   string
		stderr   string
		isErr    bool
		expectFn func(*testing.T, config.Config)
	}{
		{
			name: "instance edit api.instill.tech --default",
			input: &EditOptions{
				InstanceOptions: InstanceOptions{
					APIHostname: "api.instill.tech",
					Default:     true,
				},
				Config: config.ConfigStub{},
			},
			stdout: "Instance 'api.instill.tech' has been saved\n",
			isErr:  false,
		},
		{
			name: "instance edit api.instill.tech --oauth2 bar1 --client-secret bar2 --client-id bar3",
			input: &EditOptions{
				Config: config.ConfigStub{},
				InstanceOptions: InstanceOptions{
					APIHostname:  "api.instill.tech",
					Oauth2:       "bar1",
					ClientSecret: "bar2",
					ClientID:     "bar3",
				},
			},
			expectFn: func(t *testing.T, cfg config.Config) {
				v, err := cfg.Get("api.instill.tech", "oauth2_hostname")
				assert.NoError(t, err)
				assert.Equal(t, "bar1", v)
				v, err = cfg.Get("api.instill.tech", "oauth2_client_secret")
				assert.NoError(t, err)
				assert.Equal(t, "bar2", v)
				v, err = cfg.Get("api.instill.tech", "oauth2_client_id")
				assert.NoError(t, err)
				assert.Equal(t, "bar3", v)
			},
			stdout: "Instance 'api.instill.tech' has been saved\n",
			isErr:  false,
		},
		{
			name: "instance edit api.instill.tech --no-auth",
			input: &EditOptions{
				Config: config.ConfigStub{},
				InstanceOptions: InstanceOptions{
					APIHostname: "api.instill.tech",
				},
				NoAuth: true,
			},
			stdout: "Instance 'api.instill.tech' has been saved\n",
			isErr:  false,
		},
	}

	for _, tt := range tests {
		io, _, stdout, stderr := iostreams.Test()
		io.SetStdoutTTY(true)
		tt.input.IO = io

		t.Run(tt.name, func(t *testing.T) {
			err := runEdit(tt.input)
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
