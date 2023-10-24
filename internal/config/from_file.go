package config

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/instill-ai/cli/internal/instance"
)

// This type implements a Config interface and represents a config file on disk.
type fileConfig struct {
	ConfigMap
	documentRoot *yaml.Node
}

type HostConfig struct {
	ConfigMap
	Host string
}

func (c *fileConfig) Root() *yaml.Node {
	return c.ConfigMap.Root
}

// Get gets a string value of `key`, from optional `hostname` or from root in case of no `hostname`.
func (c *fileConfig) Get(hostname, key string) (string, error) {
	val, _, err := c.GetWithSource(hostname, key)
	return val, err
}

func (c *fileConfig) GetWithSource(hostname, key string) (string, string, error) {
	if hostname != "" {
		var notFound *NotFoundError

		hostCfg, err := c.ConfigForHost(hostname)
		if err != nil && !errors.As(err, &notFound) {
			return "", "", err
		}

		var hostValue string
		if hostCfg != nil {
			hostValue, err = hostCfg.GetStringValue(key)
			if err != nil && !errors.As(err, &notFound) {
				return "", "", err
			}
		}

		if hostValue != "" {
			return hostValue, HostsConfigFile(), nil
		}
	}

	defaultSource := ConfigFile()

	value, err := c.GetStringValue(key)

	var notFound *NotFoundError

	if err != nil && errors.As(err, &notFound) {
		return defaultFor(key), defaultSource, nil
	} else if err != nil {
		return "", defaultSource, err
	}

	if value == "" {
		return defaultFor(key), defaultSource, nil
	}

	return value, defaultSource, nil
}

func (c *fileConfig) Set(hostname, key, value string) error {
	if hostname == "" {
		return c.SetStringValue(key, value)
	} else {
		hostCfg, err := c.ConfigForHost(hostname)
		var notFound *NotFoundError
		if errors.As(err, &notFound) {
			hostCfg = c.MakeConfigForHost(hostname)
		} else if err != nil {
			return err
		}
		return hostCfg.SetStringValue(key, value)
	}
}

