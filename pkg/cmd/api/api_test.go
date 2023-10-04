package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/export"
	"github.com/instill-ai/cli/pkg/iostreams"
)

func Test_NewCmdApi(t *testing.T) {
	f := &cmdutil.Factory{
		Config: config.ConfigStubFactory,
	}

	tests := []struct {
		name     string
		cli      string
		wants    ApiOptions
		wantsErr bool
	}{
		{
			name: "override method",
			cli:  "pipelines -XDELETE",
			wants: ApiOptions{
				Hostname:            "api.instill.tech",
				RequestMethod:       "DELETE",
				RequestMethodPassed: true,
				RequestPath:         "pipelines",
				RequestInputFile:    "",
				RawFields:           []string(nil),
				MagicFields:         []string(nil),
				RequestHeaders:      []string(nil),
				ShowResponseHeaders: false,
				Silent:              false,
				CacheTTL:            0,
				Template:            "",
				FilterOutput:        "",
			},
			wantsErr: false,
		},
		{
			name: "with headers",
			cli:  "user -H 'accept: text/plain' -i",
			wants: ApiOptions{
				Hostname:            "api.instill.tech",
				RequestMethod:       "GET",
				RequestMethodPassed: false,
				RequestPath:         "user",
				RequestInputFile:    "",
				RawFields:           []string(nil),
				MagicFields:         []string(nil),
				RequestHeaders:      []string{"accept: text/plain"},
				ShowResponseHeaders: true,
				Silent:              false,
				CacheTTL:            0,
				Template:            "",
				FilterOutput:        "",
			},
			wantsErr: false,
		},
		{
			name: "with silenced output",
			cli:  "models --silent",
			wants: ApiOptions{
				Hostname:            "api.instill.tech",
				RequestMethod:       "GET",
				RequestMethodPassed: false,
				RequestPath:         "models",
				RequestInputFile:    "",
				RawFields:           []string(nil),
				MagicFields:         []string(nil),
				RequestHeaders:      []string(nil),
				ShowResponseHeaders: false,
				Silent:              true,
				CacheTTL:            0,
				Template:            "",
				FilterOutput:        "",
			},
			wantsErr: false,
		},
		{
			name: "with request body from file",
			cli:  "user --input myfile",
			wants: ApiOptions{
				Hostname:            "api.instill.tech",
				RequestMethod:       "GET",
				RequestMethodPassed: false,
				RequestPath:         "user",
				RequestInputFile:    "myfile",
				RawFields:           []string(nil),
				MagicFields:         []string(nil),
				RequestHeaders:      []string(nil),
				ShowResponseHeaders: false,
				Silent:              false,
				CacheTTL:            0,
				Template:            "",
				FilterOutput:        "",
			},
			wantsErr: false,
		},
		{
			name:     "no arguments",
			cli:      "",
			wantsErr: true,
		},
		{
			name: "with cache",
			cli:  "user --cache 5m",
			wants: ApiOptions{
				Hostname:            "api.instill.tech",
				RequestMethod:       "GET",
				RequestMethodPassed: false,
				RequestPath:         "user",
				RequestInputFile:    "",
				RawFields:           []string(nil),
				MagicFields:         []string(nil),
				RequestHeaders:      []string(nil),
				ShowResponseHeaders: false,
				Silent:              false,
				CacheTTL:            time.Minute * 5,
				Template:            "",
				FilterOutput:        "",
			},
			wantsErr: false,
		},
		{
			name: "with template",
			cli:  "user -t 'hello {{.name}}'",
			wants: ApiOptions{
				Hostname:            "api.instill.tech",
				RequestMethod:       "GET",
				RequestMethodPassed: false,
				RequestPath:         "user",
				RequestInputFile:    "",
				RawFields:           []string(nil),
				MagicFields:         []string(nil),
				RequestHeaders:      []string(nil),
				ShowResponseHeaders: false,
				Silent:              false,
				CacheTTL:            0,
				Template:            "hello {{.name}}",
				FilterOutput:        "",
			},
			wantsErr: false,
		},
		{
			name: "with jq filter",
			cli:  "user -q .name",
			wants: ApiOptions{
				Hostname:            "api.instill.tech",
				RequestMethod:       "GET",
				RequestMethodPassed: false,
				RequestPath:         "user",
				RequestInputFile:    "",
				RawFields:           []string(nil),
				MagicFields:         []string(nil),
				RequestHeaders:      []string(nil),
				ShowResponseHeaders: false,
				Silent:              false,
				CacheTTL:            0,
				Template:            "",
				FilterOutput:        ".name",
			},
			wantsErr: false,
		},
		{
			name:     "--silent with --jq",
			cli:      "user --silent -q .foo",
			wantsErr: true,
		},
		{
			name:     "--silent with --template",
			cli:      "user --silent -t '{{.foo}}'",
			wantsErr: true,
		},
		{
			name:     "--jq with --template",
			cli:      "user --jq .foo -t '{{.foo}}'",
			wantsErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts *ApiOptions
			cmd := NewCmdApi(f, func(o *ApiOptions) error {
				opts = o
				return nil
			})

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.Hostname, opts.Hostname)
			assert.Equal(t, tt.wants.RequestMethod, opts.RequestMethod)
			assert.Equal(t, tt.wants.RequestMethodPassed, opts.RequestMethodPassed)
			assert.Equal(t, tt.wants.RequestPath, opts.RequestPath)
			assert.Equal(t, tt.wants.RequestInputFile, opts.RequestInputFile)
			assert.Equal(t, tt.wants.RawFields, opts.RawFields)
			assert.Equal(t, tt.wants.MagicFields, opts.MagicFields)
			assert.Equal(t, tt.wants.RequestHeaders, opts.RequestHeaders)
			assert.Equal(t, tt.wants.ShowResponseHeaders, opts.ShowResponseHeaders)
			assert.Equal(t, tt.wants.Silent, opts.Silent)
			assert.Equal(t, tt.wants.CacheTTL, opts.CacheTTL)
			assert.Equal(t, tt.wants.Template, opts.Template)
			assert.Equal(t, tt.wants.FilterOutput, opts.FilterOutput)
		})
	}
}

