package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	BaseURL         string `env:"BASE_URL"`
	ServerAddress   string `env:"SERVER_ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Database        string `env:"DATABASE_DSN"`
}

//NewConfig - выделение памяти для новой конфигурации
func NewConfig() *Config {
	s := Config{
		BaseURL:         "http://127.0.0.1:8080",
		ServerAddress:   "127.0.0.1:8080",
		FileStoragePath: "",
		Database:        "",
	}
	return &s
}

//ReadEnv - чтение переменных окружения
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

//ReadCli - чтение флагов командной строки
func (cfg *Config) readCli() {
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

//Init - инициализация конфигурации
func (cfg *Config) Init() error {
	err := cfg.readEnv()
	if err != nil {
		return err
	}
	cfg.readCli()
	return nil
}

//isFlagPassed - проверка применение флага
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
