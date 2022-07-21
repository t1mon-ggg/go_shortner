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

func TestConfig_NewStorage(t *testing.T) {
	type fields struct {
		BaseURL         string
		ServerAddress   string
		FileStoragePath string
		Database        string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "db",
			fields: fields{
				BaseURL:       "http://127.0.0.1:8080",
				ServerAddress: "127.0.0.1:8080",
				Database:      "postgresql://postgres:postgrespw@127.0.0.1:5432/praktikum?sslmode=disable",
			},
		},
		{
			name: "file",
			fields: fields{
				BaseURL:         "http://127.0.0.1:8080",
				ServerAddress:   "127.0.0.1:8080",
				FileStoragePath: "removeme.txt",
			},
		},
		{
			name: "inmem",
			fields: fields{
				BaseURL:       "http://127.0.0.1:8080",
				ServerAddress: "127.0.0.1:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				BaseURL:         tt.fields.BaseURL,
				ServerAddress:   tt.fields.ServerAddress,
				FileStoragePath: tt.fields.FileStoragePath,
				Database:        tt.fields.Database,
			}
			_, err := cfg.NewStorage()
			require.NoError(t, err)
		})
	}
}
