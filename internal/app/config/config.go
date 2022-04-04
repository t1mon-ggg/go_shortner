package config

import (
	"github.com/caarlos0/env"
)

type OsVars struct {
	BaseURL         string `env:"BASE_URL"`
	ServerAddress   string `env:"SERVER_ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() *OsVars {
	s := &OsVars{
		BaseURL:         "http://127.0.0.1:8080",
		ServerAddress:   "127.0.0.1:8080",
		FileStoragePath: "./storage",
	}
	return s
}

func (cfg *OsVars) Read() error {
	var c OsVars
	err := env.Parse(&c)
	if err != nil {
		return err
	}
	if c.BaseURL != "" {
		cfg.BaseURL = c.BaseURL
	}
	if c.ServerAddress != "" {
		cfg.ServerAddress = c.ServerAddress
	}
	if c.FileStoragePath != "" {
		cfg.FileStoragePath = c.FileStoragePath
	}
	return nil
}
