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

type StartOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
}

func NewStartCmd(f *cmdutil.Factory, runF func(*StartOptions) error) *cobra.Command {
	opts := &StartOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a local Instill Core instance",
		Example: heredoc.Doc(`
			# start to /home/me/instill
			$ inst local start --path /home/me/instill
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

			return runStart(opts)
		},
	}

	return cmd
}

func runStart(opts *StartOptions) error {
	// init, validate
	var err error
	io2 := opts.IO
	path, err := ConfigPath(opts.Config)
	if err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}
	if err := IsDeployed(path); err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}
	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("ERROR: can't open the destination, %w", err)
	}

	p(io2, "Starting...")
	out, err := execCmd(opts.Exec, "make", "start")
	slog.Debug("make start", "out", out)
	if err != nil {
		return fmt.Errorf("ERROR: when starting, try to restart dockerd.t\n%w", err)
	}

	p(io2, "Instill Core started")

	return nil
}
