package factory

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/instill-ai/cli/pkg/iostreams"
)

func TestNewHTTPClient(t *testing.T) {
	type args struct {
		config     configHTTPClient
		appVersion string
		setAccept  bool
	}
	tests := []struct {
		name       string
		args       args
		envDebug   string
		host       string
		wantHeader map[string]string
		wantStderr string
	}{
		{
			name: "instill.tech with Accept header",
			args: args{
				config:     tinyConfig{"instill.tech:access_token": "MYTOKEN"},
				appVersion: "v1.2.3",
				setAccept:  true,
			},
			host: "instill.tech",
			wantHeader: map[string]string{
				"authorization": "bearer MYTOKEN",
				"user-agent":    "Instill CLI v1.2.3",
				"accept":        "",
			},
			wantStderr: "",
		},
		{
			name: "instill.tech no Accept header",
			args: args{
				config:     tinyConfig{"instill.tech:access_token": "MYTOKEN"},
				appVersion: "v1.2.3",
				setAccept:  false,
			},
			host: "instill.tech",
			wantHeader: map[string]string{
				"authorization": "bearer MYTOKEN",
				"user-agent":    "Instill CLI v1.2.3",
				"accept":        "",
			},
			wantStderr: "",
		},
		{
			name: "instill.tech no access token",
			args: args{
				config:     tinyConfig{"example.com:access_token": "MYTOKEN"},
				appVersion: "v1.2.3",
				setAccept:  true,
			},
			host: "instill.tech",
			wantHeader: map[string]string{
				"authorization": "",
				"user-agent":    "Instill CLI v1.2.3",
			},
			wantStderr: "",
		},
		{
			name: "instill.tech in verbose mode",
			args: args{
				config:     tinyConfig{"instill.tech:access_token": "MYTOKEN"},
				appVersion: "v1.2.3",
				setAccept:  true,
			},
			host:     "instill.tech",
			envDebug: "api",
			wantHeader: map[string]string{
				"authorization": "bearer MYTOKEN",
				"user-agent":    "Instill CLI v1.2.3",
			},
			wantStderr: heredoc.Doc(`
				* Request at <time>
				* Request to http://<host>:<port>
				> GET / HTTP/1.1
				> Host: instill.tech
				> Authorization: bearer ████████████████████
				> User-Agent: Instill CLI v1.2.3

				< HTTP/1.1 204 No Content
				< Date: <time>

				* Request took <duration>
			`),
		},
	}

	var gotReq *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldDebug := os.Getenv("DEBUG")
			os.Setenv("DEBUG", tt.envDebug)
			t.Cleanup(func() {
				os.Setenv("DEBUG", oldDebug)
			})

			io, _, _, stderr := iostreams.Test()
			client, err := NewHTTPClient(io, tt.args.config, tt.args.appVersion, tt.args.setAccept)
			require.NoError(t, err)

			req, err := http.NewRequest("GET", ts.URL, nil)
			req.Host = tt.host
			require.NoError(t, err)

			res, err := client.Do(req)
			require.NoError(t, err)

			for name, value := range tt.wantHeader {
				assert.Equal(t, value, gotReq.Header.Get(name), name)
			}

			assert.Equal(t, 204, res.StatusCode)
			assert.Equal(t, tt.wantStderr, normalizeVerboseLog(stderr.String()))
		})
	}
}

type tinyConfig map[string]string

func (c tinyConfig) Get(host, key string) (string, error) {
	return c[fmt.Sprintf("%s:%s", host, key)], nil
}

func (c tinyConfig) Set(host, key, value string) error {
	return nil
}

func (c tinyConfig) Write() error {
	return nil
}

var requestAtRE = regexp.MustCompile(`(?m)^\* Request at .+`)
var dateRE = regexp.MustCompile(`(?m)^< Date: .+`)
var hostWithPortRE = regexp.MustCompile(`127\.0\.0\.1:\d+`)
var durationRE = regexp.MustCompile(`(?m)^\* Request took .+`)

func normalizeVerboseLog(t string) string {
	t = requestAtRE.ReplaceAllString(t, "* Request at <time>")
	t = hostWithPortRE.ReplaceAllString(t, "<host>:<port>")
	t = dateRE.ReplaceAllString(t, "< Date: <time>")
	t = durationRE.ReplaceAllString(t, "* Request took <duration>")
	return t
}
