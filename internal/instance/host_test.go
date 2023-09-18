package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostnameValidator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantsErr bool
	}{
		{
			name:     "valid hostname",
			input:    "internal.instance",
			wantsErr: false,
		},
		{
			name:     "port number",
			input:    "hostname:123",
			wantsErr: false,
		},
		{
			name:     "empty",
			input:    "",
			wantsErr: true,
		},
		{
			name:     "hostname with slashes",
			input:    "//internal.instance",
			wantsErr: true,
		},
		{
			name:     "whitespace",
			input:    "   ",
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
