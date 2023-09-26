package instances

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type DeployOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
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
	pwd, err := os.Getwd()
	if err != nil {
		logger.Error("Couldn't get pwd", err)
	}
	dir := filepath.Join(filepath.Dir(pwd), "instill-core") + string(os.PathSeparator)
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
		_, err := exec.LookPath(n)
		if err != nil {
			return fmt.Errorf("ERROR: docker not found")
		}
	}

	// init the dir
	p(io2, "Deploying Instill Core to:\n%s", path)
	_, err = os.Stat(path)
	if os.IsExist(err) {
		return fmt.Errorf("ERROR: destination directory already exists")
	}

	// build and run
	p(io2, "GIT clone: in progress")
	out, err := execCmd(opts.Exec,
		"git", "clone", "--depth", "1", "--branch", opts.Branch,
		"https://github.com/instill-ai/vdp.git", path)
	if err != nil {
		return fmt.Errorf("ERROR: cant clone VDP, %w:\n%s", err, out)
	}
	// TODO progress
	p(io2, "GIT clone: done")

	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("ERROR: can't open the destination, %w", err)
	}

	p(io2, "make all: in progress")
	// TODO INS-2141
	//cmd = exec.Command("make", "all")
	out, err = execCmd(opts.Exec, "make", "latest", "PROFILE=all")
	if err != nil {
		return fmt.Errorf("ERROR: make all failed, %w\n%s", err, out)
	}
	// TODO progress
	p(io2, "make all: done")

	err = opts.Config.Set("", "local-instance-path", path)
	if err != nil {
		return fmt.Errorf("ERROR: saving config, %w", err)
	}
	err = opts.Config.Write()
	if err != nil {
		return fmt.Errorf("ERROR: saving config, %w", err)
	}

	// print a summary
	elapsed := time.Since(start)
	p(io2, "")
	p(io2, `
		Instill Core console available under http://localhost:3000

		User:     admin
		Password: password

		Deployed in %.0fs to %s
		`,
		elapsed.Seconds(), path)

	// TODO ask and open browser

	return nil
}
