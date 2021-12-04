package shared

import (
	"fmt"

	"github.com/instill-ai/cli/internal/oauth2"
	"github.com/instill-ai/cli/pkg/cmdutil"
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

func Login(f *cmdutil.Factory, opts *LoginOptions) error {
	cfg := opts.Config
	hostname := opts.Hostname

	var err error
	err = oauth2.AuthCodeFlowWithConfig(f, cfg, opts.IO, hostname)
	if err != nil {
		return fmt.Errorf("failed to authenticate via web browser: %w", err)
	}

	err = cfg.Write()
	if err != nil {
		return err
	}

	return nil
}
