package auth

import (
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/pkg/cmdutil"

	authLoginCmd "github.com/instill-ai/cli/pkg/cmd/auth/login"
	authLogoutCmd "github.com/instill-ai/cli/pkg/cmd/auth/logout"
	authStatusCmd "github.com/instill-ai/cli/pkg/cmd/auth/status"
)

// NewCmdAuth creates the `auth` command
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Login and logout",
		Long:  `Manage authentication state.`,
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(authLoginCmd.NewCmdLogin(f, nil))
	cmd.AddCommand(authLogoutCmd.NewCmdLogout(f, nil))
	cmd.AddCommand(authStatusCmd.NewCmdStatus(f, nil))

	return cmd
}
