// Build tasks for the Instill CLI project.
//
// Usage:  go run script/build.go [<tasks>...] [<env>...]
//
// Known tasks are:
//
//   bin/instill:
//     Builds the main executable.
//     Supported environment variables:
//     - INSTILL_VERSION: determined from source by default
//     - INSTILL_OAUTH_CLIENT_ID
//     - INSTILL_OAUTH_CLIENT_SECRET
//     - INSTILL_OAUTH_HOSTNAME
//     - INSTILL_OAUTH_AUDIENCE
//     - INSTILL_OAUTH_ISSUER
//     - INSTILL_OAUTH_CALLBACK_HOST
//     - INSTILL_OAUTH_CALLBACK_PORT
//     - SOURCE_DATE_EPOCH: enables reproducible builds
//     - GO_LDFLAGS
//
//   manpages:
//     Builds the man pages under `share/man/man1/`.
//
//   clean:
//     Deletes all built files.
//

package main

import (
	"errors"
	"fmt"
	"github.com/cli/safeexec"
	"github.com/dotenv-org/godotenvvault"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var oauthFields = [][]string{
	{"INSTILL_OAUTH_CLIENT_SECRET", "clientSecret"},
	{"INSTILL_OAUTH_CLIENT_ID", "clientID"},
	{"INSTILL_OAUTH_ISSUER", "issuer"},
	{"INSTILL_OAUTH_HOSTNAME", "hostname"},
	{"INSTILL_OAUTH_AUDIENCE", "audience"},
	{"INSTILL_OAUTH_CALLBACK_HOST", "callbackHost"},
	{"INSTILL_OAUTH_CALLBACK_PORT", "callbackPort"},
}

var tasks = map[string]func(string) error{
	"bin/instill": func(exe string) error {
		info, err := os.Stat(exe)
		if err == nil && !sourceFilesLaterThan(info.ModTime()) {
			fmt.Printf("%s: `%s` is up to date.\n", self, exe)
			return nil
		}

		ldflags := os.Getenv("GO_LDFLAGS")
		ldflags = fmt.Sprintf("-X github.com/instill-ai/cli/internal/build.Version=%s %s", version(), ldflags)
		ldflags = fmt.Sprintf("-X github.com/instill-ai/cli/internal/build.Date=%s %s", date(), ldflags)
		oauthSecret := os.Getenv(oauthFields[0][0])
		if oauthSecret != "" {
			for _, v := range oauthFields {
				nameEnv, nameVar := v[0], v[1]
				ldflags = fmt.Sprintf("-X github.com/instill-ai/cli/internal/oauth2.%s=%s %s",
					nameVar, os.Getenv(nameEnv), ldflags)
			}
		}

		return run("go", "build", "-trimpath", "-ldflags", ldflags, "-o", exe, "./cmd/instill")
	},
	"clean": func(_ string) error {
		return rmrf("bin", "share")
	},
}

var self string

func main() {
	// optionally load the .env file
	_ = godotenvvault.Load()
	args := os.Args[:1]
	for _, arg := range os.Args[1:] {
		if idx := strings.IndexRune(arg, '='); idx >= 0 {
			os.Setenv(arg[:idx], arg[idx+1:])
		} else {
			args = append(args, arg)
		}
	}

	if len(args) < 2 {
		if isWindowsTarget() {
			args = append(args, filepath.Join("bin", "instill.exe"))
		} else {
			args = append(args, "bin/instill")
		}
	}

	self = filepath.Base(args[0])
	if self == "build" {
		self = "build.go"
	}

	for _, task := range args[1:] {
		t := tasks[normalizeTask(task)]
		if t == nil {
			fmt.Fprintf(os.Stderr, "Don't know how to build task `%s`.\n", task)
			os.Exit(1)
		}

		err := t(task)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintf(os.Stderr, "%s: building task `%s` failed.\n", self, task)
			os.Exit(1)
		}
	}
}

func isWindowsTarget() bool {
	if os.Getenv("GOOS") == "windows" {
		return true
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return false
}

func version() string {
	if versionEnv := os.Getenv("INSTILL_VERSION"); versionEnv != "" {
		return versionEnv
	}
	if desc, err := cmdOutput("git", "describe", "--tags"); err == nil {
		return desc
	}
	rev, _ := cmdOutput("git", "rev-parse", "--short", "HEAD")
	return rev
}

func date() string {
	t := time.Now()
	if sourceDate := os.Getenv("SOURCE_DATE_EPOCH"); sourceDate != "" {
		if sec, err := strconv.ParseInt(sourceDate, 10, 64); err == nil {
			t = time.Unix(sec, 0)
		}
	}
	return t.Format("2006-01-02")
}

func sourceFilesLaterThan(t time.Time) bool {
	foundLater := false
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore errors that occur when the project contains a symlink to a filesystem or volume that
			// Windows doesn't have access to.
			if path != "." && isAccessDenied(err) {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				return nil
			}
			return err
		}
		if foundLater {
			return filepath.SkipDir
		}
		if len(path) > 1 && (path[0] == '.' || path[0] == '_') {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if info.IsDir() {
			if name := filepath.Base(path); name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if path == "go.mod" || path == "go.sum" || (strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")) {
			if info.ModTime().After(t) {
				foundLater = true
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return foundLater
}

func isAccessDenied(err error) bool {
	var pe *os.PathError
	// we would use `syscall.ERROR_ACCESS_DENIED` if this script supported build tags
	return errors.As(err, &pe) && strings.Contains(pe.Err.Error(), "Access is denied")
}

func rmrf(targets ...string) error {
	args := append([]string{"rm", "-rf"}, targets...)
	announce(args...)
	for _, target := range targets {
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	}
	return nil
}

func announce(args ...string) {
	fmt.Println(shellInspect(args))
}

func run(args ...string) error {
	exe, err := safeexec.LookPath(args[0])
	if err != nil {
		return err
	}
	announce(args...)
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cmdOutput(args ...string) (string, error) {
	exe, err := safeexec.LookPath(args[0])
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stderr = io.Discard
	out, err := cmd.Output()
	return strings.TrimSuffix(string(out), "\n"), err
}

func shellInspect(args []string) string {
	fmtArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " \t'\"") {
			fmtArgs[i] = fmt.Sprintf("%q", arg)
		} else {
			fmtArgs[i] = arg
		}
	}
	return strings.Join(fmtArgs, " ")
}

func normalizeTask(t string) string {
	return filepath.ToSlash(strings.TrimSuffix(t, ".exe"))
}
