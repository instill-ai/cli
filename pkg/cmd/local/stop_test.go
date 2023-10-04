package local

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func TestLocalStopCmd(t *testing.T) {
	tests := []struct {
		name     string
		stdin    string
		stdinTTY bool
		input    string
		output   StopOptions
		isErr    bool
	}{
		{
			name:   "no arguments",
			input:  "",
			output: StopOptions{},
			isErr:  false,
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

			cmd := NewStopCmd(f, func(opts *StopOptions) error {
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
		})
	}
}

func TestLocalStopCmdRun(t *testing.T) {
	execMock := &ExecMock{}
	osMock := &OSMock{}
	cfg := config.ConfigStub{}
	_ = cfg.Set("", ConfigKeyPath, "/foo/bar")
	tests := []struct {
		name     string
		input    *StopOptions
		stdout   string
		stderr   string
		isErr    bool
		expectFn func(*testing.T, config.Config)
	}{
		{
			name: "local stop",
			input: &StopOptions{
				Exec:   execMock,
				OS:     osMock,
				Config: cfg,
			},
			stdout: "Instill Core stopped",
			isErr:  false,
		},
	}

	for _, tt := range tests {
		io, _, stdout, stderr := iostreams.Test()
		io.SetStdoutTTY(true)
		tt.input.IO = io

		t.Run(tt.name, func(t *testing.T) {
			err := runStop(tt.input)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Regexp(t, tt.stdout, stdout.String())
			assert.Equal(t, tt.stderr, stderr.String())
			if tt.expectFn != nil {
				tt.expectFn(t, tt.input.Config)
			}
		})
	}
}
