package shared

import (
	"fmt"
)

type tinyConfig map[string]string

func (c tinyConfig) Get(host, key string) (string, error) {
	return c[fmt.Sprintf("%s:%s", host, key)], nil
}

func (c tinyConfig) Set(host string, key string, value string) error {
	c[fmt.Sprintf("%s:%s", host, key)] = value
	return nil
}

func (c tinyConfig) Write() error {
	return nil
}
