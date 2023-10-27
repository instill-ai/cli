package local

import (
	"fmt"
	"os"
	"path/filepath"

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
	var err error
	projDirPath := filepath.Join(LocalInstancePath, "core")
	_, err = os.Stat(projDirPath)
	if !os.IsNotExist(err) {
		if opts.OS != nil {
			err = opts.OS.Chdir(projDirPath)
		} else {
			err = os.Chdir(projDirPath)
		}
		if err != nil {
			return fmt.Errorf("ERROR: cannot open the directory: %w", err)
		}
		p(opts.IO, "Tearing down Instill Core...")
		_, err = os.Stat(filepath.Join(projDirPath, "Makefile"))
		if !os.IsNotExist(err) {
			if out, err := execCmd(opts.Exec, "bash", "-c", "make down"); err != nil {
				fmt.Println(fmt.Errorf("ERROR: when tearing down Instill Core, %w\n%s, continue to tear down", err, out))
			}
		}
	}

	_, err = os.Stat(LocalInstancePath)
	if os.IsNotExist(err) {
		p(opts.IO, "")
		p(opts.IO, "Instill Core instance not deployed")
	} else {
		if os.RemoveAll(LocalInstancePath); err != nil {
			return fmt.Errorf("ERROR: cannot remove %s, %w", LocalInstancePath, err)
		}
		p(opts.IO, "")
		p(opts.IO, "Instill Core instance undeployed successfully!")
	}

	if err := unregisterInstance(opts); err != nil {
		return err
	}

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
