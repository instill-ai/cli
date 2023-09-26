package instances

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
)

type ExecDep interface {
	Command(name string, arg ...string) *exec.Cmd
}

const PathConfigKey = "local-instance-path"

// TODO config
// const logLevel = slog.LevelInfo
const logLevel = slog.LevelDebug

var logger *slog.Logger
var p = cmdutil.P

func init() {
	var lvl = new(slog.LevelVar)
	lvl.Set(logLevel)
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
}

// TODO logs
func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local <command>",
		Short: "Local Instill Core instance",
		Long:  `Create and manage a local Instill Core instance with ease.`,
	}

	cmdutil.DisableAuthCheck(cmd)

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

// ConfigPath returns a configured path to the local instance.
func ConfigPath(cfg config.Config) (string, error) {
	path, err := cfg.Get("", PathConfigKey)
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", fmt.Errorf("config %s is empty", PathConfigKey)
	}
	return path, nil
}
