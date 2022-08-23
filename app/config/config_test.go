package config_test

import (
	"encoding/json"
	"log"
	"os"
	"sync"
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
		},
		{
			name: "secure",
			fields: fields{
				BaseURL:       "http://127.0.0.1:8443",
				ServerAddress: "127.0.0.1:8443",
				Crypto:        true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := sync.WaitGroup{}

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
			wg.Add(1)
			go func(name string, wg *sync.WaitGroup) {
				tt.appErr = app.Config.NewListner(app.Signal(), app.StopSig(), app.Wait(), r)
				log.Printf("Check error in %v test. Err is %v", name, tt.appErr)
				require.NoError(t, tt.appErr)
				wg.Done()
			}(tt.name, &wg)
			time.Sleep(5 * time.Second)
			app.Signal() <- syscall.SIGTERM
			log.Printf("%v sigterm sent", tt.name)
		})
	}
}

func TestConfig_readEnv(t *testing.T) {
	tests := []struct {
		name   string
		fields *config.Config
	}{
		{
			name: "read env",
			fields: &config.Config{
				BaseURL:         "https://localhost",
				ServerAddress:   "localhost:443",
				FileStoragePath: "createme.txtx",
				Database:        "",
				Config:          "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.BaseURL != "" {
				os.Setenv("BASE_URL", tt.fields.BaseURL)
			}
			if tt.fields.BaseURL != "" {
				os.Setenv("SERVER_ADDRESS", tt.fields.ServerAddress)
			}
			if tt.fields.BaseURL != "" {
				os.Setenv("FILE_STORAGE_PATH", tt.fields.FileStoragePath)
			}
			if tt.fields.Database != "" {
				os.Setenv("DATABASE_DSN", tt.fields.Database)
			}
			if tt.fields.Crypto {
				os.Setenv("ENABLE_HTTPS", "true")
			}
			if tt.fields.Config != "" {
				os.Setenv("CONFIG", tt.fields.Config)
			}
			app := webhandlers.NewApp()
			require.Equal(t, tt.fields, app.Config)
		})
	}
}

func TestConfig_readFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want *config.Config
	}{
		{
			name: "test flags",
			args: []string{"-b", "https://localhost", "-a", "localhost:443", "-f", "createme.txt"},
			want: &config.Config{
				BaseURL:         "https://localhost",
				ServerAddress:   "localhost:443",
				FileStoragePath: "createme.txt",
				Database:        "",
				Config:          "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append(os.Args, tt.args...)
			app := webhandlers.NewApp()
			require.Equal(t, tt.want, app.Config)
		})
	}
}

func TestConfig_readFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     *config.Config
	}{
		{
			name:     "test file",
			fileName: "config.json",
			want: &config.Config{
				BaseURL:         "https://localhost",
				ServerAddress:   "localhost:443",
				FileStoragePath: "createme.txt",
				Database:        "",
				Config:          "config.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.OpenFile(tt.fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			require.NoError(t, err)
			jsonConfig, err := json.MarshalIndent(tt.want, "", "  ")
			require.NoError(t, err)
			f.Write(jsonConfig)
			f.Close()
			os.Setenv("CONFIG", tt.fileName)
			app := webhandlers.NewApp()
			require.Equal(t, tt.want, app.Config)
			err = os.Remove(tt.fileName)
			require.NoError(t, err)
		})
	}
}
