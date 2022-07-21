package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"

	"github.com/t1mon-ggg/go_shortner/app/storage"
)

// Config configuration struct
type Config struct {
	BaseURL         string `env:"BASE_URL"`          // BaseURL - default url base.
	ServerAddress   string `env:"SERVER_ADDRESS"`    // ServerAddress - adress where http server will start
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // FileStoragePath - path file storage
	Database        string `env:"DATABASE_DSN"`      // Database - databse dsn connection string
}

// NewConfig - создание новой минимальной конфигурации, чтение переменных окружения и флагов коммандной строки
func New() *Config {
	s := Config{
		BaseURL:         "http:// 127.0.0.1:8080",
		ServerAddress:   "127.0.0.1:8080",
		FileStoragePath: "",
		Database:        "",
	}
	err := s.readEnv()
	if err != nil {
		log.Fatal(err)
	}
	s.readCli()
	resultconfig := fmt.Sprintf("Result config:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\n", s.BaseURL, s.ServerAddress, s.FileStoragePath, s.Database)
	log.Println(resultconfig)
	return &s
}

// ReadEnv - чтение переменных окружения
func (cfg *Config) readEnv() error {
	var c Config
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
	if c.Database != "" {
		cfg.Database = c.Database
	}
	parsed := fmt.Sprintf("Evironment parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\n", c.BaseURL, c.ServerAddress, c.FileStoragePath, c.Database)
	log.Println(parsed)
	return nil
}

// flags - map for flag iterate
var flags = map[string]string{
	"b": "BASE_URL",
	"a": "SERVER_ADDRESS",
	"f": "FILE_STORAGE_PATH",
	"d": "DATABASE_DSN",
}

// command line flags
var (
	baseURL  = flag.String("b", "", flags["b"])
	srvAddr  = flag.String("a", "", flags["a"])
	filePath = flag.String("f", "", flags["f"])
	dbPath   = flag.String("d", "", flags["d"])
)

// ReadCli - чтение флагов командной строки
func (cfg *Config) readCli() {
	flag.Parse()
	for flag, info := range flags {
		if isFlagPassed(flag) {
			switch info {
			case "BASE_URL":
				cfg.BaseURL = *baseURL
			case "SERVER_ADDRESS":
				cfg.ServerAddress = *srvAddr
			case "FILE_STORAGE_PATH":
				cfg.FileStoragePath = *filePath
			case "DATABASE_DSN":
				cfg.Database = *dbPath
			}
		}
	}
	parsed := fmt.Sprintf("Flags parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\n", *baseURL, *srvAddr, *filePath, *dbPath)
	log.Println(parsed)

}

// isFlagPassed - проверка применение флага
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// NewStorage - создание хранилища
func (cfg *Config) NewStorage() (storage.Storage, error) {
	if cfg.Database != "" {
		s, err := storage.NewPostgreSQL(cfg.Database)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
	if cfg.FileStoragePath != "" {
		stor := storage.NewFile(cfg.FileStoragePath)
		return stor, nil
	}
	s := storage.NewRAM()
	return s, nil
}
