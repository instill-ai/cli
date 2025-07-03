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

type StartOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	OS             OSDep
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
			$ inst local start
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

	if _, err := os.Stat(LocalInstancePath); os.IsNotExist(err) {
		p(opts.IO, "")
		p(opts.IO, "Instill Core instance not deployed")
		return nil
	}

	projDirPath := filepath.Join(LocalInstancePath, "instill-core")
	if err := isDeployed(opts.Exec); err == nil {
		if opts.OS != nil {
			err = opts.OS.Chdir(projDirPath)
		} else {
			err = os.Chdir(projDirPath)
		}
		if err != nil {
			return fmt.Errorf("ERROR: cannot open the directory: %w", err)
		}
		p(opts.IO, "")
		p(opts.IO, "Starting Instill Core...")
		p(opts.IO, "")
		err := execCmdStream(opts.Exec, opts.IO, "make", "start")
		if err != nil {
			return fmt.Errorf("ERROR: when starting, %w", err)
		}
	} else {
		return fmt.Errorf("ERROR: %w", err)
	}

	p(opts.IO, "")
	p(opts.IO, "Instill Core started")

	return nil
}
