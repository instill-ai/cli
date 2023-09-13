package instances

import (
	"github.com/instill-ai/cli/internal/config"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/pkg/cmdutil"
)

type InstanceOptions = struct {
	APIHostname string
	Oauth2      string
	Issuer      string
	Audience    string
	ClientID    string
	Secret      string
	Default     bool
	APIVersion  string
}

func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances <command>",
		Short: "Instances management",
		Long:  `Manage instances of Instill AI, both Cloud and Core.`,
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(NewAddCmd(f, nil))
	cmd.AddCommand(NewEditCmd(f, nil))
	cmd.AddCommand(NewListCmd(f, nil))
	cmd.AddCommand(NewRemoveCmd(f, nil))

	return cmd
}

// AddInstanceFlags adds common instances parameters, shared between commands.
func AddInstanceFlags(cmd *cobra.Command, opts *InstanceOptions) {
	defs := config.DefaultHostConfig()
	cmd.Flags().StringVarP(&opts.APIVersion, "api-version", "a", defs.APIVersion, "API version")
	cmd.Flags().BoolVar(&opts.Default, "default", defs.IsDefault, "Make this the default instance")
	// oauth2 stuff
	cmd.Flags().StringVarP(&opts.Oauth2, "oauth2", "", "", "OAuth2 hostname (optional)")
	cmd.Flags().StringVarP(&opts.Audience, "audience", "", "", "OAuth2 audience (optional)")
	cmd.Flags().StringVarP(&opts.Issuer, "issuer", "", "", "OAuth2 issuer (optional)")
	// TODO get these via a prompt to avoid the shell history?
	cmd.Flags().StringVarP(&opts.ClientID, "client-id", "", "", "OAuth2 client ID (optional)")
	cmd.Flags().StringVarP(&opts.Secret, "secret", "", "", "OAuth2 client secret (optional)")
}
