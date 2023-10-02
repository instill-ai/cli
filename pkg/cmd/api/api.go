package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/instill-ai/cli/api"
	"github.com/instill-ai/cli/internal/config"
	"github.com/instill-ai/cli/internal/instance"
	"github.com/instill-ai/cli/pkg/cmdutil"
	"github.com/instill-ai/cli/pkg/export"
	"github.com/instill-ai/cli/pkg/iostreams"
	"github.com/instill-ai/cli/pkg/jsoncolor"
)

type ApiOptions struct {
	IO *iostreams.IOStreams

	Hostname            string
	RequestMethod       string
	RequestMethodPassed bool
	RequestPath         string
	RequestInputFile    string
	MagicFields         []string
	RawFields           []string
	RequestHeaders      []string
	ShowResponseHeaders bool
	Silent              bool
	Template            string
	CacheTTL            time.Duration
	FilterOutput        string

	Config     func() (config.Config, error)
	HTTPClient func() (*http.Client, error)
}

var logger *slog.Logger

func init() {
	var lvl = new(slog.LevelVar)
	if os.Getenv("DEBUG") != "" {
		lvl.Set(slog.LevelDebug)
	} else {
		lvl.Set(slog.LevelError + 1)
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
}

func NewCmdApi(f *cmdutil.Factory, runF func(*ApiOptions) error) *cobra.Command {
	opts := ApiOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HTTPClient: f.HTTPClient,
	}
	// TODO handle error
	cfg, _ := opts.Config()

	cmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make an authenticated Instill API request",
		Long: heredoc.Docf(`
			Makes an authenticated HTTP request to the Instill API and prints the response.

			The endpoint argument should be a path of a Instill API endpoint.

			Note that in some shells, for example PowerShell, you may need to enclose any value that
			contains "{...}" in quotes to prevent the shell from applying special meaning to curly braces.

			The default HTTP request method is "GET" normally and "POST" if any parameters
			were added. Override the method with %[1]s--method%[1]s.

			Pass one or more %[1]s--raw-field%[1]s values in "key=value" format to add string
			parameters to the request payload. To add non-string parameters, see %[1]s--field%[1]s below.
			Note that adding request parameters will automatically switch the request method to POST.
			To send the parameters as a GET query string instead, use %[1]s--method%[1]s GET.

			The %[1]s--field%[1]s flag behaves like %[1]s--raw-field%[1]s with magic type conversion based
			on the format of the value:

			- literal values "true", "false", "null", and integer numbers get converted to
			  appropriate JSON types;
			- if the value starts with "@", the rest of the value is interpreted as a
			  filename to read the value from. Pass "-" to read from standard input.

			Raw request body may be passed from the outside via a file specified by %[1]s--input%[1]s.
			Pass "-" to read from standard input. In this mode, parameters specified via
			%[1]s--field%[1]s flags are serialized into URL query parameters.
		`, "`"),
		Example: heredoc.Doc(`
			# list pipelines
			$ instill api vdp/v1alpha/pipelines

			# list models
			$ instill api model/v1alpha/models

			# get user profile
			$ instill api base/v1alpha/users/me

			# add parameters to a GET request
			$ instill api model/v1alpha/models?visibility=public

			# add nested JSON body to a POST request
			$ jq -n '{"inputs":[{"image": <your image base64 encoded string>}]}' | instill api vdp/v1alpha/pipelines/trigger --input -

			# set a custom HTTP header
			$ instill api -H 'Authorization: Basic ...'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.RequestPath = args[0]
			opts.RequestMethodPassed = c.Flags().Changed("method")

			if c.Flags().Changed("instance") {
				// TODO look for the instance in the config
				if err := instance.HostnameValidator(opts.Hostname); err != nil {
					return cmdutil.FlagErrorf("error parsing `--instance`: %w", err)
				}
			}

			if err := cmdutil.MutuallyExclusive(
				"only one of `--template`, `--jq`, or `--silent` may be used",
				opts.Silent,
				opts.FilterOutput != "",
				opts.Template != "",
			); err != nil {
				return err
			}

			if runF != nil {
				return runF(&opts)
			}
			return apiRun(&opts)
		},
	}

	cmd.Flags().StringVar(&opts.Hostname, "hostname", cfg.DefaultHostname(), "Target instance")
	cmd.Flags().StringVarP(&opts.RequestMethod, "method", "X", "GET", "The HTTP method for the request")
	cmd.Flags().StringArrayVarP(&opts.MagicFields, "field", "F", nil, "Add a typed parameter in `key=value` format")
	cmd.Flags().StringArrayVarP(&opts.RawFields, "raw-field", "f", nil, "Add a string parameter in `key=value` format")
	cmd.Flags().StringArrayVarP(&opts.RequestHeaders, "header", "H", nil, "Add a HTTP request header in `key:value` format")
	cmd.Flags().BoolVarP(&opts.ShowResponseHeaders, "include", "i", false, "Include HTTP response headers in the output")
	cmd.Flags().StringVar(&opts.RequestInputFile, "input", "", "The `file` to use as body for the HTTP request (use \"-\" to read from standard input)")
	cmd.Flags().BoolVar(&opts.Silent, "silent", false, "Do not print the response body")
	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Format the response using a Go template")
	cmd.Flags().StringVarP(&opts.FilterOutput, "jq", "q", "", "Query to select values from the response using jq syntax")
	cmd.Flags().DurationVar(&opts.CacheTTL, "cache", 0, "Cache the response, e.g. \"3600s\", \"60m\", \"1h\"")
	return cmd
}

func apiRun(opts *ApiOptions) error {
	params, err := parseFields(opts)
	if err != nil {
		return err
	}

	// get the host config
	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	var host *config.HostConfigTyped
	if err != nil {
		return err
	}
	hosts, err := cfg.HostsTyped()
	if err != nil {
		return err
	}
	hostname := opts.Hostname
	if hostname == "" {
		hostname = cfg.DefaultHostname()
	}
	for i := range hosts {
		if hosts[i].APIHostname == hostname {
			host = &hosts[i]
			break
		}
	}
	if host == nil {
		return fmt.Errorf(heredoc.Docf(
			`ERROR: instance '%s' does not exist

			You can add it with:
			$ inst instances add %s`,
			hostname, hostname))
	}

	// set up the http client
	method := opts.RequestMethod
	requestPath := opts.RequestPath
	requestHeaders := opts.RequestHeaders
	var requestBody interface{} = params

	if !opts.RequestMethodPassed && (len(params) > 0 || opts.RequestInputFile != "") {
		method = "POST"
	}

	if opts.RequestInputFile != "" {
		file, size, err := openUserFile(opts.RequestInputFile, opts.IO.In)
		if err != nil {
			return err
		}
		defer file.Close()
		requestPath = addQuery(requestPath, params)
		requestBody = file
		if size >= 0 {
			requestHeaders = append([]string{fmt.Sprintf("Content-Length: %d", size)}, requestHeaders...)
		}
	}

	httpClient, err := opts.HTTPClient()
	if err != nil {
		return err
	}
	if opts.CacheTTL > 0 {
		httpClient = api.NewCachedClient(httpClient, opts.CacheTTL)
	}

	headersOutputStream := opts.IO.Out
	if opts.Silent {
		opts.IO.Out = io.Discard
	} else {
		err := opts.IO.StartPager()
		if err != nil {
			return err
		}
		defer opts.IO.StopPager()
	}

	if host.AccessToken != "" {
		requestHeaders = append(requestHeaders, "Authorization: Bearer "+host.AccessToken)
	}

	logger.Debug("api request", "host", host.APIHostname, "path", requestPath)

	// http request & output
	template := export.NewTemplate(opts.IO, opts.Template)
	resp, err := httpRequest(httpClient, host.APIHostname, method, requestPath, requestBody, requestHeaders)
	if err != nil {
		return err
	}
	err = processResponse(resp, opts, headersOutputStream, &template)
	if err != nil {
		return err
	}
	return template.End()
}

func processResponse(resp *http.Response, opts *ApiOptions, headersOutputStream io.Writer, template *export.Template) (err error) {
	if opts.ShowResponseHeaders {
		fmt.Fprintln(headersOutputStream, resp.Proto, resp.Status)
		printHeaders(headersOutputStream, resp.Header, opts.IO.ColorEnabled())
		fmt.Fprint(headersOutputStream, "\r\n")
	}

	if resp.StatusCode == 204 {
		return
	}

	var responseBody io.Reader = resp.Body
	defer resp.Body.Close()

	isJSON, _ := regexp.MatchString(`[/+]json(;|$)`, resp.Header.Get("Content-Type"))

	var serverError string
	if isJSON && resp.StatusCode >= 400 {
		responseBody, serverError, err = parseErrorResponse(responseBody, resp.StatusCode)
		if err != nil {
			return
		}
	}

	if opts.FilterOutput != "" {
		// TODO: reuse parsed query across pagination invocations
		err = export.FilterJSON(opts.IO.Out, responseBody, opts.FilterOutput)
		if err != nil {
			return
		}
	} else if opts.Template != "" {
		// TODO: reuse parsed template across pagination invocations
		err = template.Execute(responseBody)
		if err != nil {
			return
		}
	} else if isJSON && opts.IO.ColorEnabled() {
		err = jsoncolor.Write(opts.IO.Out, responseBody, "  ")
	} else {
		_, err = io.Copy(opts.IO.Out, responseBody)
	}
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			err = nil
		} else {
			return
		}
	}

	if serverError == "" && resp.StatusCode > 299 {
		serverError = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	if serverError != "" {
		fmt.Fprintf(opts.IO.ErrOut, "instill: %s\n", serverError)
		err = cmdutil.SilentError
		return
	}

	return
}

func printHeaders(w io.Writer, headers http.Header, colorize bool) {
	var names []string
	for name := range headers {
		if name == "Status" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	var headerColor, headerColorReset string
	if colorize {
		headerColor = "\x1b[1;34m" // bright blue
		headerColorReset = "\x1b[m"
	}
	for _, name := range names {
		fmt.Fprintf(w, "%s%s%s: %s\r\n", headerColor, name, headerColorReset, strings.Join(headers[name], ", "))
	}
}

func parseFields(opts *ApiOptions) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	for _, f := range opts.RawFields {
		key, value, err := parseField(f)
		if err != nil {
			return params, err
		}
		params[key] = value
	}
	for _, f := range opts.MagicFields {
		key, strValue, err := parseField(f)
		if err != nil {
			return params, err
		}
		value, err := magicFieldValue(strValue, opts)
		if err != nil {
			return params, fmt.Errorf("error parsing %q value: %w", key, err)
		}
		params[key] = value
	}
	return params, nil
}

func parseField(f string) (string, string, error) {
	idx := strings.IndexRune(f, '=')
	if idx == -1 {
		return f, "", fmt.Errorf("field %q requires a value separated by an '=' sign", f)
	}
	return f[0:idx], f[idx+1:], nil
}

func magicFieldValue(v string, opts *ApiOptions) (interface{}, error) {
	if strings.HasPrefix(v, "@") {
		return opts.IO.ReadUserFile(v[1:])
	}

	if n, err := strconv.Atoi(v); err == nil {
		return n, nil
	}

	switch v {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null":
		return nil, nil
	default:
		return v, nil
	}
}

func openUserFile(fn string, stdin io.ReadCloser) (io.ReadCloser, int64, error) {
	if fn == "-" {
		return stdin, -1, nil
	}

	r, err := os.Open(fn)
	if err != nil {
		return r, -1, err
	}

	s, err := os.Stat(fn)
	if err != nil {
		return r, -1, err
	}

	return r, s.Size(), nil
}

func parseErrorResponse(r io.Reader, statusCode int) (io.Reader, string, error) {
	bodyCopy := &bytes.Buffer{}
	b, err := io.ReadAll(io.TeeReader(r, bodyCopy))
	if err != nil {
		return r, "", err
	}

	var parsedBody struct {
		Message string
		Errors  []json.RawMessage
	}
	err = json.Unmarshal(b, &parsedBody)
	if err != nil {
		return r, "", err
	}
	if parsedBody.Message != "" {
		return bodyCopy, fmt.Sprintf("%s (HTTP %d)", parsedBody.Message, statusCode), nil
	}

	type errorMessage struct {
		Message string
	}

	var errors []string
	for _, rawErr := range parsedBody.Errors {
		if len(rawErr) == 0 {
			continue
		}
		if rawErr[0] == '{' {
			var objectError errorMessage
			err := json.Unmarshal(rawErr, &objectError)
			if err != nil {
				return r, "", err
			}
			errors = append(errors, objectError.Message)
		} else if rawErr[0] == '"' {
			var stringError string
			err := json.Unmarshal(rawErr, &stringError)
			if err != nil {
				return r, "", err
			}
			errors = append(errors, stringError)
		}
	}

	if len(errors) > 0 {
		return bodyCopy, strings.Join(errors, "\n"), nil
	}

	return bodyCopy, "", nil
}
