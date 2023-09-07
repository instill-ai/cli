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
	APIHostname    string
	Oauth2         string
	Issuer         string
	Audience       string
	Default        bool
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
				return fmt.Errorf("Error: specify an API hostname\n$ inst instances add API_HOSTNAME")
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

			return runAdd(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Oauth2, "oauth2", "o", "", "OAuth2 hostname (optional)")
	cmd.Flags().StringVarP(&opts.Audience, "audience", "a", "", "OAuth2 audience (optional)")
	cmd.Flags().StringVarP(&opts.Issuer, "issuer", "i", "", "OAuth2 issuer (optional)")
	cmd.Flags().BoolVar(&opts.Default, "default", false, "Make this the default instance")

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

	// TODO check exists
	apiHost := opts.APIHostname
	for _, h := range hosts {
		if h.APIHostname == apiHost {
			return fmt.Errorf("apiHost %s already exists", apiHost)
		}
	}

	p(`Saving config for:
		%s
		%s
		%s
		%s
	`, apiHost, opts.Issuer, opts.Audience, opts.Oauth2)

	return nil
}

// TODO move to shared
func p(txt string, args ...interface{}) {
	fmt.Print(heredoc.Docf(txt, args...))
}
