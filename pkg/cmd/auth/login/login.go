package login

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/pkg/cmd/auth/shared"
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
			Authenticate with am Instill host.

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

			if cmd.Flags().Changed("hostname") {
				if err := instance.HostnameValidator(opts.Hostname); err != nil {
					return cmdutil.FlagErrorf("error parsing --hostname: %w", err)
				}
			}

			if !opts.Interactive {
				if opts.Hostname == "" {
					opts.Hostname = instance.Default()
				}
			}

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return loginRun(f, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "", "The hostname of the Instill instance to authenticate with")

	return cmd
}

func loginRun(f *cmdutil.Factory, opts *LoginOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	hostname := opts.Hostname

	if err := cfg.CheckWriteable(hostname, ""); err != nil {
		return err
	}

	existingRefreshToken, _ := cfg.Get(hostname, "refresh_token")
	if existingRefreshToken != "" && opts.Interactive {
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

	return shared.Login(f, &shared.LoginOptions{
		IO:          opts.IO,
		Config:      cfg,
		Hostname:    opts.Hostname,
		Interactive: opts.Interactive,
		Executable:  opts.MainExecutable,
	})
}
