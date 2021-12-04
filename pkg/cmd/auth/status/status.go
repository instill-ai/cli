package status

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type StatusOptions struct {
	IO       *iostreams.IOStreams
	Config   func() (config.Config, error)
	Hostname string
}

func NewCmdStatus(f *cmdutil.Factory, runF func(*StatusOptions) error) *cobra.Command {
	opts := &StatusOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.ExactArgs(0),
		Short: "View authentication status",
		Long: heredoc.Doc(`Verifies and displays information about your authentication state.
			
			This command will test your authentication state for each Instill host that instill knows about and
			report on any issues.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return statusRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "", "Check a specific hostname's auth status")

	return cmd
}

func statusRun(opts *StatusOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	// TODO check tty

	stderr := opts.IO.ErrOut

	cs := opts.IO.ColorScheme()

	statusInfo := map[string][]string{}

	hostnames, err := cfg.Hosts()
	if err != nil {
		return err
	}
	if len(hostnames) == 0 {
		fmt.Fprintf(stderr,
			"You are not logged into any Instill hosts. Run %s to authenticate.\n", cs.Bold("instill auth login"))
		return cmdutil.SilentError
	}

	var failed bool
	var isHostnameFound bool

	for _, hostname := range hostnames {
		if opts.Hostname != "" && opts.Hostname != hostname {
			continue
		}
		isHostnameFound = true

		accessTokenExpire, accessTokenSource, _ := cfg.GetWithSource(hostname, "expire_at")
		// accessTokenIsWriteable := cfg.CheckWriteable(hostname, "access_token") == nil

		statusInfo[hostname] = []string{}
		addMsg := func(x string, ys ...interface{}) {
			statusInfo[hostname] = append(statusInfo[hostname], fmt.Sprintf(x, ys...))
		}

		addMsg("%s Logged in to %s (%s)", cs.SuccessIcon(), hostname, accessTokenSource)
		addMsg("%s Access token expires at: %s", cs.SuccessIcon(), cs.Bold(accessTokenExpire))

	}

	if !isHostnameFound {
		fmt.Fprintf(stderr,
			"Hostname %q not found among authenticated Instill hosts\n", opts.Hostname)
		return cmdutil.SilentError
	}

	for _, hostname := range hostnames {
		lines, ok := statusInfo[hostname]
		if !ok {
			continue
		}
		fmt.Fprintf(stderr, "%s\n", cs.Bold(hostname))
		for _, line := range lines {
			fmt.Fprintf(stderr, "  %s\n", line)
		}
	}

	if failed {
		return cmdutil.SilentError
	}

	return nil
}
