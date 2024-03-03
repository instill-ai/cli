package local

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

// StatusOptions contains the command line options
type StatusOptions struct {
	IO             *iostreams.IOStreams
	Exec           ExecDep
	OS             OSDep
	Config         config.Config
	MainExecutable string
	Interactive    bool
}

// NewStatusCmd creates a new command
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

	return cmd
}

// TODO separate health statuses per API
func runStatus(opts *StatusOptions) error {

	if _, err := os.Stat(LocalInstancePath); os.IsNotExist(err) {
		p(opts.IO, "")
		p(opts.IO, "Instill Core instance not deployed")
		return nil
	}

	deployed := "NO"
	started := "NO"
	healthy := "NO"
	if err := isDeployed(opts.Exec); err == nil {
		deployed = "YES"
	}
	if err := isStarted(opts.Exec); err == nil {
		started = "YES"
	}
	if err := isHealthy(opts.Exec); err == nil {
		healthy = "YES"
	}
	fmt.Printf("Instill Core - Deployed: %s | Started: %s | Healthy: %s\n", deployed, started, healthy)

	return nil
}

func isDeployed(execDep ExecDep) error {
	if _, err := execCmd(execDep, "bash", "-c", "docker compose ls -a | grep instill-core"); err != nil {
		return err
	}
	return nil
}

func isStarted(execDep ExecDep) error {
	if _, err := execCmd(execDep, "bash", "-c", "docker compose ls -a --format json --filter name=instill-core | grep running"); err != nil {
		return err
	}
	return nil
}

func isHealthy(execDep ExecDep) error {

	urls := []string{
		"localhost:8080/core/v1beta/health/mgmt",
		"localhost:8080/vdp/v1alpha/health/pipeline",
		"localhost:8080/model/v1alpha/health/model",
	}

	for _, url := range urls {
		if out, err := execCmd(execDep, "curl", url); err != nil {
			return err
		} else if !strings.Contains(out, "SERVING_STATUS_SERVING") {
			return fmt.Errorf("ERROR: %s is not healthy", url)
		}
	}

	return nil
}
