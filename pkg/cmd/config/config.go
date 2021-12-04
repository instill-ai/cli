package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	cmdGet "github.com/instill-ai/cli/pkg/cmd/config/get"
	cmdSet "github.com/instill-ai/cli/pkg/cmd/config/set"
	"github.com/instill-ai/cli/pkg/cmdutil"
)

func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	longDoc := strings.Builder{}
	longDoc.WriteString("Display or change configuration settings for instill.\n\n")
	longDoc.WriteString("Current respected settings:\n")
	for _, co := range config.ConfigOptions() {
		longDoc.WriteString(fmt.Sprintf("- %s: %s", co.Key, co.Description))
		if co.DefaultValue != "" {
			longDoc.WriteString(fmt.Sprintf(" (default: %q)", co.DefaultValue))
		}
		longDoc.WriteRune('\n')
	}

	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage configuration for instill",
		Long:  longDoc.String(),
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(cmdGet.NewCmdConfigGet(f, nil))
	cmd.AddCommand(cmdSet.NewCmdConfigSet(f, nil))

	return cmd
}
