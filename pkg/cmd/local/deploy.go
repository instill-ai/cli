package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/safeexec"
	"github.com/go-playground/validator/v10"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/build"
	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmd/factory"
	"github.com/instill-ai/cli/pkg/cmd/instances"
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
	Path           string `validate:"required,dirpath" example:"/home/instill-core/"`
	checkUpdate    func(ExecDep, string, string) (*releaseInfo, error)
	isDeployed     func(ExecDep) error
}

// NewDeployCmd creates a new command
func NewDeployCmd(f *cmdutil.Factory, runF func(*DeployOptions) error) *cobra.Command {
	opts := &DeployOptions{
		IO:          f.IOStreams,
		checkUpdate: checkForUpdate,
		isDeployed:  isDeployed,
	}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a local Instill Core instance",
		Example: heredoc.Doc(`
			# deploy to /home/me/instill
			$ inst local deploy --path /home/me/instill
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
	d, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Couldn't get Home directory", err)
	}
	dir := filepath.Join(d, ".local", "instill") + string(os.PathSeparator)
	cmd.Flags().StringVarP(&opts.Path, "path", "p", dir, "Destination directory")

	return cmd
}

func runDeploy(opts *DeployOptions) error {

	var err error
	path := opts.Path
	start := time.Now()
	err = validator.New().Struct(opts)
	if err != nil {
		return fmt.Errorf("ERROR: wrong input, %w", err)
	}

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

	// check existing local deployment
	if err := opts.isDeployed(opts.Exec); err == nil {
		p(opts.IO, "A local Instill Core deployment detected")
		for i := range projs {
			prj := strings.ToLower(projs[i])
			if opts.OS != nil {
				err = opts.OS.Chdir(filepath.Join(path, prj))
			} else {
				err = os.Chdir(filepath.Join(path, prj))
			}
			if err != nil {
				return fmt.Errorf("ERROR: can't open the destination, %w", err)
			}

			if currentVersion, err := execCmd(opts.Exec, "bash", "-c", "git name-rev --tags --name-only $(git rev-parse HEAD)"); err == nil {
				newRelease, _ := opts.checkUpdate(opts.Exec, prj, currentVersion)
				if newRelease != nil {
					cmdFactory := factory.New(build.Version)
					stderr := cmdFactory.IOStreams.ErrOut
					fmt.Fprintf(stderr, "\n\n%s %s → %s\n",
						ansi.Color(fmt.Sprintf("A new release of Instill %s is available:", prj), "yellow"),
						ansi.Color(currentVersion, "cyan"),
						ansi.Color(newRelease.Version, "cyan"))
					fmt.Fprintf(stderr, "%s\n\n",
						ansi.Color(newRelease.URL, "yellow"))
				}
			}
		}
		return nil
	}

	p(opts.IO, "Download the latest Instill Core to: %s", path)
	for i := range projs {
		prj := projs[i]
		_, err = os.Stat(filepath.Join(path, prj))
		if os.IsNotExist(err) {
			if latestVersion, err := execCmd(opts.Exec, "bash", "-c", fmt.Sprintf("curl https://api.github.com/repos/instill-ai/%s/releases | jq -r 'map(select(.prerelease)) | first | .tag_name'", strings.ToLower(prj))); err == nil {
				latestVersion = strings.Trim(latestVersion, "\n")
				if out, err := execCmd(opts.Exec, "bash", "-c",
					fmt.Sprintf("git clone --depth 1 -b %s -c advice.detachedHead=false https://github.com/instill-ai/%s.git %s", latestVersion, prj, filepath.Join(path, prj))); err != nil {
					return fmt.Errorf("ERROR: cant clone %s, %w:\n%s", prj, err, out)
				}
				_, _ = checkForUpdate(opts.Exec, prj, latestVersion)
			} else {
				return fmt.Errorf("ERROR: cant find latest release version of %s, %w:\n%s", prj, err, latestVersion)
			}
		}
	}

	p(opts.IO, "Launch Instill Core")
	for i := range projs {
		prj := strings.ToLower(projs[i])
		if opts.OS != nil {
			err = opts.OS.Chdir(filepath.Join(path, prj))
		} else {
			err = os.Chdir(filepath.Join(path, prj))
		}
		if err != nil {
			return fmt.Errorf("ERROR: can't open the destination, %w", err)
		}

		if currentVersion, err := execCmd(opts.Exec, "bash", "-c", "git name-rev --tags --name-only $(git rev-parse HEAD)"); err == nil {
			currentVersion = strings.Trim(currentVersion, "\n")
			p(opts.IO, "Spin up Instill %s %s...", projs[i], currentVersion)
			if out, err := execCmd(opts.Exec, "bash", "-c", "make all"); err != nil {
				return fmt.Errorf("ERROR: %s spin-up failed, %w\n%s", prj, err, out)
			}
		} else {
			return fmt.Errorf("ERROR: cant get current tag %s, %w:\n%s", prj, err, currentVersion)
		}
	}

	// print a summary
	elapsed := time.Since(start)
	p(opts.IO, "")
	p(opts.IO, `
		Instill Core console available at http://localhost:3000
		After changing your password, run "$ inst auth login".

		User:     %s
		Password: %s

		Deployed in %.0fs to %s
		`,
		DefUsername, DefPassword, elapsed.Seconds(), path)

	err = registerInstance(opts)
	if err != nil {
		return err
	}

	return nil
}

func registerInstance(opts *DeployOptions) error {
	// register the new instance
	err := opts.Config.Set("", ConfigKeyPath, opts.Path)
	if err != nil {
		return fmt.Errorf("ERROR: saving config, %w", err)
	}
	err = opts.Config.Write()
	if err != nil {
		return fmt.Errorf("ERROR: saving config, %w", err)
	}
	exists, err := instances.IsInstanceAdded(opts.Config, "localhost:8080")
	if err != nil {
		return err
	}
	if !exists {
		addOpts := &instances.AddOptions{
			IO:             opts.IO,
			Config:         opts.Config,
			MainExecutable: opts.MainExecutable,
			Interactive:    false,
			InstanceOptions: instances.InstanceOptions{
				APIHostname: "localhost:8080",
				Default:     true,
			},
		}
		p(opts.IO, "")
		err = instances.RunAdd(addOpts)
		if err != nil {
			return err
		}
	}
	return nil
}
