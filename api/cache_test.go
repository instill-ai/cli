package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CacheResponse(t *testing.T) {
	counter := 0
	fakeHTTP := funcTripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			counter++
			body := fmt.Sprintf("%d: %s %s", counter, req.Method, req.URL.String())
			status := 200
			if req.URL.Path == "/error" {
				status = 500
			}
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}

	cacheDir := filepath.Join(t.TempDir(), "instill-cli-cache")
	httpClient := NewHTTPClient(ReplaceTripper(fakeHTTP), CacheResponse(time.Minute, cacheDir))

	do := func(method, url string, body io.Reader) (string, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return "", err
		}
		res, err := httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			err = fmt.Errorf("ReadAll: %w", err)
		}
		return string(resBody), err
	}

	var res string
	var err error

	res, err = do("GET", "http://example.com/path", nil)
	require.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res)
	res, err = do("GET", "http://example.com/path", nil)
	require.NoError(t, err)
	assert.Equal(t, "1: GET http://example.com/path", res)

	res, err = do("GET", "http://example.com/path2", nil)
	require.NoError(t, err)
	assert.Equal(t, "2: GET http://example.com/path2", res)

	res, err = do("POST", "http://example.com/path2", nil)
	require.NoError(t, err)
	assert.Equal(t, "3: POST http://example.com/path2", res)

	res, err = do("GET", "http://example.com/error", nil)
	require.NoError(t, err)
	assert.Equal(t, "4: GET http://example.com/error", res)
	res, err = do("GET", "http://example.com/error", nil)
	require.NoError(t, err)
	assert.Equal(t, "5: GET http://example.com/error", res)
}