func Test_apiRun(t *testing.T) {
	tests := []struct {
		name         string
		options      ApiOptions
		httpResponse *http.Response
		err          error
		stdout       string
		stderr       string
	}{
		{
			name: "success",
			httpResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`bam!`)),
			},
			err:    nil,
			stdout: `bam!`,
			stderr: ``,
		},
		{
			name: "show response headers",
			options: ApiOptions{
				ShowResponseHeaders: true,
			},
			httpResponse: &http.Response{
				Proto:      "HTTP/1.1",
				Status:     "200 Okey-dokey",
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`body`)),
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
			},
			err:    nil,
			stdout: "HTTP/1.1 200 Okey-dokey\nContent-Type: text/plain\r\n\r\nbody",
			stderr: ``,
		},
		{
			name: "success 204",
			httpResponse: &http.Response{
				StatusCode: 204,
				Body:       nil,
			},
			err:    nil,
			stdout: ``,
			stderr: ``,
		},
		{
			name: "REST error",
			httpResponse: &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewBufferString(`{"message": "THIS IS FINE"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			},
			err:    cmdutil.SilentError,
			stdout: `{"message": "THIS IS FINE"}`,
			stderr: "inst: THIS IS FINE (HTTP 400)\n",
		},
		{
			name: "REST string errors",
			httpResponse: &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewBufferString(`{"errors": ["ALSO", "FINE"]}`)),
				Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			},
			err:    cmdutil.SilentError,
			stdout: `{"errors": ["ALSO", "FINE"]}`,
			stderr: "inst: ALSO\nFINE\n",
		},
		{
			name: "failure",
			httpResponse: &http.Response{
				StatusCode: 502,
				Body:       io.NopCloser(bytes.NewBufferString(`gateway timeout`)),
			},
			err:    cmdutil.SilentError,
			stdout: `gateway timeout`,
			stderr: "inst: HTTP 502\n",
		},
		{
			name: "silent",
			options: ApiOptions{
				Silent: true,
			},
			httpResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`body`)),
			},
			err:    nil,
			stdout: ``,
			stderr: ``,
		},
		{
			name: "show response headers even when silent",
			options: ApiOptions{
				ShowResponseHeaders: true,
				Silent:              true,
			},
			httpResponse: &http.Response{
				Proto:      "HTTP/1.1",
				Status:     "200 Okey-dokey",
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`body`)),
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
			},
			err:    nil,
			stdout: "HTTP/1.1 200 Okey-dokey\nContent-Type: text/plain\r\n\r\n",
			stderr: ``,
		},
		{
			name: "output template",
			options: ApiOptions{
				Template: `{{.status}}`,
			},
			httpResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"not a cat"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
			err:    nil,
			stdout: "not a cat",
			stderr: ``,
		},
		{
			name: "jq filter",
			options: ApiOptions{
				FilterOutput: `.[].name`,
			},
			httpResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"name":"Mona"},{"name":"Hubot"}]`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
			err:    nil,
			stdout: "Mona\nHubot\n",
			stderr: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, _, stdout, stderr := iostreams.Test()

			tt.options.IO = stream
			tt.options.Config = config.ConfigStubFactory
			tt.options.HTTPClient = func() (*http.Client, error) {
				var tr roundTripper = func(req *http.Request) (*http.Response, error) {
					resp := tt.httpResponse
					resp.Request = req
					return resp, nil
				}
				return &http.Client{Transport: tr}, nil
			}

			err := apiRun(&tt.options)
			if err != tt.err {
				t.Errorf("expected error %v, got %v", tt.err, err)
			}

			if stdout.String() != tt.stdout {
				t.Errorf("expected output %q, got %q", tt.stdout, stdout.String())
			}
			if stderr.String() != tt.stderr {
				t.Errorf("expected error output %q, got %q", tt.stderr, stderr.String())
			}
		})
	}
}

