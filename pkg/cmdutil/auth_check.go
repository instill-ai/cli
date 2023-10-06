package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
)

func DisableAuthCheck(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}

	cmd.Annotations["skipAuthCheck"] = "true"
}

// CheckAuth checks if the default hostname has an access token assigned (without validating it).
func CheckAuth(cfg config.Config) bool {
	hosts, err := cfg.HostsTyped()
	if err != nil {
		return false
	}
	defHostname := cfg.DefaultHostname()
	for _, h := range hosts {
		// check the token only for instance with an OAuth2 hostname
		if h.APIHostname == defHostname && h.AccessToken != "" || h.Oauth2Hostname == "" {
			// TODO use oauth2.VerifyIDToken?
			return true
		}
	}
	return false
}

func IsAuthCheckEnabled(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return false
	}

	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["skipAuthCheck"] == "true" {
			return false
		}
	}

	return true
}
