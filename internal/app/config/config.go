package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
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
	parsed := fmt.Sprintf("Evironment parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\n", c.BaseURL, c.ServerAddress, c.FileStoragePath, c.Database)
	log.Println(parsed)
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
	parsed := fmt.Sprintf("Flags parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\n", *baseurlptr, *srvaddrptr, *fpathptr, *dbpathptr)
	log.Println(parsed)

}

func (cfg *Vars) SetStorage() (storage.Database, error) {
	if cfg.Database != "" {
		db, err := storage.NewDB(cfg.Database)
		if err != nil {
			return nil, err
		}
		return db, nil
	}
	db := storage.NewFileDB(cfg.FileStoragePath)
	return db, nil
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
