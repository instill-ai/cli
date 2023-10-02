package local

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
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

var gitDescribeSuffixRE = regexp.MustCompile(`\d+-\d+-g[a-f0-9]{8}$`)

// releaseInfo stores information about a release
type releaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

type StateEntry struct {
	CheckedForUpdateAt time.Time   `yaml:"checked_for_update_at"`
	LatestRelease      releaseInfo `yaml:"latest_release"`
}

const (
	// ConfigKeyPath is the config key for the local instance path where Instill Core is installed
	ConfigKeyPath = "local-instance-path"
)

var projs = [3]string{"Base", "VDP", "Model"}

var logger *slog.Logger
var p = cmdutil.P

func init() {
	var lvl = new(slog.LevelVar)
	if os.Getenv("DEBUG") != "" {
		lvl.Set(slog.LevelDebug)
	} else {
		lvl.Set(slog.LevelError + 1)
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
}

// New creates a new command
func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local <command>",
		Short: "Local Instill Core instance",
		Long:  `Create and manage a local Instill Core instance with ease.`,
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

// getConfigPath returns a configured path to the local instance.
func getConfigPath(cfg config.Config) (string, error) {
	path, err := cfg.Get("", ConfigKeyPath)
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", fmt.Errorf("config %s is empty", ConfigKeyPath)
	}
	return path, nil
}
