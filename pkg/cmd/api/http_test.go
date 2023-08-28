package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func Test_httpRequest(t *testing.T) {
	var tr roundTripper = func(req *http.Request) (*http.Response, error) {
		return &http.Response{Request: req}, nil
	}
	httpClient := http.Client{Transport: tr}

	type args struct {
		client  *http.Client
		host    string
		method  string
		p       string
		params  interface{}
		headers []string
	}
	type expects struct {
		method  string
		u       string
		body    string
		headers string
	}
	tests := []struct {
		name    string
		args    args
		want    expects
		wantErr bool
	}{
		{
			name: "simple GET",
			args: args{
				client:  &httpClient,
				host:    "instill.tech",
				method:  "GET",
				p:       "/models",
				params:  nil,
				headers: []string{},
			},
			wantErr: false,
			want: expects{
				method:  "GET",
				u:       "https://api.instill.tech/models",
				body:    "",
				headers: "",
			},
		},
		{
			name: "GET with leading slash",
			args: args{
				client:  &httpClient,
				host:    "instill.tech",
				method:  "GET",
				p:       "/models",
				params:  nil,
				headers: []string{},
			},
			wantErr: false,
			want: expects{
				method:  "GET",
				u:       "https://api.instill.tech/models",
				body:    "",
				headers: "",
			},
		},
		{
			name: "GET with params",
			args: args{
				client: &httpClient,
				host:   "instill.tech",
				method: "GET",
				p:      "models",
				params: map[string]interface{}{
					"a": "b",
				},
				headers: []string{},
			},
			wantErr: false,
			want: expects{
				method:  "GET",
				u:       "https://api.instill.tech/models?a=b",
				body:    "",
				headers: "",
			},
		},
		{
			name: "POST with params",
			args: args{
				client: &httpClient,
				host:   "github.com",
				method: "POST",
				p:      "repos",
				params: map[string]interface{}{
					"a": "b",
				},
				headers: []string{},
			},
			wantErr: false,
			want: expects{
				method:  "POST",
				u:       "https://api.github.com/repos",
				body:    `{"a":"b"}`,
				headers: "Content-Type: application/json; charset=utf-8\r\n",
			},
		},
		{
			name: "POST with body and type",
			args: args{
				client: &httpClient,
				host:   "github.com",
				method: "POST",
				p:      "repos",
				params: bytes.NewBufferString("CUSTOM"),
				headers: []string{
					"content-type: text/plain",
					"accept: application/json",
				},
			},
			wantErr: false,
			want: expects{
				method:  "POST",
				u:       "https://api.github.com/repos",
				body:    `CUSTOM`,
				headers: "Accept: application/json\r\nContent-Type: text/plain\r\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httpRequest(tt.args.client, tt.args.host, tt.args.method, tt.args.p, tt.args.params, tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("httpRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			req := got.Request
			if req.Method != tt.want.method {
				t.Errorf("Request.Method = %q, want %q", req.Method, tt.want.method)
			}
			if req.URL.String() != tt.want.u {
				t.Errorf("Request.URL = %q, want %q", req.URL.String(), tt.want.u)
			}

			if tt.want.body != "" {
				bb, err := io.ReadAll(req.Body)
				if err != nil {
					t.Errorf("Request.Body ReadAll error = %v", err)
					return
				}
				if string(bb) != tt.want.body {
					t.Errorf("Request.Body = %q, want %q", string(bb), tt.want.body)
				}
			}

			h := bytes.Buffer{}
			err = req.Header.WriteSubset(&h, map[string]bool{})
			if err != nil {
				t.Errorf("Request.Header WriteSubset error = %v", err)
				return
			}
			if h.String() != tt.want.headers {
				t.Errorf("Request.Header = %q, want %q", h.String(), tt.want.headers)
			}
		})
	}
}

func Test_addQuery(t *testing.T) {
	type args struct {
		path   string
		params map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": "hello"},
			},
			want: "?a=hello",
		},
		{
			name: "append",
			args: args{
				path:   "path",
				params: map[string]interface{}{"a": "b"},
			},
			want: "path?a=b",
		},
		{
			name: "append query",
			args: args{
				path:   "path?foo=bar",
				params: map[string]interface{}{"a": "b"},
			},
			want: "path?foo=bar&a=b",
		},
		{
			name: "[]byte",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": []byte("hello")},
			},
			want: "?a=hello",
		},
		{
			name: "int",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": 123},
			},
			want: "?a=123",
		},
		{
			name: "nil",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": nil},
			},
			want: "?a=",
		},
		{
			name: "bool",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": true, "b": false},
			},
			want: "?a=true&b=false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addQuery(tt.args.path, tt.args.params); got != tt.want {
				t.Errorf("addQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
