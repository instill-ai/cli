package instances

import (
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/pkg/cmdutil"
)

func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local <command>",
		Short: "Local Instill Core instance",
		Long:  `Create and manage a local Instill Core instance with ease.`,
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(NewDeployCmd(f, nil))

	return cmd
}
