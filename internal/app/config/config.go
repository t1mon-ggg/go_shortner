package config

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/caarlos0/env"
)

type Config struct {
	BaseURL         string `env:"BASE_URL"`
	ServerAddress   string `env:"SERVER_ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Database        string `env:"DATABASE_DSN"`
	once            sync.Once
}

//NewConfig - выделение памяти для новой конфигурации
func New() *Config {
	s := Config{
		BaseURL:         "http://127.0.0.1:8080",
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

var flags = map[string]string{
	"b": "BASE_URL",
	"a": "SERVER_ADDRESS",
	"f": "FILE_STORAGE_PATH",
	"d": "DATABASE_DSN",
}

var baseurlptr = flag.String("b", "", flags["b"])
var srvaddrptr = flag.String("a", "", flags["a"])
var fpathptr = flag.String("f", "", flags["f"])
var dbpathptr = flag.String("d", "", flags["d"])

//ReadCli - чтение флагов командной строки
func (cfg *Config) readCli() {
	flag.Parse()
	for flag, info := range flags {
		if isFlagPassed(flag) {
			switch info {
			case "BASE_URL":
				cfg.BaseURL = *baseurlptr
			case "SERVER_ADDRESS":
				cfg.ServerAddress = *srvaddrptr
			case "FILE_STORAGE_PATH":
				cfg.FileStoragePath = *fpathptr
			case "DATABASE_DSN":
				cfg.Database = *dbpathptr
			}
		}
	}
	parsed := fmt.Sprintf("Flags parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\n", *baseurlptr, *srvaddrptr, *fpathptr, *dbpathptr)
	log.Println(parsed)

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
