package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Vars struct {
	BaseURL         string `env:"BASE_URL"`
	ServerAddress   string `env:"SERVER_ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Database        string `env:"DATABASE_DSN"`
}

func NewConfig() *Vars {
	s := Vars{
		BaseURL:         "http://127.0.0.1:8080",
		ServerAddress:   "127.0.0.1:8080",
		FileStoragePath: "./storage",
		Database:        "",
	}
	return &s
}

func (cfg *Vars) ReadEnv() error {
	var c Vars
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
	if c.Database != "" {
		cfg.FileStoragePath = c.FileStoragePath
	}
	if c.Database != "" {
		cfg.Database = c.Database
	}
	return nil
}

func (cfg *Vars) ReadCli() {
	baseurlptr := flag.String("b", "", "BASE_URL")
	srvaddrptr := flag.String("a", "", "SERVER_ADDRESS")
	fpathptr := flag.String("f", "", "FILE_STORAGE_PATH")
	dbpathptr := flag.String("d", "", "DATABASE_DSN")
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
	if isFlagPassed("d") {
		cfg.Database = *dbpathptr
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
