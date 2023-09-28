package oauth2

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/graceful"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/randx"
	"golang.org/x/oauth2"

	"github.com/instill-ai/cli/api"
	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

var (
	// The "Instill CLI" OAuth app (build-time default to api.instill.tech)
	clientID = ""
	// This value is safe to be embedded in version control (build-time default to api.instill.tech)
	clientSecret = ""
)

// HostConfigInstillCloud return a host config for the main Instill AI Cloud server.
func HostConfigInstillCloud() *config.HostConfigTyped {

	host := config.DefaultHostConfig()
	host.APIHostname = "api.instill.tech"
	host.IsDefault = true
	host.Oauth2Hostname = "auth.instill.tech"
	host.Oauth2Audience = "https://api.instill.tech"
	host.Oauth2Issuer = "https://auth.instill.tech/"
	host.Oauth2ClientID = clientID
	host.Oauth2ClientSecret = clientSecret
	return &host
}

type iconfig interface {
	SaveTyped(*config.HostConfigTyped) error
	Get(string, string) (string, error)
	Set(string, string, string) error
	Write() error
}

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

// NewAuthenticator instantiates the *Authenticator.
func NewAuthenticator(issuer, clientID, clientSecret, callbackHost string, callbackPort int) (*Authenticator, error) {
	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  fmt.Sprintf("http://%s:%d/%s", callbackHost, callbackPort, "callback"),
		Scopes:       []string{"offline_access", "openid", "email", "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

// AuthCodeFlowWithConfig authorizes a user via Authorization Code Flow
func AuthCodeFlowWithConfig(f *cmdutil.Factory, host *config.HostConfigTyped, cfg iconfig, IO *iostreams.IOStreams) error {

	serverHost := "localhost"
	serverPort := 8085

	issuer := host.APIHostname
	if host.Oauth2Issuer != "" {
		issuer = host.Oauth2Issuer
	}
	audience := host.APIHostname
	if host.Oauth2Audience != "" {
		audience = host.Oauth2Audience
	}
	auth, err := NewAuthenticator(issuer, host.Oauth2ClientID, host.Oauth2ClientSecret, serverHost, serverPort)
	if err != nil {
		return err
	}

	prompt := []string{""}
	maxAge := 0
	loginURL, state := auth.LoginURL([]string{audience}, prompt, maxAge)

	fmt.Fprintf(IO.Out, "Log into %s. Press ctrl + c to end the process.\n\n", host.Oauth2Hostname)
	fmt.Fprintf(IO.Out, "Complete the login via your OIDC provider. Launching a browser to:\n\n\t%s\n\n", loginURL)

	if err := f.Browser.Browse(loginURL); err != nil {
		return err
	}

	tokenChan := make(chan *oauth2.Token)
	go handleCallback(auth, serverHost, serverPort, state, IO, tokenChan)
	token := <-tokenChan
	if token == nil {
		return errors.New("error receiving the token")
	}

	if verbose := os.Getenv("DEBUG"); strings.Contains(verbose, "oauth") {
		fmt.Fprintf(IO.Out, "[DEBUG] Token Type:\n\t%s\n", token.Type())
		fmt.Fprintf(IO.Out, "[DEBUG] Access Token:\n\t%s\n", token.AccessToken)
		fmt.Fprintf(IO.Out, "[DEBUG] Expires at:\n\t%s\n", token.Expiry.Format(time.RFC1123))
		fmt.Fprintf(IO.Out, "[DEBUG] Refresh Token:\n\t%s\n", token.RefreshToken)
		fmt.Fprintf(IO.Out, "[DEBUG] ID Token:\n\t%s\n\n", token.Extra("id_token"))
	}

	// TODO use HostConfigTyped
	host.TokenType = token.Type()
	host.AccessToken = token.AccessToken
	host.RefreshToken = token.RefreshToken
	host.Expiry = token.Expiry.Format(time.RFC1123)
	host.IDToken = token.Extra("id_token").(string)
	if err := cfg.SaveTyped(host); err != nil {
		return err
	}

	fmt.Fprintf(IO.Out, "%s Authentication complete. %s to continue...\n", IO.ColorScheme().SuccessIcon(), IO.ColorScheme().Bold("Press Enter"))
	_ = waitForInput(IO.In)
	return nil

}

func (a *Authenticator) LoginURL(audience []string, prompt []string, maxAge int) (string, []rune) {

	state, err := randx.RuneSequence(32, randx.AlphaLower)
	cmdx.Must(err, "Could not generate random state: %s", err)

	// TODO redundant?
	nonce, err := randx.RuneSequence(32, randx.AlphaLower)
	cmdx.Must(err, "Could not generate random state: %s", err)

	authCodeURL := a.AuthCodeURL(
		string(state),
		oauth2.SetAuthURLParam("audience", strings.Join(audience, "+")),
		oauth2.SetAuthURLParam("nonce", string(nonce)),
		oauth2.SetAuthURLParam("prompt", strings.Join(prompt, "+")),
		oauth2.SetAuthURLParam("max_age", strconv.Itoa(maxAge)),
		oauth2.SetAuthURLParam("access_type", "offline"),
	)

	return authCodeURL, state
}

func handleCallback(auth *Authenticator, serverHost string, serverPort int, state []rune, IO *iostreams.IOStreams, tokenChan chan *oauth2.Token) {

	//var err error
	var token *oauth2.Token

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use customised http.Client: https://stackoverflow.com/a/38150943
	httpClient := http.DefaultClient
	if verbose := os.Getenv("DEBUG"); verbose != "" {
		logTraffic := strings.Contains(verbose, "api") || strings.Contains(verbose, "oauth")
		httpClient.Transport = api.VerboseLog(IO.ErrOut, logTraffic, IO.ColorEnabled())(httpClient.Transport)
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}

	r := httprouter.New()
	server := graceful.WithDefaults(&http.Server{
		Addr:    fmt.Sprintf("%s:%v", serverHost, serverPort),
		Handler: r,
	})

	r.GET("/callback", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if len(r.URL.Query().Get("error")) > 0 {
			fmt.Fprintf(IO.ErrOut, "Got error: %s\n", r.URL.Query().Get("error_description"))
			tokenChan <- token
			close(tokenChan)
			cancel()
			return
		}

		if r.URL.Query().Get("state") != string(state) {
			fmt.Fprintf(IO.ErrOut, "States do not match. Expected %s, got %s\n", string(state), r.URL.Query().Get("state"))
			w.WriteHeader(http.StatusInternalServerError)
			tokenChan <- token
			close(tokenChan)
			cancel()
			return
		}

		code := r.URL.Query().Get("code")

		if verbose := os.Getenv("DEBUG"); strings.Contains(verbose, "oauth") {
			fmt.Printf("[DEBUG] Exchange code:\n\t%s\n", code)
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(ctx, code)
		if err != nil {
			fmt.Fprintf(IO.ErrOut, "Unable to exchange code for token: %s\n", err)
			tokenChan <- token
			close(tokenChan)
			cancel()
			return
		}
		_, err = auth.VerifyIDToken(ctx, token)
		if err != nil {
			fmt.Fprintf(IO.ErrOut, "Unable to validate token: %s\n", err)
			tokenChan <- token
			close(tokenChan)
			cancel()
			return
		}

		fmt.Fprint(w, oauthSuccessPage)

		tokenChan <- token
		close(tokenChan)
		cancel()
	})

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Fprintf(IO.ErrOut, "Local auth server error: %s\n", err)
			}
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		if err.Error() != context.Canceled.Error() {
			fmt.Fprintf(IO.ErrOut, "Local auth server error: %s", err)
		}
	}
}

func waitForInput(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return scanner.Err()
}
