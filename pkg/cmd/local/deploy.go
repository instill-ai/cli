package local

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
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
	cmd.Flags().BoolVarP(&opts.Upgrade, "upgrade", "u", false, "Upgrade Instill Core instance to the latest version")
	cmd.MarkFlagsMutuallyExclusive("force", "upgrade")

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
	for _, proj := range projs {
		projDirPath := filepath.Join(LocalInstancePath, proj)
		_, err = os.Stat(projDirPath)
		if os.IsNotExist(err) {
			if latestVersion, err := execCmd(opts.Exec, "bash", "-c", fmt.Sprintf("curl https://api.github.com/repos/instill-ai/%s/releases | jq -r 'map(select(.prerelease)) | first | .tag_name'", proj)); err == nil {
				latestVersion = strings.Trim(latestVersion, "\n")
				if out, err := execCmd(opts.Exec, "bash", "-c",
					fmt.Sprintf("git clone --depth 1 -b %s -c advice.detachedHead=false https://github.com/instill-ai/%s.git %s", latestVersion, proj, projDirPath)); err != nil {
					return fmt.Errorf("ERROR: cannot clone %s, %w:\n%s", proj, err, out)
				}
				if _, err := opts.checkForUpdate(opts.Exec, filepath.Join(config.StateDir(), fmt.Sprintf("%s.yml", proj)), fmt.Sprintf("instill-ai/%s", proj), latestVersion); err != nil {
					return fmt.Errorf("ERROR: cannot check for update %s, %w:\n%s", proj, err, latestVersion)
				}
			} else {
				return fmt.Errorf("ERROR: cannot find latest release version of %s, %w:\n%s", proj, err, latestVersion)
			}
		}
	}

	if opts.Force {
		p(opts.IO, "Tear down Instill Core instance if existing...")
		for _, proj := range projs {
			projDirPath := filepath.Join(LocalInstancePath, proj)
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
					return fmt.Errorf("ERROR: cannot force tearing down %s, %w:\n%s", proj, err, out)
				}
			}
		}
	} else if opts.Upgrade {
		hasNewVersion := false
		for _, proj := range projs {
			projDirPath := filepath.Join(LocalInstancePath, proj)
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
					if newRelease, err := opts.checkForUpdate(opts.Exec, filepath.Join(config.StateDir(), fmt.Sprintf("%s.yml", proj)), fmt.Sprintf("instill-ai/%s", proj), currentVersion); err != nil {
						return fmt.Errorf("ERROR: cannot check for update %s, %w:\n%s", proj, err, currentVersion)
					} else if newRelease != nil {
						p(opts.IO, "Upgrade %s to %s...", proj, newRelease.Version)
						hasNewVersion = true
					}
				} else {
					return fmt.Errorf("ERROR: cannot find current release version of %s, %w:\n%s", proj, err, currentVersion)
				}
			}
		}

		if hasNewVersion {
			p(opts.IO, "Tear down Instill Core instance if existing...")
			for _, proj := range projs {
				projDirPath := filepath.Join(LocalInstancePath, proj)
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
						return fmt.Errorf("ERROR: cannot force tearing down %s, %w:\n%s", proj, err, out)
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

				if latestVersion, err := execCmd(opts.Exec, "bash", "-c", fmt.Sprintf("curl https://api.github.com/repos/instill-ai/%s/releases | jq -r 'map(select(.prerelease)) | first | .tag_name'", proj)); err == nil {
					latestVersion = strings.Trim(latestVersion, "\n")
					if out, err := execCmd(opts.Exec, "bash", "-c",
						fmt.Sprintf("git clone --depth 1 -b %s -c advice.detachedHead=false https://github.com/instill-ai/%s.git %s", latestVersion, proj, projDirPath)); err != nil {
						return fmt.Errorf("ERROR: cannot clone %s, %w:\n%s", proj, err, out)
					}
				} else {
					return fmt.Errorf("ERROR: cannot find latest release version of %s, %w:\n%s", proj, err, latestVersion)
				}
			}
		} else {
			p(opts.IO, "No upgrade available")
			return nil
		}

	} else {
		for _, proj := range projs {
			if err := opts.isDeployed(opts.Exec, proj); err == nil {
				p(opts.IO, "A local Instill Core deployment detected")
				return nil
			}
		}
	}

	p(opts.IO, "Launch Instill Core...")
	for _, proj := range projs {
		projDirPath := filepath.Join(LocalInstancePath, proj)
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

		if currentVersion, err := execCmd(opts.Exec, "bash", "-c", "git name-rev --tags --name-only $(git rev-parse HEAD)"); err == nil {
			currentVersion = strings.Trim(currentVersion, "\n")
			p(opts.IO, "Spin up %s %s...", proj, currentVersion)
			if out, err := execCmd(opts.Exec, "bash", "-c", "make all"); err != nil {
				return fmt.Errorf("ERROR: %s spin-up failed, %w\n%s", proj, err, out)
			}
		} else {
			return fmt.Errorf("ERROR: cannot get current tag %s, %w:\n%s", proj, err, currentVersion)
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
