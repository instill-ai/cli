package config

import (
	"errors"
	"github.com/instill-ai/cli/internal/instance"
)

type ConfigStub map[string]string

func genKey(host, key string) string {
	if host != "" {
		return host + ":" + key
	}
	return key
}

func (c ConfigStub) Get(host, key string) (string, error) {
	val, _, err := c.GetWithSource(host, key)
	return val, err
}

func (c ConfigStub) GetWithSource(host, key string) (string, string, error) {
	if v, found := c[genKey(host, key)]; found {
		return v, "(memory)", nil
	}
	return "", "", errors.New("not found")
}

func (c ConfigStub) Set(host, key, value string) error {
	c[genKey(host, key)] = value
	return nil
}

func (c ConfigStub) Hosts() ([]string, error) {
	return nil, nil
}

func (c ConfigStub) UnsetHost(hostname string) {
	// TODO
}

func (c ConfigStub) CheckWriteable(host, key string) error {
	return nil
}

func (c ConfigStub) Write() error {
	c["_written"] = "true"
	return nil
}

func (c ConfigStub) DefaultHost() (string, error) {
	return "", nil
}

func (c ConfigStub) DefaultHostWithSource() (string, string, error) {
	return "", "", nil
}

func (c ConfigStub) DefaultHostname() string {
	return instance.FallbackHostname()
}

func (c ConfigStub) MakeConfigForHost(hostname string) *HostConfig {
	return nil
}

func (c ConfigStub) HostsTyped() ([]HostConfigTyped, error) {
	ins := []HostConfigTyped{
		{
			APIHostname:    "api.instill.tech",
			IsDefault:      true,
			APIVersion:     "v1alpha",
			Oauth2Hostname: "auth.instill.tech",
			Oauth2Audience: "https://api.instill.tech",
			Oauth2Issuer:   "https://auth.instill.tech/",
			Oauth2Secret:   "foobar",
			Oauth2ClientID: "barfoo",
		},
	}
	return ins, nil
}

func (c ConfigStub) SaveTyped(host *HostConfigTyped) error {
	h := host.APIHostname
	_ = c.Set(h, "token_type", host.TokenType)
	_ = c.Set(h, "access_token", host.AccessToken)
	_ = c.Set(h, "expiry", host.Expiry)
	_ = c.Set(h, "refresh_token", host.RefreshToken)
	_ = c.Set(h, "id_token", host.IDToken)
	_ = c.Set(h, "oauth2_audience", host.Oauth2Audience)
	_ = c.Set(h, "oauth2_issuer", host.Oauth2Issuer)
	_ = c.Set(h, "oauth2_hostname", host.Oauth2Hostname)
	_ = c.Set(h, "oauth2_client_id", host.Oauth2ClientID)
	_ = c.Set(h, "oauth2_secret", host.Oauth2Secret)
	_ = c.Set(h, "api_version", host.APIVersion)
	// TODO default instance
	return c.Write()
}

func ConfigStubFactory() (Config, error) {
	return ConfigStub{}, nil
}

type HostConfigMock struct {
	HostConfig
}
