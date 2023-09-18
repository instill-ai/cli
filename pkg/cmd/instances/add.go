package instances

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

type AddOptions struct {
	IO             *iostreams.IOStreams
	Config         func() (config.Config, error)
	MainExecutable string
	Interactive    bool
	InstanceOptions
}

func NewAddCmd(f *cmdutil.Factory, runF func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use: "add",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return fmt.Errorf("ERROR: specify an API hostname\n$ inst instances add API_HOSTNAME")
			}
			if err := instance.HostnameValidator(args[0]); err != nil {
				return fmt.Errorf("error parsing API hostname %w", err)
			}
			return nil
		},
		Short: "Add a new instance",
		Long: heredoc.Docf(`
			Add a new Instill AI instance, either Cloud or Core.
		`),
		Example: heredoc.Doc(`
			# add a local instance as the default one
			$ inst instances add instill.localhost --default

			# add a cloud instance
			$ inst instances add api.instill.tech \
				--oauth2 auth.instill.tech \
				--audience https://instill.tech \
				--issuer https://auth.instill.tech/ \
				--secret YOUR_SECRET \
				--client-id CLIENT_ID
		`),
		RunE: func(cmd *cobra.Command, args []string) error {

			if opts.IO.CanPrompt() {
				opts.Interactive = true
			}
			opts.APIHostname = args[0]

			if cmd.Flags().Changed("oauth") {
				if err := instance.HostnameValidator(opts.Oauth2); err != nil {
					return cmdutil.FlagErrorf("error parsing OAuth2 hostname: %w", err)
				}
			}

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return runAdd(opts)
		},
	}

	AddInstanceFlags(cmd, &opts.InstanceOptions)

	return cmd
}

func runAdd(opts *AddOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	hosts, err := cfg.HostsTyped()
	if err != nil {
		return err
	}

	apiHost := opts.APIHostname
	for _, h := range hosts {
		if h.APIHostname == apiHost {
			return fmt.Errorf("ERROR: instance '%s' already exists", apiHost)
		}
	}

	if opts.Oauth2 != "" && (opts.Secret == "" || opts.ClientID == "") {
		return fmt.Errorf("ERROR: --secret and --client-id required when --oauth2 is specified")
	}

	host := config.DefaultHostConfig()
	host.APIHostname = opts.APIHostname
	host.IsDefault = opts.Default
	host.Oauth2Hostname = opts.Oauth2
	host.Oauth2Audience = opts.Audience
	host.Oauth2Issuer = opts.Issuer
	host.Oauth2ClientID = opts.ClientID
	host.Oauth2Secret = opts.Secret
	host.APIVersion = opts.APIVersion

	err = cfg.SaveTyped(&host)
	if err != nil {
		return fmt.Errorf("ERROR: failed to add instance '%s': %w", opts.APIHostname, err)
	}

	cmdutil.P("Instance '%s' has been added", host.APIHostname)

	return nil
}
