package instance

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

// Default returns the host name of the default Instill instance
func Default() string {
	return "api.instill.tech"
}

// ExtractHostname returns the canonical host name of a Instill instance
func ExtractHostname(h string) string {
	hostname := strings.ToLower(h)

	parts := strings.Split(hostname, ".")
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

// HostnameValidator validates a hostname
// TODO move to utils
func HostnameValidator(v string) error {
	host := struct {
		name string `validate:"required,hostname"`
	}{
		name: v,
	}
	err := validator.New(validator.WithRequiredStructEnabled()).Struct(host)
	if err != nil {
		return err
	}
	return nil
}

// RESTPrefix sets the prefix of Instill API URL
// TODO remove
func RESTPrefix(hostname string) string {
	if strings.EqualFold(hostname, "localhost") {
		return fmt.Sprintf("http://api.%s/", hostname)
	}
	return fmt.Sprintf("https://api.%s/", hostname)
}

// HostPrefix sets the prefix of Instill domain
func HostPrefix(hostname string) string {
	if strings.HasSuffix(hostname, "localhost") {
		return fmt.Sprintf("http://%s/", hostname)
	}
	return fmt.Sprintf("https://%s/", hostname)
}
