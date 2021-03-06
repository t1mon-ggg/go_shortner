package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOsVars_Read(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Test FILE_STORAGE_PATH",
			want: "",
		},
		{
			name: "Test SERVER_ADDRESS",
			want: "127.0.0.1:8080",
		},
		{
			name: "Test BASE_URL",
			want: "http://127.0.0.1:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			switch tt.name {
			case "Test FILE_STORAGE_PATH":
				require.Equal(t, tt.want, cfg.FileStoragePath)
			case "Test SERVER_ADDRESS":
				require.Equal(t, tt.want, cfg.ServerAddress)
			case "Test BASE_URL":
				require.Equal(t, tt.want, cfg.BaseURL)
			}
		})
	}
}
