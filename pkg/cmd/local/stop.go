package local

import (
	"fmt"
	"os"
	"path/filepath"

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

	if _, err := os.Stat(LocalInstancePath); os.IsNotExist(err) {
		p(opts.IO, "")
		p(opts.IO, "Instill Core instance not deployed")
		return nil
	}

	projDirPath := filepath.Join(LocalInstancePath, "core")
	if err := isProjectDeployed(opts.Exec, "core"); err == nil {
		if opts.OS != nil {
			err = opts.OS.Chdir(projDirPath)
		} else {
			err = os.Chdir(projDirPath)
		}
		if err != nil {
			return fmt.Errorf("ERROR: cannot open the directory: %w", err)
		}
		p(opts.IO, "Stopping Instill Core...")
		out, err := execCmd(opts.Exec, "make", "stop")
		if err != nil {
			return fmt.Errorf("ERROR: when stopping, %w", err)
		}
		if err != nil {
			return fmt.Errorf("ERROR: when stopping Instill Core, %w\n%s", err, out)
		}
	} else {
		return fmt.Errorf("ERROR: %w", err)
	}

	p(opts.IO, "Instill Core stopped")

	return nil
}
