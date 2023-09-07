package instances

import (
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/pkg/cmdutil"
)

func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances <command>",
		Short: "Instances management",
		Long:  `Manage instances of Instill AI, both Cloud and Core.`,
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(NewAddCmd(f, nil))
	cmd.AddCommand(NewListCmd(f, nil))

	return cmd
}
