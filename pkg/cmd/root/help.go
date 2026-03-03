package root

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/markdown"
	"github.com/instill-ai/cli/pkg/text"
)

func rootUsageFunc(command *cobra.Command) error {
	command.Printf("Usage:  %s", command.UseLine())

	subcommands := command.Commands()
	if len(subcommands) > 0 {
		command.Print("\n\nAvailable commands:\n")
		for _, c := range subcommands {
			if c.Hidden {
				continue
			}
			command.Printf("  %s\n", c.Name())
		}
		return nil
	}

	flagUsages := command.LocalFlags().FlagUsages()
	if flagUsages != "" {
		command.Println("\n\nFlags:")
		command.Print(text.Indent(dedent(flagUsages), "  "))
	}
	return nil
}

func rootFlagErrorFunc(cmd *cobra.Command, err error) error {
	if err == pflag.ErrHelp {
		return err
	}
	return cmdutil.FlagErrorWrap(err)
}

var hasFailed bool

type helpEntry struct {
	Title string
	Body  string
}

// HasFailed signals that the main process should exit with non-zero status
func HasFailed() bool {
	return hasFailed
}

// Display helpful error message in case subcommand name was mistyped.
// This matches Cobra's behavior for root command, which Cobra
// confusingly doesn't apply to nested commands.
func nestedSuggestFunc(command *cobra.Command, arg string) {
	command.Printf("unknown command %q for %q\n", arg, command.CommandPath())

	var candidates []string
	if arg == "help" {
		candidates = []string{"--help"}
	} else {
		if command.SuggestionsMinimumDistance <= 0 {
			command.SuggestionsMinimumDistance = 2
		}
		candidates = command.SuggestionsFor(arg)
	}

	if len(candidates) > 0 {
		command.Print("\nDid you mean this?\n")
		for _, c := range candidates {
			command.Printf("\t%s\n", c)
		}
	}

	command.Print("\n")
	_ = rootUsageFunc(command)
}

func isRootCmd(command *cobra.Command) bool {
	return command != nil && !command.HasParent()
}

