package local

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type StatusOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	OS             OSDep
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

	deployed := "NO"
	started := "NO"
	healthy := "NO"
	if err := isDeployed(opts.Exec); err == nil {
		deployed = "YES"
	}
	if err := isStarted(opts.Exec); err == nil {
		started = "YES"
	}
	errHealthy := isHealthy(opts.Exec)
	if errHealthy == nil {
		healthy = "YES"
	}

	p(opts.IO, `
		Status of the local Instill Core instance:

		Deployed: %s
		Started: %s
		Healthy: %s
	`, deployed, started, healthy)

	if opts.Verbose && errHealthy != nil {
		p(opts.IO, "")
		p(opts.IO, "Error:\n%s", errHealthy)
	}

	return nil
}

// isDeployed returns no errors if an local instance is detected
func isDeployed(execDep ExecDep) error {

	var checkList = make([]bool, len(projs))
	for i := range checkList {
		proj := strings.ToLower(projs[i])
		if _, err := execCmd(execDep, "bash", "-c", fmt.Sprintf("docker compose ls -a | grep instill-%s", proj)); err == nil {
			checkList[i] = true
		}
	}

	suiteCheck := 0
	for i := range checkList {
		if checkList[i] {
			suiteCheck++
		}
	}

	if suiteCheck == len(checkList) {
		return nil
	}

	return fmt.Errorf("No local Instill Core deployment detected")
}

// isStarted returns no errors if an instance is running.
func isStarted(execDep ExecDep) error {

	if err := isDeployed(execDep); err != nil {
		return err
	}

	for i := range projs {
		proj := strings.ToLower(projs[i])
		if _, err := execCmd(execDep, "bash", "-c", fmt.Sprintf("docker compose ls -a --format json --filter name=instill-%s | grep running", proj)); err != nil {
			return fmt.Errorf("%s is not running", projs[i])
		}
	}

	return nil
}

// isHealthy returns no error if an instance in `path` is responding.
// TODO assert responses
func isHealthy(execDep ExecDep) error {
	if err := isDeployed(execDep); err != nil {
		return err
	}
	if err := isStarted(execDep); err != nil {
		return err
	}
	urls := []string{
		":8080/base/v1alpha/health/mgmt",
		":8080/vdp/v1alpha/health/pipeline",
		":8080/vdp/v1alpha/health/connector",
		":8080/model/v1alpha/health/model",
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
