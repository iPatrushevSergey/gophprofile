package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress_Set(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSchema string
		wantHost   string
		wantPort   int
		wantListen string
		wantURL    string
		wantErr    bool
	}{
		{"host:port", "localhost:8080", "http", "localhost", 8080, "localhost:8080", "http://localhost:8080", false},
		{"with http", "http://example.com:443", "http", "example.com", 443, "example.com:443", "http://example.com:443", false},
		{"no port", "localhost", "", "", 0, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Address
			err := a.Set(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSchema, a.Schema)
			assert.Equal(t, tt.wantHost, a.Host)
			assert.Equal(t, tt.wantPort, a.Port)
			assert.Equal(t, tt.wantListen, a.String())
			assert.Equal(t, tt.wantURL, a.URL())
		})
	}
}

func TestDuration_Set(t *testing.T) {
	t.Run("integer seconds", func(t *testing.T) {
		var d Duration
		require.NoError(t, d.Set("60"))
		assert.Equal(t, 60*time.Second, d.Duration)
	})

	t.Run("duration string", func(t *testing.T) {
		var d Duration
		require.NoError(t, d.Set("2s"))
		assert.Equal(t, 2*time.Second, d.Duration)
		assert.Equal(t, "2s", d.String())
	})

	t.Run("invalid", func(t *testing.T) {
		var d Duration
		assert.Error(t, d.Set("abc"))
	})
}

func TestDuration_UnmarshalText(t *testing.T) {
	var d Duration
	require.NoError(t, d.UnmarshalText([]byte("10")))
	assert.Equal(t, 10*time.Second, d.Duration)
	assert.Equal(t, "10s", d.String())
}

func TestAddress_UnmarshalText(t *testing.T) {
	var a Address
	require.NoError(t, a.UnmarshalText([]byte("localhost:9090")))
	assert.Equal(t, 9090, a.Port)
	assert.Equal(t, "localhost:9090", a.String())
	assert.Equal(t, "http://localhost:9090", a.URL())
}
