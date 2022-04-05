package config

import (
	"flag"

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

func (cfg *OsVars) Cli() {
	baseurlptr := flag.String("b", "", "BASE_URL")
	srvaddrptr := flag.String("a", "", "SERVER_ADDRESS")
	fpathptr := flag.String("f", "", "FILE_STORAGE_PATH")

	flag.Parse()
	if isFlagPassed("b") {
		cfg.BaseURL = *baseurlptr
	}
	if isFlagPassed("a") {
		cfg.ServerAddress = *srvaddrptr
	}

	if isFlagPassed("f") {
		cfg.FileStoragePath = *fpathptr
	}

}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
