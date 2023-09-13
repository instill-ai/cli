package shared

import (
	"github.com/instill-ai/cli/pkg/iostreams"
)

type iconfig interface {
	Get(string, string) (string, error)
	Set(string, string, string) error
	Write() error
}

type LoginOptions struct {
	IO          *iostreams.IOStreams
	Config      iconfig
	Hostname    string
	Interactive bool
	Executable  string
}
