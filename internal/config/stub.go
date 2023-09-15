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

func (c ConfigStub) SaveTyped(*HostConfigTyped) error {
	return nil
}

func ConfigStubFactory() (Config, error) {
	return ConfigStub{}, nil
}
