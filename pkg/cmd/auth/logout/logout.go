package logout

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
	"github.com/instill-ai/cli/pkg/prompt"
)

type LogoutOptions struct {
	IO       *iostreams.IOStreams
	Config   func() (config.Config, error)
	Hostname string
}

func NewCmdLogout(f *cmdutil.Factory, runF func(*LogoutOptions) error) *cobra.Command {
	opts := &LogoutOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "logout",
		Args:  cobra.ExactArgs(0),
		Short: "Log out of an Instill Core/Cloud host",
		Long: heredoc.Doc(`Remove authentication for an Instill Core/Cloud host.

			This command removes the authentication configuration for a host either specified
			interactively or via --hostname.
		`),
		Example: heredoc.Doc(`
			$ inst auth logout
			# => select what host to log out of via a prompt

			$ inst auth logout --hostname api.instill.tech
			# => log out of specified host
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Hostname == "" && !opts.IO.CanPrompt() {
				return cmdutil.FlagErrorf("--hostname required when not running interactively")
			}

			if runF != nil {
				return runF(opts)
			}

			return logoutRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "", "The hostname of the Instill instance to log out of")

	return cmd
}

func logoutRun(opts *LogoutOptions) error {
	hostname := opts.Hostname

	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	candidates, err := cfg.Hosts()
	// TODO filter candidates with oauth2 hostname only
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("not logged in to any hosts")
	}

	if hostname == "" {
		if len(candidates) == 1 {
			hostname = candidates[0]
		} else {
			err = prompt.SurveyAskOne(&survey.Select{
				Message: "Which account do you want to log out of?",
				Options: candidates,
			}, &hostname)

			if err != nil {
				return fmt.Errorf("could not prompt: %w", err)
			}
		}
	} else {
		var found bool
		for _, c := range candidates {
			if c == hostname {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("not logged into %s", hostname)
		}
	}

	if err := cfg.CheckWriteable(hostname, ""); err != nil {
		return err
	}

	if opts.IO.CanPrompt() {
		var keepGoing bool
		err := prompt.SurveyAskOne(&survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to log out of %s?", hostname),
			Default: true,
		}, &keepGoing)
		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}

		if !keepGoing {
			return nil
		}
	}

	// TODO invalidate the token instead of removing the whole instance
	cfg.UnsetHost(hostname)
	err = cfg.Write()
	if err != nil {
		return fmt.Errorf("failed to write config, authentication configuration not updated: %w", err)
	}

	isTTY := opts.IO.IsStdinTTY() && opts.IO.IsStdoutTTY()

	if isTTY {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut, "%s Logged out of %s\n",
			cs.SuccessIcon(), cs.Bold(hostname))
	}

	return nil
}
