package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmd/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

// UndeployOptions contains the command line options
type UndeployOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	OS             OSDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
}

// NewUndeployCmd creates a new command
func NewUndeployCmd(f *cmdutil.Factory, runF func(*UndeployOptions) error) *cobra.Command {
	opts := &UndeployOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "undeploy",
		Short: "Undeploy a local Instill Core instance",
		Example: heredoc.Doc(`
			$ inst local undeploy
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

			return runUndeploy(opts)
		},
	}

	return cmd
}

func runUndeploy(opts *UndeployOptions) error {

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
		p(opts.IO, fmt.Sprintf("Tearing down Instill %s...", projs[i]))
		out, err := execCmd(opts.Exec, "make", "down")
		if err != nil {
			return fmt.Errorf("ERROR: when tearing down, %w", err)
		}
		if err != nil {
			return fmt.Errorf("ERROR: %s when tearing down, %w\n%s", projs[i], err, out)
		}
	}

	p(opts.IO, "Remove local Instill Core files in: %s", path)
	os.RemoveAll(path)

	if err := unregisterInstance(opts); err != nil {
		return err
	}

	p(opts.IO, "")
	p(opts.IO, "Instill Core undeployed")

	return nil
}

func unregisterInstance(opts *UndeployOptions) error {
	exists, err := instance.IsInstanceAdded(opts.Config, "localhost:8080")
	if err != nil {
		return err
	}
	if exists {
		addOpts := &instance.RemoveOptions{
			IO:             opts.IO,
			Config:         opts.Config,
			MainExecutable: opts.MainExecutable,
			Interactive:    false,
			APIHostname:    "localhost:8080",
		}
		p(opts.IO, "")
		err = instance.RunRemove(addOpts)
		if err != nil {
			return err
		}
	}
	return nil
}
