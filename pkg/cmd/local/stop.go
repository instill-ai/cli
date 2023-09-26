package instances

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type StopOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
	Path           string `validate:"required,dirpath"`
}

func NewStopCmd(f *cmdutil.Factory, runF func(*StopOptions) error) *cobra.Command {
	opts := &StopOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a local Instill Core instance",
		Example: heredoc.Doc(`
			# stop to /home/me/instill
			$ inst local stop --path /home/me/instill
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
	// init, validate
	var err error
	io2 := opts.IO
	path, err := ConfigPath(opts.Config)
	if err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}
	if err := IsDeployed(path); err != nil {
		return fmt.Errorf("ERROR: %s", err)
	}
	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("ERROR: can't open the destination, %w", err)
	}

	p(io2, "Stopping...")
	out, err := execCmd(opts.Exec, "make", "stop")
	slog.Debug("make stop", "out", out)
	if err != nil {
		return fmt.Errorf("ERROR: when stopping, %w", err)
	}

	p(io2, "Instill Core stopped")

	return nil
}
