package instances

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type ListOptions struct {
	IO             *iostreams.IOStreams
	Config         config.Config
	MainExecutable string
	Interactive    bool
}

func NewListCmd(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.ExactArgs(0),
		Short: "View added instances",
		Long: heredoc.Docf(`
			View added cloud and local instances.
		`),
		Example: heredoc.Doc(`
			# list instances
			$ inst instances list
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

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return runList(opts)
		},
	}

	return cmd
}

func runList(opts *ListOptions) error {
	hosts, err := opts.Config.HostsTyped()
	if err != nil {
		return err
	}
	cols := []string{"Default", "API Hostname", "Oauth2 Hostname", "Oauth2 Audience", "Oauth2 Issuer", "API Version"}
	var data [][]string
	defHostname := opts.Config.DefaultHostname()
	for _, h := range hosts {
		def := ""
		if h.APIHostname == defHostname {
			def = "*"
		}
		row := []string{def, h.APIHostname, h.Oauth2Hostname, h.Oauth2Audience, h.Oauth2Issuer, h.APIVersion}
		data = append(data, row)
	}

	md := cmdutil.GenTable(cols, data)
	err = cmdutil.PrintMarkdown(opts.IO, md)
	if err != nil {
		return fmt.Errorf("ERROR: failed to list instances: %w", err)
	}

	return nil
}
