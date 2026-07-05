package prometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("disabled skips address validation", func(t *testing.T) {
		cfg := Config{Enabled: false, Address: "bad"}
		require.NoError(t, cfg.Validate())
	})

	t.Run("enabled requires address", func(t *testing.T) {
		cfg := Config{Enabled: true}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "address is required")
	})

	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{name: "host and port", address: "localhost:9090"},
		{name: "ip and port", address: "127.0.0.1:9090"},
		{name: "all interfaces port only", address: ":9090"},
		{name: "invalid format", address: "bad", wantErr: true},
		{name: "missing port", address: "localhost", wantErr: true},
		{name: "invalid port", address: "localhost:abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Enabled: true, Address: tt.address}
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.address, cfg.Address)
		})
	}
}
