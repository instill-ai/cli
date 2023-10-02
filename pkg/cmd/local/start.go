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
		p(opts.IO, fmt.Sprintf("Starting %s...", projs[i]))
		out, err := execCmd(opts.Exec, "make", "start")
		if err != nil {
			return fmt.Errorf("ERROR: when starting, %w", err)
		}
		if err != nil {
			return fmt.Errorf("ERROR: %s when starting, %w\n%s", projs[i], err, out)
		}
	}

	p(opts.IO, "Instill Core started")

	return nil
}
