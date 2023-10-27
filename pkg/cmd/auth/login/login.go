package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/internal/oauth2"
	"github.com/instill-ai/cli/pkg/cmd/local"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
	"github.com/instill-ai/cli/pkg/prompt"
)

// LoginOptions contains the command line options
type LoginOptions struct {
	IO             *iostreams.IOStreams
	Config         func() (config.Config, error)
	MainExecutable string
	Interactive    bool
	Hostname       string
}

// NewCmdLogin creates a new command
func NewCmdLogin(f *cmdutil.Factory, runF func(*LoginOptions) error) *cobra.Command {
	opts := &LoginOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "login",
		Args:  cobra.ExactArgs(0),
		Short: "Authenticate with an Instill Core/Cloud host",
		Long: heredoc.Docf(`
			Authenticate with an Instill Core/Cloud host.

			The default authentication mode is an authorization code flow.
		`),
		Example: heredoc.Doc(`
			# login to the default host
			$ inst auth login
		`),
		RunE: func(cmd *cobra.Command, args []string) error {

			if opts.IO.CanPrompt() {
				opts.Interactive = true
			}

			opts.MainExecutable = f.Executable()
			if runF != nil {
				return runF(opts)
			}

			return loginRun(f, opts)
		},
	}

	// TODO handle err
	cfg, _ := opts.Config()

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", cfg.DefaultHostname(), "Hostname of an already added Instill Core/Cloud instance")

	return cmd
}

func loginRun(f *cmdutil.Factory, opts *LoginOptions) error {

	var host *config.HostConfigTyped

	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	// in case there's no hosts.yml config, create one with the default instance
	fs, err := os.Stat(config.HostsConfigFile())
	if !(err == nil && !fs.IsDir()) {
		// get the (hardcoded) default cloud instance
		host = oauth2.HostConfigInstillCloud()
		err = cfg.SaveTyped(host)
		if err != nil {
			return err
		}
	} else {
		// in case the hosts.yml is empty
		cfg, _ := opts.Config()
		if hosts, err := cfg.HostEntries(); err != nil {
			return err
		} else if len(hosts) == 0 {
			// no hosts, fallback to the default hostname
			host = oauth2.HostConfigInstillCloud()
			err = cfg.SaveTyped(host)
			if err != nil {
				return err
			}
			err = cfg.Set("", "default_hostname", host.APIHostname)
			if err != nil {
				return err
			}
		} else {
			hostname := opts.Hostname
			hosts, err := cfg.HostsTyped()
			if err != nil {
				return err
			}
			for _, h := range hosts {
				if h.APIHostname == hostname {
					host = &h
					break
				}
			}
			if host == nil {
				return fmt.Errorf("ERROR: instance '%s' does not exists", hostname)
			}
		}
	}

	// TODO INS-1659 drop in favor of OAuth2
	if instance.IsLocal(host.APIHostname) {
		fmt.Println("Logging into the local Instill Core instance...")
		var pass string
		err = prompt.SurveyAskOne(&survey.Password{
			Message: "Enter your password",
		}, &pass)
		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}
		token, err := loginLocal(nil, host.APIHostname, pass)
		if err != nil {
			return fmt.Errorf("ERROR: login failed, %w", err)
		}
		host.AccessToken = token
		err = cfg.SaveTyped(host)
		if err != nil {
			return err
		}
		return nil
	}

	if host.Oauth2Hostname == "" || host.Oauth2ClientID == "" || host.Oauth2ClientSecret == "" {
		e := heredoc.Docf(`ERROR: OAuth2 config isn't complete for '%s'

			You can fix it with:
			$ inst instance edit %s \
				--oauth2 HOSTNAME \
				--client-id CLIENT_ID \
				--client-secret CLIENT_SECRET`, host.APIHostname, host.APIHostname)
		return fmt.Errorf(e)
	}

	if host.RefreshToken != "" && opts.Interactive {
		var keepGoing bool
		err = prompt.SurveyAskOne(&survey.Confirm{
			Message: fmt.Sprintf(
				"You're already logged into %s. Do you want to re-authenticate?",
				host.APIHostname),
			Default: false,
		}, &keepGoing)
		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}
		if !keepGoing {
			return nil
		}
	}

	return oauth2.AuthCodeFlowWithConfig(f, host, cfg, opts.IO)
}

type localLoginResponse struct {
	AccessToken string `json:"access_token"`
}

type localLoginRequest struct {
	Name string `json:"username"`
	Pass string `json:"password"`
}

// loginLocal handles dedicated auth flow for Instill Core.
func loginLocal(transport http.RoundTripper, hostname, password string) (string, error) {
	url := instance.GetProtocol(hostname) + "core/v1alpha/auth/login"
	data := &localLoginRequest{
		Name: local.DefUsername,
		Pass: password,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var response localLoginResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	return response.AccessToken, nil
}
