package instances

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/pkg/cmdutil"
)

type InstanceOptions = struct {
	APIHostname string
	Oauth2      string
	Issuer      string
	Audience    string
	Default     bool
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
	cmd.Flags().StringVarP(&opts.Oauth2, "oauth2", "o", "", "OAuth2 hostname (optional)")
	cmd.Flags().StringVarP(&opts.Audience, "audience", "a", "", "OAuth2 audience (optional)")
	cmd.Flags().StringVarP(&opts.Issuer, "issuer", "i", "", "OAuth2 issuer (optional)")
	cmd.Flags().BoolVar(&opts.Default, "default", false, "Make this the default instance")
}

// TODO move to utils
func p(txt string, args ...interface{}) {
	fmt.Print(heredoc.Docf(txt, args...))
}
