package list

import (
	"bytes"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/glamour"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"io"
	"strings"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type Options struct {
	IO             *iostreams.IOStreams
	Config         func() (config.Config, error)
	MainExecutable string
	Interactive    bool
	Hostname       string
}

func New(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
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
			$ instill instances list
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

			return run(f, opts)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	hosts, err := cfg.Hosts()
	if err != nil {
		return err
	}
	cols := []string{"Type", "API Hostname", "Oauth2 Hostname", "Issuer"}
	var data [][]string
	for _, host := range hosts {
		// TODO
		hostType := "Cloud"
		api, err := cfg.Get(host, "api_hostname")
		if err != nil {
			return err
		}
		oauth, err := cfg.Get(host, "oauth2_hostname")
		if err != nil {
			return err
		}
		issuer, err := cfg.Get(host, "oauth2_issuer")
		if err != nil {
			return err
		}
		data = append(data, []string{hostType, host, api, oauth, issuer})
	}

	txt := heredoc.Doc(`# Hello World

		This is a simple example of Markdown rendering with Glamour!

		{{table}}

		Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.
		
		Bye!
	`)

	txt = strings.Replace(txt, "{{table}}", genTable(cols, data), 1)

	out, err := glamour.Render(txt, "dark")
	if err != nil {
		return err
	}

	fmt.Print(out)

	return nil
}

// genTable generates a markdown table as a string.
// TODO move to shared
func genTable(columns []string, data [][]string) string {
	var buf bytes.Buffer
	writer := io.Writer(&buf)
	table := tablewriter.NewWriter(writer)
	table.SetHeader(columns)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()
	return buf.String()
}
