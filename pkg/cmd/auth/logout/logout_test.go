package logout

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"

	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func Test_NewCmdLogout(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		wants    LogoutOptions
		wantsErr bool
		tty      bool
	}{
		{
			name: "tty with hostname",
			tty:  true,
			cli:  "--hostname harry.mason",
			wants: LogoutOptions{
				Hostname: "harry.mason",
			},
		},
		{
			name: "tty no arguments",
			tty:  true,
			cli:  "",
			wants: LogoutOptions{
				Hostname: "",
			},
		},
		{
			name: "nontty with hostname",
			cli:  "--hostname harry.mason",
			wants: LogoutOptions{
				Hostname: "harry.mason",
			},
		},
		{
			name:     "nontty no arguments",
			cli:      "",
			wantsErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: io,
			}
			io.SetStdinTTY(tt.tty)
			io.SetStdoutTTY(tt.tty)

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *LogoutOptions
			cmd := NewCmdLogout(f, func(opts *LogoutOptions) error {
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
		})

	}
}
