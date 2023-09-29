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
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmd/instances"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

const (
	DefUsername = "admin"
	DefPassword = "password"
)

type DeployOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	OS             OSDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
	Path           string `validate:"required,dirpath" example:"/home/instill-core/"`
	Branch         string `validate:"required" example:"main"`
}

func NewDeployCmd(f *cmdutil.Factory, runF func(*DeployOptions) error) *cobra.Command {
	opts := &DeployOptions{
		IO: f.IOStreams,
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
	dir := filepath.Join(d, ".config", "instill") + string(os.PathSeparator)
	cmd.Flags().StringVarP(&opts.Path, "path", "p", dir, "Destination directory")
	cmd.Flags().StringVarP(&opts.Branch, "branch", "b", "main", "Source branch, used to test new features")

	return cmd
}

func runDeploy(opts *DeployOptions) error {
	// init, validate
	var err error
	path := opts.Path
	io2 := opts.IO
	start := time.Now()
	err = validator.New().Struct(opts)
	if err != nil {
		return fmt.Errorf("ERROR: wrong input, %w", err)
	}

	// check the deps
	apps := []string{"docker", "make", "git"}
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

	comps := []string{"Base", "VDP", "Model"}

	p(io2, "Download Instill Core to: %s", path)
	for _, c := range comps {
		_, err = os.Stat(filepath.Join(path, c))
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			p(io2, "git clone Instill %s...", c)
			out, err := execCmd(opts.Exec,
				"git", "clone", "--depth", "1", "--branch", opts.Branch,
				fmt.Sprintf("https://github.com/instill-ai/%s.git", c), filepath.Join(path, strings.ToLower(c)))
			if err != nil {
				return fmt.Errorf("ERROR: cant clone %s, %w:\n%s", c, err, out)
			}
			if err != nil {
				return fmt.Errorf("ERROR: cant clone %s, %w:\n%s", c, err, out)
			}
		}
	}

	p(io2, "Launch Instill Core")
	for _, c := range comps {
		_, err := execCmd(opts.Exec, "bash", "-c", fmt.Sprintf("docker compose ls | grep instill-%s", strings.ToLower(c)))
		if err != nil {
			if opts.OS != nil {
				err = opts.OS.Chdir(filepath.Join(path, strings.ToLower(c)))
			} else {
				err = os.Chdir(filepath.Join(path, strings.ToLower(c)))
			}
			if err != nil {
				return fmt.Errorf("ERROR: can't open the destination, %w", err)
			}
			p(io2, "Spin up Instill %s...", c)
			// TODO INS-2141 use make all
			//cmd = exec.Command("make", "all")
			out, err := execCmd(opts.Exec, "make", "latest", "PROFILE=all")
			if err != nil {
				return fmt.Errorf("ERROR: %s spin-up failed, %w\n%s", c, err, out)
			}
		}
	}

	// print a summary
	elapsed := time.Since(start)
	p(io2, "")
	p(io2, `
		Instill Core console available under http://localhost:3000
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

	// TODO ask and open browser

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
	exists, err := instances.IsInstanceAdded(opts.Config, "localhost")
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