func (c *fileConfig) UnsetHost(hostname string) error {

	if hostname == "" {
		return nil
	}

	hostsEntry, err := c.FindEntry("hosts")
	if err != nil {
		return err
	}

	cm := ConfigMap{hostsEntry.ValueNode}
	cm.RemoveEntry(hostname)

	_, err = c.hostEntries()
	if strings.Contains(err.Error(), "could not find any host configurations") {
		// no hosts, fallback to the default hostname
		defaultHost := instance.FallbackHostname()
		err = c.Set("", "default_hostname", defaultHost)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *fileConfig) ConfigForHost(hostname string) (*HostConfig, error) {
	hosts, err := c.hostEntries()
	if err != nil {
		return nil, err
	}

	for _, hc := range hosts {
		if strings.EqualFold(hc.Host, hostname) {
			return hc, nil
		}
	}
	return nil, &NotFoundError{fmt.Errorf("could not find config entry for %q", hostname)}
}

func (c *fileConfig) CheckWriteable(hostname, key string) error {
	// TODO: check filesystem permissions
	return nil
}

func (c *fileConfig) Write() error {
	mainData := yaml.Node{Kind: yaml.MappingNode}
	hostsData := yaml.Node{Kind: yaml.MappingNode}

	nodes := c.documentRoot.Content[0].Content
	for i := 0; i < len(nodes)-1; i += 2 {
		if nodes[i].Value == "hosts" {
			hostsData.Content = append(hostsData.Content, nodes[i+1].Content...)
		} else {
			mainData.Content = append(mainData.Content, nodes[i], nodes[i+1])
		}
	}

	mainBytes, err := yaml.Marshal(&mainData)
	if err != nil {
		return err
	}

	filename := ConfigFile()
	err = WriteConfigFile(filename, yamlNormalize(mainBytes))
	if err != nil {
		return err
	}

	hostsBytes, err := yaml.Marshal(&hostsData)
	if err != nil {
		return err
	}

	return WriteConfigFile(HostsConfigFile(), yamlNormalize(hostsBytes))
}

func (c *fileConfig) hostEntries() ([]*HostConfig, error) {
	entry, err := c.FindEntry("hosts")
	if err != nil {
		return []*HostConfig{}, nil
	}

	hostConfigs, err := c.parseHosts(entry.ValueNode)
	if err != nil {
		return nil, fmt.Errorf("could not parse hosts config: %w", err)
	}

	return hostConfigs, nil
}

// Hosts returns a list of all known hostnames configured in hosts.yml
// TODO replace with HostsTyped
func (c *fileConfig) Hosts() ([]string, error) {
	entries, err := c.hostEntries()
	if err != nil {
		return nil, err
	}

	hostnames := []string{}
	for _, entry := range entries {
		hostnames = append(hostnames, entry.Host)
	}

	sort.SliceStable(hostnames, func(i, j int) bool { return hostnames[i] == instance.FallbackHostname() })

	return hostnames, nil
}

// HostsTyped returns an array of typesafe host configs.
// Every call re-reads the config file.
func (c *fileConfig) HostsTyped() ([]HostConfigTyped, error) {
	var ret []HostConfigTyped
	hosts, err := c.hostEntries()
	if err != nil {
		return nil, err
	}
	defaultHost, err := c.Get("", "default_hostname")
	if err != nil {
		return nil, err
	}
	defaultMatched := false
	for _, h := range hosts {
		ht, err := hostConfigToTyped(h)
		if err != nil {
			return nil, err
		}
		// max 1 default host
		if defaultHost == ht.APIHostname {
			ht.IsDefault = true
			defaultMatched = true
		}

		ret = append(ret, *ht)
	}
	// at least 1 default
	if len(ret) > 0 && !defaultMatched {
		ret[0].IsDefault = true
		defaultHost = ret[0].APIHostname
		// update the default hostname
		err = c.Set("", "default_hostname", defaultHost)
		if err != nil {
			return nil, err
		}
		err = c.Write()
		if err != nil {
			return nil, err
		}
	}
	// sort by name
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].APIHostname < ret[j].APIHostname
	})
	return ret, nil
}

// DefaultHostname returns the default API hostname, or a fallback in case of none or an error.
func (c *fileConfig) DefaultHostname() string {
	hostname, err := c.Get("", "default_hostname")
	if err != nil {
		return instance.FallbackHostname()
	}
	return hostname
}

func (c *fileConfig) MakeConfigForHost(hostname string) *HostConfig {
	hostRoot := &yaml.Node{Kind: yaml.MappingNode}
	hostCfg := &HostConfig{
		Host:      hostname,
		ConfigMap: ConfigMap{Root: hostRoot},
	}

	var notFound *NotFoundError
	hostsEntry, err := c.FindEntry("hosts")
	if errors.As(err, &notFound) {
		hostsEntry.KeyNode = &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "hosts",
		}
		hostsEntry.ValueNode = &yaml.Node{Kind: yaml.MappingNode}
		root := c.Root()
		root.Content = append(root.Content, hostsEntry.KeyNode, hostsEntry.ValueNode)
	} else if err != nil {
		panic(err)
	}

	hostsEntry.ValueNode.Content = append(hostsEntry.ValueNode.Content,
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: hostname,
		}, hostRoot)

	return hostCfg
}

func (c *fileConfig) parseHosts(hostsEntry *yaml.Node) ([]*HostConfig, error) {
	hostConfigs := []*HostConfig{}

	for i := 0; i < len(hostsEntry.Content)-1; i = i + 2 {
		hostname := hostsEntry.Content[i].Value
		hostRoot := hostsEntry.Content[i+1]
		hostConfig := HostConfig{
			ConfigMap: ConfigMap{Root: hostRoot},
			Host:      hostname,
		}
		hostConfigs = append(hostConfigs, &hostConfig)
	}

	if len(hostConfigs) == 0 {
		return nil, errors.New("could not find any host configurations")
	}

	return hostConfigs, nil
}

