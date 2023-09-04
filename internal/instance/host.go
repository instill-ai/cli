package instance

import (
	"errors"
	"fmt"
	"strings"
)

const defaultHostname = "instill.tech"

// localhost is the domain name of a local Instill instance
const localhost = "instill.localhost"

// Default returns the host name of the default Instill instance
// TODO update with instances
func Default() string {
	return defaultHostname
}

// ExtractHostname returns the canonical host name of a Instill instance
func ExtractHostname(h string) string {
	hostname := strings.ToLower(h)

	parts := strings.Split(hostname, ".")
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

// HostnameValidator validates a hostname
func HostnameValidator(v interface{}) error {
	hostname, valid := v.(string)
	if !valid {
		return errors.New("hostname is not a string")
	}

	if len(strings.TrimSpace(hostname)) < 1 {
		return errors.New("a value is required")
	}
	if strings.ContainsRune(hostname, '/') || strings.ContainsRune(hostname, ':') {
		return errors.New("invalid hostname")
	}
	return nil
}

// RESTPrefix sets the prefix of Instill API URL
func RESTPrefix(hostname string) string {
	if strings.EqualFold(hostname, localhost) {
		return fmt.Sprintf("http://api.%s/", hostname)
	}
	return fmt.Sprintf("https://api.%s/", hostname)
}

// HostPrefix sets the prefix of Instill domain
func HostPrefix(hostname string) string {
	if strings.EqualFold(hostname, localhost) {
		return fmt.Sprintf("http://%s/", hostname)
	}
	return fmt.Sprintf("https://%s/", hostname)
}
