package instances

import (
	"bytes"
	"github.com/instill-ai/cli/pkg/iostreams"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
)

func TestInstancesRemoveCmd(t *testing.T) {
	tests := []struct {
		name     string
		stdin    string
		stdinTTY bool
		input    string
		output   RemoveOptions
		isErr    bool
	}{
		{
			name:   "no arguments",
			input:  "",
			output: RemoveOptions{},
			isErr:  true,
		},
		{
			name:  "instances remove api.instill.tech",
			input: "api.instill.tech",
			output: RemoveOptions{
				APIHostname: "api.instill.tech",
			},
			isErr: false,
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

			var gotOpts *RemoveOptions
			cmd := NewRemoveCmd(f, func(opts *RemoveOptions) error {
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
		})
	}
}

func TestInstancesRemoveCmdRun(t *testing.T) {
	tests := []struct {
		name     string
		input    *RemoveOptions
		stdout   string
		stderr   string
		isErr    bool
		expectFn func(*testing.T, config.Config)
	}{
		{
			name: "instances remove api.instill.tech",
			input: &RemoveOptions{
				APIHostname: "api.instill.tech",
				Config:      config.ConfigStub{},
			},
			stdout: "Instance 'api.instill.tech' has been removed\n",
			isErr:  false,
		},
	}

	for _, tt := range tests {
		io, _, stdout, stderr := iostreams.Test()
		io.SetStdoutTTY(true)
		tt.input.IO = io

		t.Run(tt.name, func(t *testing.T) {
			err := runRemove(tt.input)
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