// Save persists the host config into the passed `fileConfig`.
func (c *fileConfig) SaveTyped(host *HostConfigTyped) error {
	conf, err := c.ConfigForHost(host.APIHostname)
	var notFound *NotFoundError
	if errors.As(err, &notFound) {
		conf = c.MakeConfigForHost(host.APIHostname)
	} else if err != nil {
		return err
	}
	err = hostTypedToConfig(host, conf)
	if err != nil {
		return err
	}
	// update the default instance in the main config
	if host.IsDefault {
		err = c.Set("", "default_hostname", host.APIHostname)
		if err != nil {
			return err
		}
	}
	return c.Write()
}

// HostConfigTyped is a type safe representation of an instance config.
// TODO keep in sync with `hostConfigToTyped`
// TODO bind directly to yaml via struct tags
// TODO validation
type HostConfigTyped struct {
	// instance info
	APIHostname string `example:"api.instill.tech"`
	IsDefault   bool
	APIVersion  string `example:"v1alpha"`

	// oauth config
	Oauth2Hostname     string `example:"auth.instill.tech"`
	Oauth2Audience     string `example:"https://instill.tech"`
	Oauth2Issuer       string `example:"https://auth.instill.tech/"`
	Oauth2ClientSecret string
	Oauth2ClientID     string

	// oauth token
	TokenType    string
	AccessToken  string
	Expiry       string
	RefreshToken string
	IDToken      string
}

// DefaultHostConfig returns default values for an instance config, as a single "source of truth" for other packages.
func DefaultHostConfig() HostConfigTyped {
	return HostConfigTyped{
		APIVersion: "v1alpha",
		IsDefault:  false,
	}
}

// hostConfigToTyped reads an untyped config into a `HostConfigTyped` struct.
func hostConfigToTyped(conf *HostConfig) (*HostConfigTyped, error) {
	ht := &HostConfigTyped{
		APIHostname: conf.Host,
	}
	ht.TokenType = conf.GetOptionalStringValue("token_type")
	ht.AccessToken = conf.GetOptionalStringValue("access_token")
	ht.Expiry = conf.GetOptionalStringValue("expiry")
	ht.RefreshToken = conf.GetOptionalStringValue("refresh_token")
	ht.IDToken = conf.GetOptionalStringValue("id_token")
	ht.Oauth2Audience = conf.GetOptionalStringValue("oauth2_audience")
	ht.Oauth2Issuer = conf.GetOptionalStringValue("oauth2_issuer")
	ht.Oauth2Hostname = conf.GetOptionalStringValue("oauth2_hostname")
	ht.Oauth2ClientSecret = conf.GetOptionalStringValue("oauth2_client_secret")
	ht.Oauth2ClientID = conf.GetOptionalStringValue("oauth2_client_id")
	ht.APIVersion = conf.GetOptionalStringValue("api_version")
	return ht, nil
}

// hostTypedToConfig propagates `HostConfigTyped` into `HostConfig`, so it can be persisted.s
func hostTypedToConfig(host *HostConfigTyped, conf *HostConfig) error {
	err := conf.SetStringValue("token_type", host.TokenType)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("access_token", host.AccessToken)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("expiry", host.Expiry)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("refresh_token", host.RefreshToken)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("id_token", host.IDToken)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("oauth2_audience", host.Oauth2Audience)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("oauth2_issuer", host.Oauth2Issuer)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("oauth2_hostname", host.Oauth2Hostname)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("oauth2_client_id", host.Oauth2ClientID)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("oauth2_client_secret", host.Oauth2ClientSecret)
	if err != nil {
		return err
	}
	err = conf.SetStringValue("api_version", host.APIVersion)
	if err != nil {
		return err
	}
	return nil
}

func yamlNormalize(b []byte) []byte {
	if bytes.Equal(b, []byte("{}\n")) {
		return []byte{}
	}
	return b
}

func defaultFor(key string) string {
	for _, co := range configOptions {
		if co.Key == key {
			return co.DefaultValue
		}
	}
	return ""
}
