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

type EditOptions struct {
	IO             *iostreams.IOStreams
	Config         func() (config.Config, error)
	MainExecutable string
	Interactive    bool
	InstanceOptions
}

func NewEditCmd(f *cmdutil.Factory, runF func(*EditOptions) error) *cobra.Command {
	opts := &EditOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use: "edit",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return fmt.Errorf("Error: specify an API hostname\n$ inst instances edit API_HOSTNAME")
			}
			if err := instance.HostnameValidator(args[0]); err != nil {
				return fmt.Errorf("error parsing API hostname %w", err)
			}
			return nil
		},
		Short: "Edit an existing instance",
		Long: heredoc.Docf(`
			Edit an existing Instill AI instance, either Cloud or Core.
		`),
		Example: heredoc.Doc(`
			# make instill.localhost the default instance
			$ inst instances edit instill.localhost --default

			# update the issuer for api.instill.tech
			$ inst instances edit api.instill.tech \
				--issuer https://auth.instill.tech/
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

			return runEdit(opts)
		},
	}

	AddInstanceFlags(cmd, &opts.InstanceOptions)

	return cmd
}

func runEdit(opts *EditOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	hosts, err := cfg.HostsTyped()
	if err != nil {
		return err
	}

	apiHost := opts.APIHostname
	var host *config.HostConfigTyped
	for _, h := range hosts {
		if h.APIHostname == apiHost {
			host = &h
		}
	}
	if host == nil {
		return fmt.Errorf("ERROR: instance '%s' does not exists", apiHost)
	}

	if opts.Oauth2 != "" && (opts.Secret == "" || opts.ClientID == "") {
		return fmt.Errorf("ERROR: --secret and --client-id required when --oauth2 is specified")
	}

	host.APIHostname = opts.APIHostname
	host.IsDefault = opts.Default
	host.Oauth2Hostname = opts.Oauth2
	host.Oauth2Audience = opts.Audience
	host.Oauth2Issuer = opts.Issuer
	host.Oauth2Secret = opts.Secret
	host.Oauth2ClientID = opts.ClientID
	host.APIVersion = opts.APIVersion

	err = cfg.SaveTyped(host)
	if err != nil {
		return fmt.Errorf("ERROR: failed to edit instance '%s':\n%w", opts.APIHostname, err)
	}

	cmdutil.P("Instance '%s' has been saved", host.APIHostname)

	return nil
}
