package instance

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FallbackHostname returns the host name of the default Instill Cloud instance.
func FallbackHostname() string {
	return "api.instill.tech"
}

// HostnameValidator validates a hostname, with an optional port number.
// TODO move to utils
func HostnameValidator(v string) error {
	// without a port
	host1 := struct {
		Name string `validate:"required,hostname"`
	}{
		Name: v,
	}
	err1 := validator.New().Struct(&host1)
	// with a port
	host2 := struct {
		Name string `validate:"required,hostname_port"`
	}{
		Name: v,
	}
	err2 := validator.New().Struct(&host2)
	if err1 != nil && err2 != nil {
		return fmt.Errorf("hostname not valid")
	}
	return nil
}

// RESTPrefix sets the prefix of Instill API URL
// TODO remove
func RESTPrefix(hostname string) string {
	if strings.EqualFold(hostname, "localhost") {
		return fmt.Sprintf("http://%s/", hostname)
	}
	return fmt.Sprintf("https://%s/", hostname)
}

// GetProtocol returns the correct protocol based on a hostname
func GetProtocol(hostname string) string {
	// TODO support port numbers
	if strings.HasSuffix(hostname, "localhost") {
		return fmt.Sprintf("http://%s/", hostname)
	}
	return fmt.Sprintf("https://%s/", hostname)
}
