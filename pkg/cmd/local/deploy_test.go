package local

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func TestLocalDeployCmd(t *testing.T) {
	d, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Couldn't get home directory", err)
	}
	dir := filepath.Join(d, ".local", "instill") + string(os.PathSeparator)
	tests := []struct {
		name     string
		stdin    string
		stdinTTY bool
		input    string
		output   DeployOptions
		isErr    bool
	}{
		{
			name:  "no arguments",
			input: "",
			output: DeployOptions{
				Path: dir,
			},
			isErr: false,
		},
		{
			name:  "local deploy --path /home",
			input: " --path /home",
			output: DeployOptions{
				Path: "/home",
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

			var gotOpts *DeployOptions
			cmd := NewDeployCmd(f, func(opts *DeployOptions) error {
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
			assert.Equal(t, tt.output.Path, gotOpts.Path)
		})
	}
}

func TestLocalDeployCmdRun(t *testing.T) {
	execMock := &ExecMock{}
	osMock := &OSMock{}
	d, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Couldn't get home directory", err)
	}
	dir := filepath.Join(d, ".local", "instill") + string(os.PathSeparator)
	tests := []struct {
		name     string
		input    *DeployOptions
		stdout   string
		stderr   string
		isErr    bool
		expectFn func(*testing.T, config.Config)
	}{
		{
			name: "local deploy",
			input: &DeployOptions{
				Path:   dir,
				Exec:   execMock,
				OS:     osMock,
				Config: config.ConfigStub{},
			},
			stdout: "",
			isErr:  false,
		},
	}

	for _, tt := range tests {
		io, _, stdout, stderr := iostreams.Test()
		io.SetStdoutTTY(true)
		tt.input.IO = io

		t.Run(tt.name, func(t *testing.T) {
			err := runDeploy(tt.input)
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
