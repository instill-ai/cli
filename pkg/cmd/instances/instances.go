package auth

import (
	"github.com/spf13/cobra"

	listInsCmd "github.com/instill-ai/cli/pkg/cmd/instances/list"
	"github.com/instill-ai/cli/pkg/cmdutil"
)

func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances <command>",
		Short: "Instances management",
		Long:  `Manage instill's instances.`,
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(listInsCmd.New(f, nil))

	return cmd
}