func Test_apiRun_inputFile(t *testing.T) {
	tests := []struct {
		name          string
		inputFile     string
		inputContents []byte

		contentLength    int64
		expectedContents []byte
	}{
		{
			name:          "stdin",
			inputFile:     "-",
			inputContents: []byte("I WORK OUT"),
			contentLength: 10,
		},
		{
			name:          "from file",
			inputFile:     "instill-test-file",
			inputContents: []byte("I WORK OUT"),
			contentLength: 10,
		},
	}

	tempDir := t.TempDir()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, stdin, _, _ := iostreams.Test()
			resp := &http.Response{StatusCode: 204}

			inputFile := tt.inputFile
			if tt.inputFile == "-" {
				_, _ = stdin.Write(tt.inputContents)
			} else {
				f, err := os.CreateTemp(tempDir, tt.inputFile)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = f.Write(tt.inputContents)
				defer f.Close()
				inputFile = f.Name()
			}

			var bodyBytes []byte
			options := ApiOptions{
				RequestPath:      "/vdp/v1alpha/hello",
				RequestInputFile: inputFile,
				RawFields:        []string{"a=b", "c=d"},

				IO: stream,
				HTTPClient: func() (*http.Client, error) {
					var tr roundTripper = func(req *http.Request) (*http.Response, error) {
						var err error
						if bodyBytes, err = io.ReadAll(req.Body); err != nil {
							return nil, err
						}
						resp.Request = req
						return resp, nil
					}
					return &http.Client{Transport: tr}, nil
				},
				Config: config.ConfigStubFactory,
			}

			err := apiRun(&options)
			if err != nil {
				t.Errorf("got error %v", err)
			}

			assert.Equal(t, "POST", resp.Request.Method)
			assert.Equal(t, "/vdp/v1alpha/hello?a=b&c=d", resp.Request.URL.RequestURI())
			assert.Equal(t, tt.contentLength, resp.Request.ContentLength)
			assert.Equal(t, "", resp.Request.Header.Get("Content-Type"))
			assert.Equal(t, tt.inputContents, bodyBytes)
		})
	}
}

