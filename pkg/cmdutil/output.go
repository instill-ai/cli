package cmdutil

import (
	"bytes"
	"fmt"
	"io"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/glamour"
	"github.com/olekukonko/tablewriter"

	"github.com/instill-ai/cli/pkg/iostreams"
)

func PrintMarkdown(io *iostreams.IOStreams, md string) error {
	tr, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)
	out, err := tr.Render(md)
	if err != nil {
		return err
	}
	fmt.Fprint(io.Out, out)
	return nil
}

// GenTable generates a Markdown table as a string.
func GenTable(columns []string, data [][]string) string {
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

// P is a shorthand for print and heredoc.
func P(io *iostreams.IOStreams, txt string, args ...interface{}) {
	_, _ = fmt.Fprint(io.Out, heredoc.Docf(txt, args...))
}
