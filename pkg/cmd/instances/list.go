package instances

import (
	"bytes"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/glamour"
	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"io"
)

type ListOptions struct {
	IO             *iostreams.IOStreams
	Config         func() (config.Config, error)
	MainExecutable string
	Interactive    bool
}

func NewListCmd(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
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
			$ inst instances list
		`),
		RunE: func(cmd *cobra.Command, args []string) error {

			if opts.IO.CanPrompt() {
				opts.Interactive = true
			}

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return runListCmd(opts)
		},
	}

	return cmd
}

func runListCmd(opts *ListOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	hosts, err := cfg.HostsTyped()
	if err != nil {
		return err
	}
	cols := []string{"Default", "API Hostname", "Oauth2 Hostname", "Oauth2 Audience", "Oauth2 Issuer"}
	var data [][]string
	for _, h := range hosts {
		def := ""
		if h.IsDefault {
			def = "*"
		}
		row := []string{def, h.APIHostname, h.Oauth2, h.Audience, h.Issuer}
		data = append(data, row)
	}

	md := genTable(cols, data)
	err = printMarkdown(md)
	if err != nil {
		return err
	}

	return nil
}

// TODO move to shared
func printMarkdown(md string) error {
	tr, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)
	out, err := tr.Render(md)
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
