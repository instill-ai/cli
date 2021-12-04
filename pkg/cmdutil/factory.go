package cmdutil

import (
	"net/http"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/iostreams"
)

type Browser interface {
	Browse(string) error
}

type Factory struct {
	IOStreams *iostreams.IOStreams
	Browser   Browser

	HTTPClient func() (*http.Client, error)
	Config     func() (config.Config, error)

	// Executable is the path to the currently invoked instill binary
	Executable func() string
}
