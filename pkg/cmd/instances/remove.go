package instances

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type RemoveOptions struct {
	IO             *iostreams.IOStreams
	Config         config.Config
	MainExecutable string
	Interactive    bool
	APIHostname    string
}

func NewRemoveCmd(f *cmdutil.Factory, runF func(*RemoveOptions) error) *cobra.Command {
	opts := &RemoveOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use: "remove",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return fmt.Errorf("ERROR: specify an API hostname\n$ inst instances remove API_HOSTNAME")
			}
			if err := instance.HostnameValidator(args[0]); err != nil {
				return fmt.Errorf("error parsing API hostname %w", err)
			}
			return nil
		},
		Short: "Remove an existing instance",
		Long: heredoc.Docf(`
			Remove an existing Instill AI instance, either Cloud or Core.
		`),
		Example: heredoc.Doc(`
			# remove the local instance instance
			$ inst instances remove instill.localhost
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

			return runRemove(opts)
		},
	}

	return cmd
}

func runRemove(opts *RemoveOptions) error {
	hosts, err := opts.Config.HostsTyped()
	if err != nil {
		return err
	}

	// TODO use cfg.HostnameExists()
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

	opts.Config.UnsetHost(opts.APIHostname)
	err = opts.Config.Write()
	if err != nil {
		return fmt.Errorf("error removing hostname '%s' - %w", opts.APIHostname, err)
	}

	cmdutil.P(opts.IO, "Instance '%s' has been removed\n", opts.APIHostname)

	return nil
}