func rootHelpFunc(f *cmdutil.Factory, command *cobra.Command, args []string) {
	io := f.IOStreams

	if isRootCmd(command.Parent()) && len(args) >= 2 && args[1] != "--help" && args[1] != "-h" {
		nestedSuggestFunc(command, args[1])
		hasFailed = true
		return
	}

	coreCommands := []string{}
	actionsCommands := []string{}
	additionalCommands := []string{}
	for _, c := range command.Commands() {
		if c.Short == "" {
			continue
		}
		if c.Hidden {
			continue
		}

		s := rpad(c.Name()+":", c.NamePadding()) + c.Short
		if _, ok := c.Annotations["IsCore"]; ok {
			coreCommands = append(coreCommands, s)
		} else if _, ok := c.Annotations["IsActions"]; ok {
			actionsCommands = append(actionsCommands, s)
		} else {
			additionalCommands = append(additionalCommands, s)
		}
	}

	// If there are no core commands, assume everything is a core command
	if len(coreCommands) == 0 {
		coreCommands = additionalCommands
		additionalCommands = []string{}
	}

	type helpEntry struct {
		Title string
		Body  string
	}

	longText := command.Long
	if longText == "" {
		longText = command.Short
	}
	if longText != "" && command.LocalFlags().Lookup("jq") != nil {
		longText = strings.TrimRight(longText, "\n") +
			"\n\nFor more information about output formatting flags, see `inst help formatting`."
	}

	entries := []helpEntry{}
	if longText != "" {
		entries = append(entries, helpEntry{"", longText})
	}
	entries = append(entries, helpEntry{"USAGE", command.UseLine()})
	if len(coreCommands) > 0 {
		entries = append(entries, helpEntry{"CORE COMMANDS", strings.Join(coreCommands, "\n")})
	}
	if len(actionsCommands) > 0 {
		entries = append(entries, helpEntry{"ACTIONS COMMANDS", strings.Join(actionsCommands, "\n")})
	}
	if len(additionalCommands) > 0 {
		entries = append(entries, helpEntry{"ADDITIONAL COMMANDS", strings.Join(additionalCommands, "\n")})
	}

	flagUsages := command.LocalFlags().FlagUsages()
	if flagUsages != "" {
		entries = append(entries, helpEntry{"FLAGS", dedent(flagUsages)})
	}
	inheritedFlagUsages := command.InheritedFlags().FlagUsages()
	if inheritedFlagUsages != "" {
		entries = append(entries, helpEntry{"INHERITED FLAGS", dedent(inheritedFlagUsages)})
	}
	if _, ok := command.Annotations["help:arguments"]; ok {
		entries = append(entries, helpEntry{"ARGUMENTS", command.Annotations["help:arguments"]})
	}
	if command.Example != "" {
		entries = append(entries, helpEntry{"EXAMPLES", command.Example})
	}
	if _, ok := command.Annotations["help:environment"]; ok {
		entries = append(entries, helpEntry{"ENVIRONMENT VARIABLES", command.Annotations["help:environment"]})
	}
	entries = append(entries, helpEntry{"LEARN MORE", `
Use 'inst <command> <subcommand> --help' for more information about a command.
Read the manual at https://www.instill.tech/docs`})
	if _, ok := command.Annotations["help:feedback"]; ok {
		entries = append(entries, helpEntry{"FEEDBACK", command.Annotations["help:feedback"]})
	}

	// Build markdown for rendering
	md := buildHelpMarkdown(entries)

	// Render with glamour when output is TTY for better readability
	if io.IsStdoutTTY() {
		wrapWidth := io.TerminalWidth()
		style := markdown.GetStyle(io.DetectTerminalTheme())
		rendered, err := markdown.RenderWithWrap(md, style, wrapWidth)
		if err == nil {
			_ = io.StartPager()
			defer io.StopPager()
			fmt.Fprint(io.Out, rendered)
			return
		}
		// Fall back to plain output on render error
	}

	// Plain output when not TTY or on render error
	out := command.OutOrStdout()
	cs := io.ColorScheme()
	for _, e := range entries {
		if e.Title != "" {
			fmt.Fprintln(out, cs.Bold(e.Title))
			fmt.Fprintln(out, text.Indent(strings.Trim(e.Body, "\r\n"), "  "))
		} else {
			fmt.Fprintln(out, e.Body)
		}
		fmt.Fprintln(out)
	}
}

// buildHelpMarkdown converts help entries to markdown format for glamour rendering.
func buildHelpMarkdown(entries []helpEntry) string {
	codeBlockSections := map[string]bool{
		"USAGE": true, "FLAGS": true, "INHERITED FLAGS": true,
		"EXAMPLES": true, "ARGUMENTS": true,
		"ENVIRONMENT VARIABLES": true, "LEARN MORE": true, "FEEDBACK": true,
	}
	var buf strings.Builder
	for _, e := range entries {
		if e.Title != "" {
			buf.WriteString("## ")
			buf.WriteString(e.Title)
			buf.WriteString("\n\n")
			if codeBlockSections[e.Title] {
				buf.WriteString("```\n")
			}
			buf.WriteString(strings.Trim(e.Body, "\r\n"))
			if codeBlockSections[e.Title] {
				buf.WriteString("\n```")
			}
			buf.WriteString("\n\n")
		} else {
			buf.WriteString(e.Body)
			buf.WriteString("\n\n")
		}
	}
	return buf.String()
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds ", padding)
	return fmt.Sprintf(template, s)
}

func dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	var buf bytes.Buffer
	for _, l := range lines {
		fmt.Fprintln(&buf, strings.TrimPrefix(l, strings.Repeat(" ", minIndent)))
	}
	return strings.TrimSuffix(buf.String(), "\n")
}
