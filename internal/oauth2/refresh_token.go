package oauth2

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

func RefreshToken(cfg iconfig, hostname string) (string, error) {

	var (
		accessToken  string
		refreshToken string
		expiry       time.Time
		err          error
	)

	accessToken, err = cfg.Get(hostname, "access_token")
	if err != nil {
		return "", err
	}

	refreshToken, err = cfg.Get(hostname, "refresh_token")
	if err != nil {
		return "", err
	}

	expiry, err = func() (time.Time, error) {
		if str, err := cfg.Get(hostname, "expiry"); err == nil {
			if expiry, err := time.Parse(time.RFC1123, str); err == nil {
				return expiry, nil
			}
			return time.Time{}, err
		}
		return time.Time{}, err
	}()

	if err != nil {
		return "", err
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://auth.%s/oauth2/auth", hostname),
			TokenURL: fmt.Sprintf("https://auth.%s/oauth2/token", hostname),
		},
	}

	token, err := conf.TokenSource(
		context.Background(),
		&oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Expiry:       expiry,
		}).Token()

	if err != nil {
		return "", err
	}

	if accessToken != token.AccessToken {
		if err := cfg.Set(hostname, "access_token", token.AccessToken); err != nil {
			return "", err
		}
	}

	if refreshToken != token.RefreshToken {
		if err := cfg.Set(hostname, "refresh_token", token.RefreshToken); err != nil {
			return "", err
		}
	}

	if expiry != token.Expiry {
		if err := cfg.Set(hostname, "expiry", token.Expiry.Format(time.RFC1123)); err != nil {
			return "", err
		}
	}

	if err := cfg.Write(); err != nil {
		return "", err
	}

	return token.AccessToken, nil

}
