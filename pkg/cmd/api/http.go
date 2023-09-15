package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/instill-ai/cli/internal/instance"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func httpRequest(client *http.Client, hostname string, method string, path string, params interface{}, headers []string) (*http.Response, error) {

	requestURL := instance.GetProtocol(hostname) + strings.TrimPrefix(path, "/")

	var body io.Reader
	var bodyIsJSON bool

	switch pp := params.(type) {
	case map[string]interface{}:
		if strings.EqualFold(method, "GET") {
			requestURL = addQuery(requestURL, pp)
		} else {
			if len(pp) == 0 {
				body = nil
				bodyIsJSON = true
			} else {
				for key, value := range pp {
					switch vv := value.(type) {
					case []byte:
						pp[key] = string(vv)
					}
				}
				b, err := json.Marshal(pp)
				if err != nil {
					return nil, fmt.Errorf("error serializing parameters: %w", err)
				}
				body = bytes.NewBuffer(b)
				bodyIsJSON = true
			}
		}
	case io.Reader:
		if data, err := io.ReadAll(pp); err == nil {
			// Check if body data is JSON
			if json.Valid(data) {
				bodyIsJSON = true
			}
			body = bytes.NewBuffer(data)
		}

	case nil:
		body = nil
	default:
		return nil, fmt.Errorf("unrecognized parameters type: %v", params)
	}

	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, err
	}

	for _, h := range headers {
		idx := strings.IndexRune(h, ':')
		if idx == -1 {
			return nil, fmt.Errorf("header %q requires a value separated by ':'", h)
		}
		name, value := h[0:idx], strings.TrimSpace(h[idx+1:])
		if strings.EqualFold(name, "Content-Length") {
			length, err := strconv.ParseInt(value, 10, 0)
			if err != nil {
				return nil, err
			}
			req.ContentLength = length
		} else {
			req.Header.Add(name, value)
		}
	}

	if bodyIsJSON && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	return client.Do(req)
}

func addQuery(path string, params map[string]interface{}) string {
	if len(params) == 0 {
		return path
	}

	query := url.Values{}
	for key, value := range params {
		switch v := value.(type) {
		case string:
			query.Add(key, v)
		case []byte:
			query.Add(key, string(v))
		case nil:
			query.Add(key, "")
		case int:
			query.Add(key, fmt.Sprintf("%d", v))
		case bool:
			query.Add(key, fmt.Sprintf("%v", v))
		default:
			panic(fmt.Sprintf("unknown type %v", v))
		}
	}

	sep := "?"
	if strings.ContainsRune(path, '?') {
		sep = "&"
	}
	return path + sep + query.Encode()
}
