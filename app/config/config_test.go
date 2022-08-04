package config_test

import (
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t1mon-ggg/go_shortner/app/config"
	"github.com/t1mon-ggg/go_shortner/app/webhandlers"
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
			cfg := config.New()
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
		want   func()
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
			cfg := &config.Config{
				BaseURL:         tt.fields.BaseURL,
				ServerAddress:   tt.fields.ServerAddress,
				FileStoragePath: tt.fields.FileStoragePath,
				Database:        tt.fields.Database,
			}
			s, err := cfg.NewStorage()
			if err != nil {
				if tt.fields.Database != "" {
					t.Logf("Ignorred error for DSN: %v", err)
				}
			} else {
				require.NotNil(t, s)
				require.NoError(t, err)
			}
		})
	}
}

func Test_generateCerts(t *testing.T) {
	t.Run("Generate random cert/key pair", func(t *testing.T) {
		err := config.GenerateCerts()
		require.NoError(t, err)
		var cert, key bool
		_, err = os.Stat("cert.pem")
		if err == nil {
			cert = true
		}
		_, err = os.Stat("key.pem")
		if err == nil {
			key = true
		}
		require.True(t, cert)
		require.True(t, key)
	})
}

func TestConfig_NewListner(t *testing.T) {
	var secureAppErr, clearAppErr error
	type fields struct {
		BaseURL         string
		ServerAddress   string
		FileStoragePath string
		Database        string
		Crypto          bool
	}
	tests := []struct {
		name   string
		fields fields
		appErr error
	}{
		{
			name: "clear",
			fields: fields{
				BaseURL:       "http://127.0.0.1:8888",
				ServerAddress: "127.0.0.1:8888",
				Crypto:        false,
			},
			appErr: clearAppErr,
		},
		{
			name: "secure",
			fields: fields{
				BaseURL:       "http://127.0.0.1:8443",
				ServerAddress: "127.0.0.1:8443",
				Crypto:        true,
			},
			appErr: secureAppErr,
		},
	}
	var ends []bool
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := webhandlers.App{}
			app.Config = &config.Config{
				BaseURL:         tt.fields.BaseURL,
				ServerAddress:   tt.fields.ServerAddress,
				FileStoragePath: tt.fields.FileStoragePath,
				Database:        tt.fields.Database,
				Crypto:          tt.fields.Crypto,
			}
			var err error
			app.Storage, err = app.Config.NewStorage()
			require.NoError(t, err)
			r := app.NewWebProcessor(10)
			go func() {
				tt.appErr = app.Config.NewListner(r)
				ends = append(ends, true)
			}()
		})
	}
	time.Sleep(10 * time.Second)
	log.Println("SIGTEM")
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	time.Sleep(3 * time.Second)
	for _, status := range ends {
		require.True(t, status)
	}
	require.NoError(t, clearAppErr)
	require.NoError(t, secureAppErr)
}
