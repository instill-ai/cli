package oauth2

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/instill-ai/cli/api"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/graceful"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/randx"
	"golang.org/x/oauth2"
)

var (
	// The "Instill CLI" OAuth app
	oauthClientID = ""
	// This value i	s safe to be embedded in version control
	oauthClientSecret = ""
)

type iconfig interface {
	Get(string, string) (string, error)
	Set(string, string, string) error
	Write() error
}

// AuthCodeFlowWithConfig authorizes a user via Authorization Code Flow
func AuthCodeFlowWithConfig(f *cmdutil.Factory, cfg iconfig, IO *iostreams.IOStreams, hostname string) error {

	serverHost := "localhost"
	serverPort := 8085

	fmt.Fprintf(IO.Out, "Login to %s. Press ctrl + c to end the process.\n\n", hostname)

	conf := &oauth2.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://auth.%s/oauth2/auth", hostname),
			TokenURL: fmt.Sprintf("https://auth.%s/oauth2/token", hostname),
		},
		RedirectURL: fmt.Sprintf("http://%s:%d/%s", serverHost, serverPort, "callback"),
		Scopes:      []string{"offline", "openid", "email", "profile"},
	}

	audience := []string{fmt.Sprintf("https://api.%s", hostname)}
	prompt := []string{""}
	maxAge := 0

	authCodeURL, state := generateAuthCodeURL(conf, audience, prompt, maxAge)
	fmt.Fprintf(IO.Out, "Complete the login via your OIDC provider. Launching a browser to:\n\n\t%s\n\n", authCodeURL)

	if err := f.Browser.Browse(authCodeURL); err != nil {
		return err
	}

	tokenChen := make(chan *oauth2.Token)
	go setLocalAuthServer("localhost", 8085, conf, state, IO, tokenChen)
	token := <-tokenChen

	if verbose := os.Getenv("DEBUG"); strings.Contains(verbose, "oauth") {
		fmt.Fprintf(IO.Out, "[DEBUG] Token Type:\n\t%s\n", token.Type())
		fmt.Fprintf(IO.Out, "[DEBUG] Access Token:\n\t%s\n", token.AccessToken)
		fmt.Fprintf(IO.Out, "[DEBUG] Expires at:\n\t%s\n", token.Expiry.Format(time.RFC1123))
		fmt.Fprintf(IO.Out, "[DEBUG] Refresh Token:\n\t%s\n", token.RefreshToken)
		fmt.Fprintf(IO.Out, "[DEBUG] ID Token:\n\t%s\n\n", token.Extra("id_token"))
	}

	if err := cfg.Set(hostname, "token_type", token.Type()); err != nil {
		return err
	}

	if err := cfg.Set(hostname, "access_token", token.AccessToken); err != nil {
		return err
	}

	if err := cfg.Set(hostname, "expiry", token.Expiry.Format(time.RFC1123)); err != nil {
		return err
	}

	if err := cfg.Set(hostname, "refresh_token", token.RefreshToken); err != nil {
		return err
	}

	if err := cfg.Set(hostname, "id_token", token.Extra("id_token").(string)); err != nil {
		return err
	}

	if err := cfg.Write(); err != nil {
		return err
	}

	fmt.Fprintf(IO.Out, "%s Authentication complete. %s to continue...\n", IO.ColorScheme().SuccessIcon(), IO.ColorScheme().Bold("Press Enter"))
	_ = waitForEnter(IO.In)
	return nil

}

func generateAuthCodeURL(conf *oauth2.Config, audience []string, prompt []string, maxAge int) (string, []rune) {

	state, err := randx.RuneSequence(24, randx.AlphaLower)
	cmdx.Must(err, "Could not generate random state: %s", err)

	nonce, err := randx.RuneSequence(24, randx.AlphaLower)
	cmdx.Must(err, "Could not generate random state: %s", err)

	authCodeURL := conf.AuthCodeURL(
		string(state),
		oauth2.SetAuthURLParam("audience", strings.Join(audience, "+")),
		oauth2.SetAuthURLParam("nonce", string(nonce)),
		oauth2.SetAuthURLParam("prompt", strings.Join(prompt, "+")),
		oauth2.SetAuthURLParam("max_age", strconv.Itoa(maxAge)),
		oauth2.SetAuthURLParam("access_type", "offline"),
	)

	return authCodeURL, state
}

func setLocalAuthServer(serverHost string, serverPort int, conf *oauth2.Config, state []rune, IO *iostreams.IOStreams, tokenChan chan *oauth2.Token) {

	var err error
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

		token, err = conf.Exchange(ctx, code)
		if err != nil {
			fmt.Fprintf(IO.ErrOut, "Unable to exchange code for token: %s\n", err)
			tokenChan <- token
			close(tokenChan)
			cancel()
			return
		}

		// TODO: Replace the endpoint with auth success page
		http.Redirect(w, r, "https://instill.tech", 302)

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
		fmt.Fprintf(IO.ErrOut, "local auth server error: %s", err)
	}
}

func waitForEnter(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return scanner.Err()
}
