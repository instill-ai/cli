package cmdutil

import (
	"bytes"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/glamour"
	"github.com/olekukonko/tablewriter"
	"io"
)

func PrintMarkdown(md string) error {
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
func P(txt string, args ...interface{}) {
	fmt.Print(heredoc.Docf(txt, args...))
}
