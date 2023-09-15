package factory

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/iostreams"
)

// New initiate a factory instance
func New(appVersion string) *cmdutil.Factory {
	var exe string
	f := &cmdutil.Factory{
		Config: configFunc(), // No factory dependencies
		Executable: func() string {
			if exe != "" {
				return exe
			}
			exe = executable("instill")
			return exe
		},
	}

	f.IOStreams = ioStreams(f)                   // Depends on Config
	f.HTTPClient = httpClientFunc(f, appVersion) // Depends on Config, IOStreams, and appVersion
	f.Browser = browser(f)                       // Depends on Config, and IOStreams

	return f
}

func httpClientFunc(f *cmdutil.Factory, appVersion string) func() (*http.Client, error) {
	return func() (*http.Client, error) {
		io := f.IOStreams
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		return NewHTTPClient(io, cfg, appVersion, true)
	}
}

func browser(f *cmdutil.Factory) cmdutil.Browser {
	io := f.IOStreams
	return cmdutil.NewBrowser(browserLauncher(f), io.Out, io.ErrOut)
}

// Browser precedence
// 1. INSTILL_BROWSER
// 2. browser from config
// 3. BROWSER
func browserLauncher(f *cmdutil.Factory) string {
	if browser := os.Getenv("INSTILL_BROWSER"); browser != "" {
		return browser
	}

	cfg, err := f.Config()
	if err == nil {
		if cfgBrowser, _ := cfg.Get("", "browser"); cfgBrowser != "" {
			return cfgBrowser
		}
	}

	return os.Getenv("BROWSER")
}

// Finds the location of the executable for the current process as it's found in PATH, respecting symlinks.
// If the process couldn't determine its location, return fallbackName. If the executable wasn't found in
// PATH, return the absolute location to the program.
//
// The idea is that the result of this function is callable in the future and refers to the same
// installation of instill, even across upgrades. This is needed primarily for Homebrew, which installs software
// under a location such as `/usr/local/Cellar/instill/1.13.1/bin/instill` and symlinks it from `/usr/local/bin/instill`.
// When the version is upgraded, Homebrew will often delete older versions, but keep the symlink. Because of
// this, we want to refer to the `instill` binary as `/usr/local/bin/instill` and not as its internal Homebrew
// location.
//
// None of this would be needed if we could just refer to Instill CLI as `instill`, i.e. without using an absolute
// path. However, for some reason Homebrew does not include `/usr/local/bin` in PATH when it invokes git
// commands to update its taps. If `instill` (no path) is being used as git credential helper, as set up by `instill
// auth login`, running `brew update` will print out authentication errors as git is unable to locate
// Homebrew-installed `instill`.
func executable(fallbackName string) string {
	exe, err := os.Executable()
	if err != nil {
		return fallbackName
	}

	base := filepath.Base(exe)
	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		p, err := filepath.Abs(filepath.Join(dir, base))
		if err != nil {
			continue
		}
		f, err := os.Stat(p)
		if err != nil {
			continue
		}

		if p == exe {
			return p
		} else if f.Mode()&os.ModeSymlink != 0 {
			if t, err := os.Readlink(p); err == nil && t == exe {
				return p
			}
		}
	}

	return exe
}

func configFunc() func() (config.Config, error) {
	var cachedConfig config.Config
	var err error
	return func() (config.Config, error) {
		if cachedConfig != nil || err != nil {
			if err != nil {
				fmt.Printf("ERROR: cant read the config\n%s", err)
			}
			return cachedConfig, err
		}
		cachedConfig, err = config.ParseDefaultConfig()
		if errors.Is(err, os.ErrNotExist) {
			cachedConfig = config.NewBlankConfig()
			err = nil
		}
		return cachedConfig, err
	}
}

func ioStreams(f *cmdutil.Factory) *iostreams.IOStreams {
	io := iostreams.System()
	cfg, err := f.Config()
	if err != nil {
		return io
	}

	if prompt, _ := cfg.Get("", "prompt"); prompt == "disabled" {
		io.SetNeverPrompt(true)
	}

	// Pager precedence
	// 1. INSTILL_PAGER
	// 2. pager from config
	// 3. PAGER
	if pager, pagerExists := os.LookupEnv("INSTILL_PAGER"); pagerExists {
		io.SetPager(pager)
	} else if pager, _ := cfg.Get("", "pager"); pager != "" {
		io.SetPager(pager)
	}

	return io
}
