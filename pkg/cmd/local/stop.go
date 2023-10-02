package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

// StopOptions contains the command line options
type StopOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	OS             OSDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
}

// NewStopCmd creates a new command
func NewStopCmd(f *cmdutil.Factory, runF func(*StopOptions) error) *cobra.Command {
	opts := &StopOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a local Instill Core instance",
		Example: heredoc.Doc(`
			$ inst local stop
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			opts.Config = cfg

			if opts.IO.CanPrompt() {
				opts.Interactive = true
			}

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return runStop(opts)
		},
	}

	return cmd
}

func runStop(opts *StopOptions) error {

	path, err := getConfigPath(opts.Config)
	if err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}
	if err := isDeployed(opts.Exec); err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}

	for i := range projs {
		proj := strings.ToLower(projs[i])
		if opts.OS != nil {
			err = opts.OS.Chdir(filepath.Join(path, proj))
		} else {
			err = os.Chdir(filepath.Join(path, proj))
		}
		if err != nil {
			return fmt.Errorf("ERROR: can't open the destination, %w", err)
		}
		p(opts.IO, fmt.Sprintf("Stopping %s...", projs[i]))
		out, err := execCmd(opts.Exec, "make", "stop")
		if err != nil {
			return fmt.Errorf("ERROR: when stopping, %w", err)
		}
		if err != nil {
			return fmt.Errorf("ERROR: %s when stopping, %w\n%s", projs[i], err, out)
		}
	}

	p(opts.IO, "Instill Core stopped")

	return nil
}
