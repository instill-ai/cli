package instances

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type StatusOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
	Verbose        bool
}

func NewStatusCmd(f *cmdutil.Factory, runF func(*StatusOptions) error) *cobra.Command {
	opts := &StatusOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of the local deployment",
		Long: heredoc.Doc(`
			Checks the local deployment for:
			- is it deployed?
			- is it started?
			- is it healthy?
		`),
		Example: heredoc.Doc(`
			# check the local Instill Core
			$ inst local status
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

			return runStatus(opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Show verbose output")

	return cmd
}

// TODO separate health statuses per API
func runStatus(opts *StatusOptions) error {
	// init, validate
	io2 := opts.IO
	exec2 := opts.Exec
	path, err := ConfigPath(opts.Config)
	if err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}
	err = validator.New().Struct(opts)
	if err != nil {
		return fmt.Errorf("ERROR: wrong input, %w", err)
	}

	deployed := "NO"
	started := "NO"
	healthy := "NO"
	if err := IsDeployed(path); err == nil {
		deployed = "YES"
	}
	if err := IsStarted(exec2, path); err == nil {
		started = "YES"
	}
	errHealthy := IsHealthy(exec2, path)
	if errHealthy == nil {
		healthy = "YES"
	}

	p(io2, `
		Status of Instill Core in %s

		Deployed: %s
		Started: %s
		Healthy: %s
	`, path, deployed, started, healthy)

	if opts.Verbose && errHealthy != nil {
		p(io2, "")
		p(io2, "Error:\n%s", errHealthy)
	}

	return nil
}

// IsDeployed returns no errors if an instance in `path` is a clone of VDP.
func IsDeployed(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s doesn't exist", path)
	}
	_, err = os.Stat(filepath.Join(path, "Makefile"))
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s isn't Instill Core", path)
	}
	return nil
}

// IsStarted returns no errors if an instance in `path` is running.
// execDep is used for DI and can be nil.
func IsStarted(execDep ExecDep, path string) error {
	if err := IsDeployed(path); err != nil {
		return err
	}
	err := os.Chdir(path)
	if err != nil {
		return err
	}
	out, err := execCmd(execDep, "make", "top")
	logger.Debug("IsStarted", "out", out)
	if err != nil {
		logger.Error("make top", "err", err.Error())
		return err
	}
	if out == "" {
		return fmt.Errorf("make top empty")
	}
	return nil
}

// IsHealthy returns no error if an instance in `path` is responding.
// execDep is used for DI and can be nil.
// TODO assert responses
func IsHealthy(execDep ExecDep, path string) error {
	if err := IsDeployed(path); err != nil {
		return err
	}
	if err := IsStarted(execDep, path); err != nil {
		return err
	}
	urls := []string{
		":8080/vdp/v1alpha/health/pipeline",
		":8080/base/v1alpha/health/mgmt",
		":8080/vdp/v1alpha/health/connector",
		":3000/",
	}
	for _, url := range urls {
		u := fmt.Sprintf("localhost%s", url)
		out, err := execCmd(execDep, "curl", u)
		logger.Debug("IsHealthy", "url", u, "out", out)
		if err != nil {
			logger.Error("IsHealthy", "url", u, "err", err.Error())
			return err
		}
		if out == "" {
			return fmt.Errorf("cant reach %s", u)
		}
	}
	return nil
}
