package instance

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type SetDefaultOptions struct {
	IO             *iostreams.IOStreams
	Config         config.Config
	MainExecutable string
	Interactive    bool
	APIHostname    string
}

func NewSetDefaultCmd(f *cmdutil.Factory, runF func(*SetDefaultOptions) error) *cobra.Command {
	opts := &SetDefaultOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use: "set-default",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return fmt.Errorf("ERROR: specify an API hostname\n$ inst instance set-default API_HOSTNAME")
			}
			if err := instance.HostnameValidator(args[0]); err != nil {
				return fmt.Errorf("error parsing API hostname %w", err)
			}
			return nil
		},
		Short: "Mark an instance as the default one",
		Long: heredoc.Docf(`
			Mark an instance as the default one for commands like "auth" and "api".
			Can be optionally overloaded with the --hostname param.
		`),
		Example: heredoc.Doc(`
			# make instill.localhost the default instance
			$ inst instance set-default instill.localhost
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			opts.Config = cfg

			if opts.IO.CanPrompt() {
				opts.Interactive = true
			}
			opts.APIHostname = args[0]

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return runSetDefault(opts)
		},
	}

	return cmd
}

func runSetDefault(opts *SetDefaultOptions) error {
	hosts, err := opts.Config.HostsTyped()
	if err != nil {
		return err
	}

	apiHost := opts.APIHostname
	var host *config.HostConfigTyped
	for _, h := range hosts {
		if h.APIHostname == apiHost {
			host = &h
			break
		}
	}
	if host == nil {
		return fmt.Errorf("ERROR: instance '%s' does not exists", apiHost)
	}

	host.IsDefault = true

	err = opts.Config.SaveTyped(host)
	if err != nil {
		return fmt.Errorf("ERROR: failed to set instance '%s' as the default one:\n%w", opts.APIHostname, err)
	}

	cmdutil.P(opts.IO, "Instance '%s' has been set as the default one\n", host.APIHostname)

	return nil
}