func Test_apiRun_cache(t *testing.T) {
	stream, _, stdout, stderr := iostreams.Test()

	requestCount := 0
	options := ApiOptions{
		IO: stream,
		HTTPClient: func() (*http.Client, error) {
			var tr roundTripper = func(req *http.Request) (*http.Response, error) {
				requestCount++
				return &http.Response{
					Request:    req,
					StatusCode: 204,
				}, nil
			}
			return &http.Client{Transport: tr}, nil
		},
		Config: config.ConfigStubFactory,

		RequestPath: "pipelines",
		CacheTTL:    time.Minute,
	}

	t.Cleanup(func() {
		cacheDir := filepath.Join(os.TempDir(), "instill-cli-cache")
		os.RemoveAll(cacheDir)
	})

	err := apiRun(&options)
	assert.NoError(t, err)
	err = apiRun(&options)
	assert.NoError(t, err)

	assert.Equal(t, 1, requestCount)
	assert.Equal(t, "", stdout.String(), "stdout")
	assert.Equal(t, "", stderr.String(), "stderr")
}

func Test_parseFields(t *testing.T) {
	stream, stdin, _, _ := iostreams.Test()
	fmt.Fprint(stdin, "pasted contents")

	opts := ApiOptions{
		IO: stream,
		RawFields: []string{
			"robot=Hubot",
			"destroyer=false",
			"helper=true",
			"location=@work",
		},
		MagicFields: []string{
			"input=@-",
			"enabled=true",
			"victories=123",
		},
	}

	params, err := parseFields(&opts)
	if err != nil {
		t.Fatalf("parseFields error: %v", err)
	}

	expect := map[string]interface{}{
		"robot":     "Hubot",
		"destroyer": "false",
		"helper":    "true",
		"location":  "@work",
		"input":     []byte("pasted contents"),
		"enabled":   true,
		"victories": 123,
	}
	assert.Equal(t, expect, params)
}

func Test_magicFieldValue(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "instill-test")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fmt.Fprint(f, "file contents")

	stream, _, _, _ := iostreams.Test()

	type args struct {
		v    string
		opts *ApiOptions
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "string",
			args:    args{v: "hello"},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "bool true",
			args:    args{v: "true"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "bool false",
			args:    args{v: "false"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "null",
			args:    args{v: "null"},
			want:    nil,
			wantErr: false,
		},
		{
			name: "file",
			args: args{
				v:    "@" + f.Name(),
				opts: &ApiOptions{IO: stream},
			},
			want:    []byte("file contents"),
			wantErr: false,
		},
		{
			name: "file error",
			args: args{
				v:    "@",
				opts: &ApiOptions{IO: stream},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := magicFieldValue(tt.args.v, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("magicFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_openUserFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "instill-test")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fmt.Fprint(f, "file contents")

	file, length, err := openUserFile(f.Name(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	fb, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(13), length)
	assert.Equal(t, "file contents", string(fb))
}

func Test_processResponse_template(t *testing.T) {
	stream, _, stdout, stderr := iostreams.Test()

	resp := http.Response{
		StatusCode: 200,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(`[
			{
				"title": "First title",
				"labels": [{"name":"bug"}, {"name":"help wanted"}]
			},
			{
				"title": "Second but not last"
			},
			{
				"title": "Alas, tis' the end",
				"labels": [{}, {"name":"feature"}]
			}
		]`)),
	}

	opts := ApiOptions{
		IO:       stream,
		Template: `{{range .}}{{.title}} ({{.labels | pluck "name" | join ", " }}){{"\n"}}{{end}}`,
	}
	template := export.NewTemplate(stream, opts.Template)
	err := processResponse(&resp, &opts, io.Discard, &template)
	require.NoError(t, err)

	err = template.End()
	require.NoError(t, err)

	assert.Equal(t, heredoc.Doc(`
		First title (bug, help wanted)
		Second but not last ()
		Alas, tis' the end (, feature)
	`), stdout.String())
	assert.Equal(t, "", stderr.String())
}
