package root

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/pkg/text"
)

var HelpTopics = map[string]map[string]string{
	"mintty": {
		"short": "Information about using instill with MinTTY",
		"long": heredoc.Doc(`
			MinTTY is the terminal emulator that comes by default with Git
			for Windows.  It has known issues with instill's ability to prompt a
			user for input.

			There are a few workarounds to make instill work with MinTTY:

			- Reinstall Git for Windows, checking "Enable experimental support for pseudo consoles".

			- Use a different terminal emulator with Git for Windows like Windows Terminal.
			  You can run "C:\Program Files\Git\bin\bash.exe" from any terminal emulator to continue
			  using all of the tooling in Git For Windows without MinTTY.

			- Prefix invocations of instill with winpty, eg: "winpty instill auth login".
			  NOTE: this can lead to some UI bugs.
		`),
	},
	"environment": {
		"short": "Environment variables that can be used with instill",
		"long": heredoc.Doc(`
			INSTILL_HOSTNAME: specify the Instill hostname for commands that would otherwise assume the
			"github.com" host when not in a context of an existing repository.

			INSTILL_EDITOR, GIT_EDITOR, VISUAL, EDITOR (in order of precedence): the editor tool to use
			for authoring text.

			INSTILL_BROWSER, BROWSER (in order of precedence): the web browser to use for opening links.

			DEBUG: set to any value to enable verbose output to standard error. Include values "api"
			or "oauth" to print detailed information about HTTP requests or authentication flow.

			INSTILL_PAGER, PAGER (in order of precedence): a terminal paging program to send standard output
			to, e.g. "less".

			GLAMOUR_STYLE: the style to use for rendering Markdown. See
			<https://github.com/charmbracelet/glamour#styles>

			NO_COLOR: set to any value to avoid printing ANSI escape sequences for color output.

			CLICOLOR: set to "0" to disable printing ANSI colors in output.

			CLICOLOR_FORCE: set to a value other than "0" to keep ANSI colors in output
			even when the output is piped.

			INSTILL_FORCE_TTY: set to any value to force terminal-style output even when the output is
			redirected. When the value is a number, it is interpreted as the number of columns
			available in the viewport. When the value is a percentage, it will be applied against
			the number of columns available in the current viewport.

			INSTILL_NO_UPDATE_NOTIFIER: set to any value to disable update notifications. By default, instill
			checks for new releases once every 24 hours and displays an upgrade notice on standard
			error if a newer version was found.

			INSTILL_CONFIG_DIR: the directory where instill will store configuration files. Default:
			"$XDG_CONFIG_HOME/instill" or "$HOME/.config/instill".
		`),
	},
	"reference": {
		"short": "A comprehensive reference of all instill commands",
	},
	"formatting": {
		"short": "Formatting options for JSON data exported from instill",
		"long": heredoc.Docf(`
			Some instill commands support exporting the data as JSON as an alternative to their usual
			line-based plain text output. This is suitable for passing structured data to scripts.
			The JSON output is enabled with the %[1]s--json%[1]s option, followed by the list of fields
			to fetch. Use the flag without a value to get the list of available fields.

			The %[1]s--jq%[1]s option accepts a query in jq syntax and will print only the resulting
			values that match the query. This is equivalent to piping the output to %[1]sjq -r%[1]s,
			but does not require the jq utility to be installed on the system. To learn more
			about the query syntax, see: <https://stedolan.github.io/jq/manual/v1.6/>

			With %[1]s--template%[1]s, the provided Go template is rendered using the JSON data as input.
			For the syntax of Go templates, see: <https://golang.org/pkg/text/template/>

			The following functions are available in templates:
			- %[1]sautocolor%[1]s: like %[1]scolor%[1]s, but only emits color to terminals
			- %[1]scolor <style> <input>%[1]s: colorize input using <https://github.com/mgutz/ansi>
			- %[1]sjoin <sep> <list>%[1]s: joins values in the list using a separator
			- %[1]spluck <field> <list>%[1]s: collects values of a field from all items in the input
			- %[1]stablerow <fields>...%[1]s: aligns fields in output vertically as a table
			- %[1]stablerender%[1]s: renders fields added by tablerow in place
			- %[1]stimeago <time>%[1]s: renders a timestamp as relative to now
			- %[1]stimefmt <format> <time>%[1]s: formats a timestamp using Go's Time.Format function
			- %[1]struncate <length> <input>%[1]s: ensures input fits within length
		`, "`"),
		"example": heredoc.Doc(`
			# format issues as table
			$ instill issue list --json number,title --template \
			  '{{range .}}{{tablerow (printf "#%v" .number | autocolor "green") .title}}{{end}}'

			# format a pull request using multiple tables with headers
			$ instill pr view 3519 --json number,title,body,reviews,assignees --template \
			  '{{printf "#%v" .number}} {{.title}}

			  {{.body}}

			  {{tablerow "ASSIGNEE" "NAME"}}{{range .assignees}}{{tablerow .login .name}}{{end}}{{tablerender}}
			  {{tablerow "REVIEWER" "STATE" "COMMENT"}}{{range .reviews}}{{tablerow .author.login .state .body}}{{end}}
			  '
		`),
	},
}

func NewHelpTopic(topic string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     topic,
		Short:   HelpTopics[topic]["short"],
		Long:    HelpTopics[topic]["long"],
		Example: HelpTopics[topic]["example"],
		Hidden:  true,
		Annotations: map[string]string{
			"markdown:generate": "true",
			"markdown:basename": "instill_help_" + topic,
		},
	}

	cmd.SetHelpFunc(helpTopicHelpFunc)
	cmd.SetUsageFunc(helpTopicUsageFunc)

	return cmd
}

func helpTopicHelpFunc(command *cobra.Command, args []string) {
	command.Print(command.Long)
	if command.Example != "" {
		command.Printf("\n\nEXAMPLES\n")
		command.Print(text.Indent(command.Example, "  "))
	}
}

func helpTopicUsageFunc(command *cobra.Command) error {
	command.Printf("Usage: instill help %s", command.Use)
	return nil
}
