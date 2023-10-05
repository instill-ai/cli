package root

import (
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/pkg/cmd/factory"
	"github.com/instill-ai/cli/pkg/cmdutil"

	apiCmd "github.com/instill-ai/cli/pkg/cmd/api"
	authCmd "github.com/instill-ai/cli/pkg/cmd/auth"
	completionCmd "github.com/instill-ai/cli/pkg/cmd/completion"
	configCmd "github.com/instill-ai/cli/pkg/cmd/config"
	instancesCmd "github.com/instill-ai/cli/pkg/cmd/instances"
	localCmd "github.com/instill-ai/cli/pkg/cmd/local"
	versionCmd "github.com/instill-ai/cli/pkg/cmd/version"
)

// NewCmdRoot initiates the Cobra command root
func NewCmdRoot(f *cmdutil.Factory, version, buildDate string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inst <command> <subcommand> [flags]",
		Short: "Instill CLI",
		Long:  `Access Instill Core/Cloud from the command line.`,

		SilenceErrors: true,
		SilenceUsage:  true,
		Example: heredoc.Doc(`
			$ inst api pipelines
			$ inst config get editor
			$ inst auth login
		`),
		Annotations: map[string]string{
			"help:feedback": heredoc.Doc(`
				Please open an issue on https://github.com/instill-ai/community.
			`),
			"help:environment": heredoc.Doc(`
				See 'inst help environment' for the list of supported environment variables.
			`),
		},
	}

	cmd.SetOut(f.IOStreams.Out)
	cmd.SetErr(f.IOStreams.ErrOut)

	cmd.PersistentFlags().Bool("help", false, "Show help for command")
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		rootHelpFunc(f, cmd, args)
	})
	cmd.SetUsageFunc(rootUsageFunc)
	cmd.SetFlagErrorFunc(rootFlagErrorFunc)

	formattedVersion := versionCmd.Format(version, buildDate)
	cmd.SetVersionTemplate(formattedVersion)
	cmd.Version = formattedVersion
	cmd.Flags().Bool("version", false, "Show inst version")

	// Child commands
	cmd.AddCommand(versionCmd.NewCmdVersion(f, version, buildDate))
	cmd.AddCommand(authCmd.NewCmdAuth(f))
	cmd.AddCommand(instancesCmd.New(f))
	cmd.AddCommand(configCmd.NewCmdConfig(f))
	cmd.AddCommand(localCmd.New(f))
	cmd.AddCommand(completionCmd.NewCmdCompletion(f.IOStreams))

	// the `api` command should not inherit any extra HTTP headers
	bareHTTPCmdFactory := *f
	bareHTTPCmdFactory.HTTPClient = bareHTTPClient(f, version)

	cmd.AddCommand(apiCmd.NewCmdAPI(&bareHTTPCmdFactory, nil))

	// Help topics
	cmd.AddCommand(NewHelpTopic("environment"))
	cmd.AddCommand(NewHelpTopic("formatting"))
	cmd.AddCommand(NewHelpTopic("mintty"))
	referenceCmd := NewHelpTopic("reference")
	referenceCmd.SetHelpFunc(referenceHelpFn(f.IOStreams))
	cmd.AddCommand(referenceCmd)

	cmdutil.DisableAuthCheck(cmd)

	// this needs to appear last:
	referenceCmd.Long = referenceLong(cmd)
	return cmd
}

func bareHTTPClient(f *cmdutil.Factory, version string) func() (*http.Client, error) {
	return func() (*http.Client, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		return factory.NewHTTPClient(f.IOStreams, cfg, version, false)
	}
}
