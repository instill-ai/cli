package local

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/safeexec"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmd/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

const (
	// DefUsername is the default username for the local deployment
	DefUsername = "admin"
	// DefPassword is the default password for the local deployment
	DefPassword = "password"
)

// DeployOptions contains the command line options
type DeployOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	OS             OSDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
	Force          bool
	Upgrade        bool
	Latest         bool
	Build          bool
	checkForUpdate func(ExecDep, string, string, string) (*releaseInfo, error)
	isDeployed     func(ExecDep, string) error
}

// NewDeployCmd creates a new command
func NewDeployCmd(f *cmdutil.Factory, runF func(*DeployOptions) error) *cobra.Command {
	opts := &DeployOptions{
		IO:             f.IOStreams,
		checkForUpdate: checkForUpdate,
		isDeployed:     isProjectDeployed,
	}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a local Instill Core instance",
		Example: heredoc.Doc(`
			$ inst local deploy
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

			return runDeploy(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force to deploy a new local Instill Core instance")
	cmd.Flags().BoolVarP(&opts.Upgrade, "upgrade", "u", false, "Upgrade Instill Core instance to the latest release version")
	cmd.Flags().BoolVarP(&opts.Build, "build", "b", false, "Deploy an Instill Core instance and build latest release version")
	cmd.Flags().BoolVarP(&opts.Latest, "latest", "l", false, "Deploy an Instill Core instance with the latest version (unstable)")
	cmd.MarkFlagsMutuallyExclusive("force", "upgrade")
	cmd.MarkFlagsMutuallyExclusive("upgrade", "latest")

	return cmd
}

func runDeploy(opts *DeployOptions) error {

	var err error
	start := time.Now()

	// check the deps
	apps := []string{"bash", "docker", "make", "git", "jq", "grep", "curl"}
	for _, n := range apps {
		if opts.Exec != nil {
			_, err = opts.Exec.LookPath(n)
		} else {
			_, err = safeexec.LookPath(n)
		}
		if err != nil {
			return fmt.Errorf("ERROR: %s not found", n)
		}
	}

	// download the latest version of the projects, if local repos are not present
	projDirPath := filepath.Join(LocalInstancePath, "core")
	_, err = os.Stat(projDirPath)
	if os.IsNotExist(err) {
		if opts.Latest {
			if out, err := execCmd(opts.Exec, "bash", "-c",
				fmt.Sprintf("git clone --depth 1 https://github.com/instill-ai/core.git %s", projDirPath)); err != nil {
				return fmt.Errorf("ERROR: cannot clone Instill Core repo, %w:\n%s", err, out)
			}
		} else {
			if latestReleaseVersion, err := execCmd(opts.Exec, "bash", "-c", "curl https://api.github.com/repos/instill-ai/core/releases | jq -r 'map(select(.prerelease)) | first | .tag_name'"); err == nil {
				latestReleaseVersion = strings.Trim(latestReleaseVersion, "\n")
				if out, err := execCmd(opts.Exec, "bash", "-c",
					fmt.Sprintf("git clone --depth 1 -b %s -c advice.detachedHead=false https://github.com/instill-ai/core.git %s", latestReleaseVersion, projDirPath)); err != nil {
					return fmt.Errorf("ERROR: cannot clone core, %w:\n%s", err, out)
				}
				if _, err := opts.checkForUpdate(opts.Exec, filepath.Join(config.StateDir(), "core.yml"), "instill-ai/core", latestReleaseVersion); err != nil {
					return fmt.Errorf("ERROR: cannot check for the update of Instill Core, %w:\n%s", err, latestReleaseVersion)
				}
			} else {
				return fmt.Errorf("ERROR: cannot find latest release version of Instill Core, %w:\n%s", err, latestReleaseVersion)
			}
		}
	}

	if opts.Force {
		p(opts.IO, "Tear down Instill Core instance if existing...")
		projDirPath := filepath.Join(LocalInstancePath, "core")
		_, err = os.Stat(projDirPath)
		if !os.IsNotExist(err) {
			if opts.OS != nil {
				if err = opts.OS.Chdir(projDirPath); err != nil {
					return err
				}
			} else {
				if err = os.Chdir(projDirPath); err != nil {
					return err
				}
			}
			if out, err := execCmd(opts.Exec, "bash", "-c", "make down"); err != nil {
				return fmt.Errorf("ERROR: cannot force tearing down Instill Core, %w:\n%s", err, out)
			}
		}

	} else if opts.Upgrade && !opts.Latest {
		hasNewVersion := false
		projDirPath := filepath.Join(LocalInstancePath, "core")
		_, err = os.Stat(projDirPath)
		if !os.IsNotExist(err) {
			if opts.OS != nil {
				if err = opts.OS.Chdir(projDirPath); err != nil {
					return err
				}
			} else {
				if err = os.Chdir(projDirPath); err != nil {
					return err
				}
			}
			if currentVersion, err := execCmd(opts.Exec, "bash", "-c", "git name-rev --tags --name-only $(git rev-parse HEAD)"); err == nil {
				currentVersion = strings.Trim(currentVersion, "\n")
				if currentVersion == "undefined" {
					currentVersion = "latest"
				}
				if newRelease, err := opts.checkForUpdate(opts.Exec, filepath.Join(config.StateDir(), "core.yml"), "instill-ai/core", currentVersion); err != nil {
					return fmt.Errorf("ERROR: cannot check for the update of Instill Core, %w:\n%s", err, currentVersion)
				} else if newRelease != nil {
					p(opts.IO, "Upgrade Instill Core to %s...", newRelease.Version)
					hasNewVersion = true
				}
			} else {
				return fmt.Errorf("ERROR: cannot find current release version of Instill Core, %w:\n%s", err, currentVersion)
			}
		}

		if hasNewVersion {
			p(opts.IO, "Tear down Instill Core instance if existing...")
			projDirPath := filepath.Join(LocalInstancePath, "core")
			_, err = os.Stat(projDirPath)
			if !os.IsNotExist(err) {
				if opts.OS != nil {
					if err = opts.OS.Chdir(projDirPath); err != nil {
						return err
					}
				} else {
					if err = os.Chdir(projDirPath); err != nil {
						return err
					}
				}
				if out, err := execCmd(opts.Exec, "bash", "-c", "make down"); err != nil {
					return fmt.Errorf("ERROR: cannot force tearing down Instill Core, %w:\n%s", err, out)
				}
			}

			if dir, err := os.ReadDir(projDirPath); err == nil {
				for _, d := range dir {
					if os.RemoveAll(path.Join([]string{projDirPath, d.Name()}...)); err != nil {
						return fmt.Errorf("ERROR: cannot remove %s, %w", projDirPath, err)
					}
				}
			} else {
				return fmt.Errorf("ERROR: cannot read %s, %w", projDirPath, err)
			}

			if latestVersion, err := execCmd(opts.Exec, "bash", "-c", fmt.Sprintf("curl https://api.github.com/repos/instill-ai/%s/releases | jq -r 'map(select(.prerelease)) | first | .tag_name'", "core")); err == nil {
				latestVersion = strings.Trim(latestVersion, "\n")
				if out, err := execCmd(opts.Exec, "bash", "-c",
					fmt.Sprintf("git clone --depth 1 -b %s -c advice.detachedHead=false https://github.com/instill-ai/core.git %s", latestVersion, projDirPath)); err != nil {
					return fmt.Errorf("ERROR: cannot clone Instill Core, %w:\n%s", err, out)
				}
			} else {
				return fmt.Errorf("ERROR: cannot find latest release version of Instill Core, %w:\n%s", err, latestVersion)
			}
		} else {
			p(opts.IO, "No upgrade available")
			return nil
		}

	} else {
		if err := opts.isDeployed(opts.Exec, "core"); err == nil {
			p(opts.IO, "A local Instill Core deployment detected")
			return nil
		}
	}

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
	}

	if opts.Latest {
		p(opts.IO, "Spin up latest Instill Core...")
		if out, err := execCmd(opts.Exec, "bash", "-c", "make latest"); err != nil {
			return fmt.Errorf("ERROR: Instill Core spin-up failed, %w\n%s", err, out)
		}
	} else {
		if currentVersion, err := execCmd(opts.Exec, "bash", "-c", "git name-rev --tags --name-only $(git rev-parse HEAD)"); err == nil {
			currentVersion = strings.Trim(currentVersion, "\n")
			p(opts.IO, "Spin up Instill Core %s...", currentVersion)
		} else {
			return fmt.Errorf("ERROR: cannot get the current tag of Instill Core repo, %w:\n%s", err, currentVersion)
		}
		if out, err := execCmd(opts.Exec, "bash", "-c", fmt.Sprintf("make all BUILD=%s", strconv.FormatBool(opts.Build))); err != nil {
			return fmt.Errorf("ERROR: Instill Core spin-up failed, %w\n%s", err, out)
		}
	}

	// print a summary
	elapsed := time.Since(start)
	p(opts.IO, "")
	p(opts.IO, `
			Instill Core console available at http://localhost:3000

			User:     %s
			Password: %s

			After changing your password, run "inst auth login" with your new password.

			Deployed in %.0fs to %s
			`,
		DefUsername, DefPassword, elapsed.Seconds(), LocalInstancePath)

	err = registerInstance(opts)
	if err != nil {
		return err
	}

	return nil
}

func registerInstance(opts *DeployOptions) error {
	// register the new instance
	exists, err := instance.IsInstanceAdded(opts.Config, "localhost:8080")
	if err != nil {
		return err
	}
	if !exists {
		addOpts := &instance.AddOptions{
			IO:             opts.IO,
			Config:         opts.Config,
			MainExecutable: opts.MainExecutable,
			Interactive:    false,
			InstanceOptions: instance.InstanceOptions{
				APIHostname: "localhost:8080",
				Default:     true,
			},
		}
		p(opts.IO, "")
		err = instance.RunAdd(addOpts)
		if err != nil {
			return err
		}
	}
	return nil
}
