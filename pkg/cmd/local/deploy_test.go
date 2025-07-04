package local

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func TestLocalDeployCmd(t *testing.T) {

	tests := []struct {
		name     string
		stdin    string
		stdinTTY bool
		input    string
		output   DeployOptions
		isErr    bool
	}{
		{
			name:   "no arguments",
			input:  "",
			output: DeployOptions{},
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

			cmd := NewDeployCmd(f, func(opts *DeployOptions) error {
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

func checkForUpdateMock(string, string, string) (*releaseInfo, error) {
	return &releaseInfo{}, nil
}

func TestLocalDeployCmdRun(t *testing.T) {
	execMock := &ExecMock{}
	osMock := &OSMock{}
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
				Exec:           execMock,
				OS:             osMock,
				Config:         config.ConfigStub{},
				checkForUpdate: checkForUpdateMock,
				isDeployed: func(ExecDep) error {
					return nil
				},
			},
			stdout: "",
			isErr:  false,
		},
		{
			name: "local deploy",
			input: &DeployOptions{
				Exec:           execMock,
				OS:             osMock,
				Config:         config.ConfigStub{},
				checkForUpdate: checkForUpdateMock,
				isDeployed: func(ExecDep) error {
					return fmt.Errorf("")
				},
			},
			stdout: "",
			isErr:  false,
		},
		{
			name: "local deploy",
			input: &DeployOptions{
				Exec:           execMock,
				OS:             osMock,
				Config:         config.ConfigStub{},
				checkForUpdate: checkForUpdateMock,
				isDeployed: func(ExecDep) error {
					return nil
				},
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
