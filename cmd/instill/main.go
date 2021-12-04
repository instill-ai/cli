package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	surveyCore "github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"

	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/api"
	"github.com/instill-ai/cli/internal/build"
	"github.com/instill-ai/cli/pkg/cmd/factory"
	"github.com/instill-ai/cli/pkg/cmd/root"
	"github.com/instill-ai/cli/pkg/cmdutil"
)

var updaterEnabled = ""

type exitCode int

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
	exitAuth   exitCode = 4
)

func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() exitCode {

	buildDate := build.Date
	buildVersion := build.Version

	// TODO: Implement update check
	// updateMessageChan := make(chan *update.ReleaseInfo)
	// go func() {
	// 	rel, _ := checkForUpdate(buildVersion)
	// 	updateMessageChan <- rel
	// }()

	hasDebug := os.Getenv("DEBUG") != ""

	cmdFactory := factory.New(buildVersion)
	stderr := cmdFactory.IOStreams.ErrOut

	if spec := os.Getenv("INSTILL_FORCE_TTY"); spec != "" {
		cmdFactory.IOStreams.ForceTerminal(spec)
	}

	if !cmdFactory.IOStreams.ColorEnabled() {
		surveyCore.DisableColor = true
	} else {
		// override survey's poor choice of color
		surveyCore.TemplateFuncsWithColor["color"] = func(style string) string {
			switch style {
			case "white":
				if cmdFactory.IOStreams.ColorSupport256() {
					return fmt.Sprintf("\x1b[%d;5;%dm", 38, 242)
				}
				return ansi.ColorCode("default")
			default:
				return ansi.ColorCode(style)
			}
		}
	}

	// Enable running instill from Windows File Explorer's address bar. Without this, the user is told to stop and run from a
	// terminal. With this, a user can clone a repo (or take other actions) directly from explorer.
	if len(os.Args) > 1 && os.Args[1] != "" {
		cobra.MousetrapHelpText = ""
	}

	rootCmd := root.NewCmdRoot(cmdFactory, buildVersion, buildDate)

	cfg, err := cmdFactory.Config()
	if err != nil {
		fmt.Fprintf(stderr, "failed to read configuration:  %s\n", err)
		return exitError
	}

	expandedArgs := []string{}
	if len(os.Args) > 0 {
		expandedArgs = os.Args[1:]
	}

	// translate `instill help <command>` to `instill <command> --help` for extensions
	if len(expandedArgs) == 2 && expandedArgs[0] == "help" && !hasCommand(rootCmd, expandedArgs[1:]) {
		expandedArgs = []string{expandedArgs[1], "--help"}
	}

	cs := cmdFactory.IOStreams.ColorScheme()

	authError := errors.New("authError")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// require that the user is authenticated before running most commands
		if cmdutil.IsAuthCheckEnabled(cmd) && !cmdutil.CheckAuth(cfg) {
			fmt.Fprintln(stderr, cs.Bold("Welcome to Instill CLI!"))
			fmt.Fprintln(stderr)
			fmt.Fprintln(stderr, "To authenticate, please run `instill auth login`.")
			return authError
		}

		return nil
	}

	rootCmd.SetArgs(expandedArgs)

	if cmd, err := rootCmd.ExecuteC(); err != nil {
		if err == cmdutil.SilentError {
			return exitError
		} else if cmdutil.IsUserCancellation(err) {
			if errors.Is(err, terminal.InterruptErr) {
				// ensure the next shell prompt will start on its own line
				fmt.Fprint(stderr, "\n")
			}
			return exitCancel
		} else if errors.Is(err, authError) {
			return exitAuth
		}

		printError(stderr, err, cmd, hasDebug)

		if strings.Contains(err.Error(), "Incorrect function") {
			fmt.Fprintln(stderr, "You appear to be running in MinTTY without pseudo terminal support.")
			fmt.Fprintln(stderr, "To learn about workarounds for this error, run:  instill help mintty")
			return exitError
		}

		var httpErr api.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 401 {
			fmt.Fprintln(stderr, "Try authenticating with:  instill auth login")
		}

		return exitError
	}
	if root.HasFailed() {
		return exitError
	}

	//TODO: Implement update check (Homebrew)
	// newRelease := <-updateMessageChan
	// if newRelease != nil {
	// 	isHomebrew := isUnderHomebrew(cmdFactory.Executable())
	// 	if isHomebrew && isRecentRelease(newRelease.PublishedAt) {
	// 		// do not notify Homebrew users before the version bump had a chance to get merged into homebrew-core
	// 		return exitOK
	// 	}
	// 	fmt.Fprintf(stderr, "\n\n%s %s → %s\n",
	// 		ansi.Color("A new release of instill is available:", "yellow"),
	// 		ansi.Color(buildVersion, "cyan"),
	// 		ansi.Color(newRelease.Version, "cyan"))
	// 	if isHomebrew {
	// 		fmt.Fprintf(stderr, "To upgrade, run: %s\n", "brew update && brew upgrade instill")
	// 	}
	// 	fmt.Fprintf(stderr, "%s\n\n",
	// 		ansi.Color(newRelease.URL, "yellow"))
	// }

	return exitOK
}

// hasCommand returns true if args resolve to a built-in command
func hasCommand(rootCmd *cobra.Command, args []string) bool {
	c, _, err := rootCmd.Traverse(args)
	return err == nil && c != rootCmd
}

func printError(out io.Writer, err error, cmd *cobra.Command, debug bool) {
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		fmt.Fprintf(out, "error connecting to %s\n", dnsError.Name)
		if debug {
			fmt.Fprintln(out, dnsError)
		}
		fmt.Fprintln(out, "check your internet connection or https://githubstatus.com")
		return
	}

	fmt.Fprintln(out, err)

	var flagError *cmdutil.FlagError
	if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "Unknown command ") {
		if !strings.HasSuffix(err.Error(), "\n") {
			fmt.Fprintln(out)
		}
		fmt.Fprintln(out, cmd.UsageString())
	}
}
