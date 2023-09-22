package login

import (
	"fmt"

	"github.com/instill-ai/cli/internal/oauth2"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
	"github.com/instill-ai/cli/pkg/prompt"
)

type LoginOptions struct {
	IO             *iostreams.IOStreams
	Config         func() (config.Config, error)
	MainExecutable string
	Interactive    bool
	Hostname       string
}

func NewCmdLogin(f *cmdutil.Factory, runF func(*LoginOptions) error) *cobra.Command {
	opts := &LoginOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "login",
		Args:  cobra.ExactArgs(0),
		Short: "Authenticate with an Instill host",
		Long: heredoc.Docf(`
			Authenticate with an Instill host.

			The default authentication mode is an authorization code flow.
		`),
		Example: heredoc.Doc(`
			# start login
			$ instill auth login
		`),
		RunE: func(cmd *cobra.Command, args []string) error {

			if opts.IO.CanPrompt() {
				opts.Interactive = true
			}

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return loginRun(f, opts)
		},
	}

	// TODO handle err
	cfg, _ := opts.Config()

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", cfg.DefaultHostname(), "Hostname of an already added Instill AI instance")

	return cmd
}

func loginRun(f *cmdutil.Factory, opts *LoginOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	hostname := opts.Hostname

	hosts, err := cfg.HostsTyped()
	if err != nil {
		return err
	}

	var host *config.HostConfigTyped
	for _, h := range hosts {
		if h.APIHostname == hostname {
			host = &h
			break
		}
	}
	if host == nil {
		return fmt.Errorf("ERROR: instance '%s' does not exists", hostname)
	}

	if host.Oauth2Hostname == "" || host.Oauth2ClientID == "" || host.Oauth2ClientSecret == "" {
		e := heredoc.Docf(`ERROR: OAuth2 config isn't complete for '%s'

			You can fix it with:
			$ instill instances edit %s \
				--oauth2 HOSTNAME \
				--client-id CLIENT_ID \
				--secret SECRET`, hostname, hostname)
		return fmt.Errorf(e)
	}

	if host.RefreshToken != "" && opts.Interactive {
		var keepGoing bool
		err = prompt.SurveyAskOne(&survey.Confirm{
			Message: fmt.Sprintf(
				"You're already logged into %s. Do you want to re-authenticate?",
				hostname),
			Default: false,
		}, &keepGoing)
		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}
		if !keepGoing {
			return nil
		}
	}

	return oauth2.AuthCodeFlowWithConfig(f, host, cfg, opts.IO)
}
