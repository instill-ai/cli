package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/build"
	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmd/factory"
	"github.com/instill-ai/cli/pkg/cmdutil"
)

// ExecDep is an interface for executing commands
type ExecDep interface {
	Command(name string, arg ...string) *exec.Cmd
	LookPath(file string) (string, error)
}

// OSDep is an interface for OS operations
type OSDep interface {
	Chdir(path string) error
	Stat(name string) (os.FileInfo, error)
}

// LocalInstancePath is the path to keep files for local instance deployment
var LocalInstancePath string

// releaseInfo stores information about a release
type releaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

type stateEntry struct {
	CheckedForUpdateAt time.Time   `yaml:"checked_for_update_at"`
	LatestRelease      releaseInfo `yaml:"latest_release"`
}

var p = cmdutil.P

// New creates a new command
func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local <command>",
		Short: "Local Instill Core instance",
		Long:  `Create and manage a local Instill Core instance with ease.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			d, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			LocalInstancePath = filepath.Join(d, ".local", "instill")

			// check for update
			if cmd.Flags().Lookup("upgrade") == nil || (cmd.Flags().Lookup("upgrade") != nil && !cmd.Flags().Lookup("upgrade").Changed) {
				projDirPath := filepath.Join(LocalInstancePath, "core")
				_, err = os.Stat(projDirPath)
				if !os.IsNotExist(err) {
					if err = os.Chdir(projDirPath); err != nil {
						return err
					}
					if currentVersion, err := execCmd(nil, "bash", "-c", "git name-rev --tags --name-only $(git rev-parse HEAD)"); err == nil {
						currentVersion = strings.Trim(currentVersion, "\n")
						if currentVersion != "undefined" {
							if newRelease, err := checkForUpdate(nil, filepath.Join(config.StateDir(), "core.yml"), "instill-ai/core", currentVersion); err != nil {
								return fmt.Errorf("ERROR: cannot check for the update Instill Core, %w:\n%s", err, currentVersion)
							} else if newRelease != nil {
								cmdFactory := factory.New(build.Version)
								stderr := cmdFactory.IOStreams.ErrOut
								fmt.Fprintf(stderr, "\n%s %s → %s\n",
									ansi.Color("A new release of Instill Core is available:", "yellow"),
									ansi.Color(currentVersion, "cyan"),
									ansi.Color(newRelease.Version, "cyan"))
								fmt.Fprintf(stderr, "%s\n\n",
									ansi.Color("Run 'inst local deploy --upgrade' to deploy the latest version", "yellow"))
							}
						}
					}
				}
			}

			return nil
		},
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(NewUndeployCmd(f, nil))
	cmd.AddCommand(NewDeployCmd(f, nil))
	cmd.AddCommand(NewStartCmd(f, nil))
	cmd.AddCommand(NewStopCmd(f, nil))
	cmd.AddCommand(NewStatusCmd(f, nil))

	return cmd
}

func execCmd(execDep ExecDep, cmd string, params ...string) (string, error) {
	var c *exec.Cmd
	if execDep != nil {
		c = execDep.Command(cmd, params...)
	} else {
		c = exec.Command(cmd, params...)
	}
	out, err := c.Output()
	outStr := strings.Trim(string(out[:]), " ")
	return outStr, err
}
