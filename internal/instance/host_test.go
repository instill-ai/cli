package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{
			host: "Instill.tech",
			want: "instill.tech",
		},
		{
			host: "api.instill.tech",
			want: "instill.tech",
		},
		{
			host: "ssh.instill.tech",
			want: "instill.tech",
		},
		{
			host: "upload.instill.tech",
			want: "instill.tech",
		},
		{
			host: "Instill.localhost",
			want: "instill.localhost",
		},
		{
			host: "api.instill.localhost",
			want: "instill.localhost",
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := ExtractHostname(tt.host); got != tt.want {
				t.Errorf("NormalizeHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostnameValidator(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantsErr bool
	}{
		{
			name:     "valid hostname",
			input:    "internal.instance",
			wantsErr: false,
		},
		{
			name:     "hostname with slashes",
			input:    "//internal.instance",
			wantsErr: true,
		},
		{
			name:     "empty hostname",
			input:    "   ",
			wantsErr: true,
		},
		{
			name:     "hostname with colon",
			input:    "internal.instance:2205",
			wantsErr: true,
		},
		{
			name:     "non-string hostname",
			input:    62,
			wantsErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HostnameValidator(tt.input)
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRESTPrefix(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{
			host: "instill.tech",
			want: "https://api.instill.tech/",
		},
		{
			host: "instill.localhost",
			want: "http://api.instill.localhost/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := RESTPrefix(tt.host); got != tt.want {
				t.Errorf("RESTPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
