package config

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	mrand "math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env"

	"github.com/t1mon-ggg/go_shortner/app/storage"
)

// Config configuration struct
type Config struct {
	BaseURL         string `env:"BASE_URL" json:"base_url"`                   // BaseURL - default url base.
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address"`       // ServerAddress - adress where http server will start
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"` // FileStoragePath - path file storage
	Database        string `env:"DATABASE_DSN" json:"database_dsn"`           // Database - databse dsn connection string
	Crypto          bool   `env:"ENABLE_HTTPS" json:"enable_https"`           // Crypto - enable https
	Config          string `env:"CONFIG" json:"-"`                            // Config - configuration file path
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`       // TrustesSubnet - trusted subnet
	GRPCAddress     string `env:"GRPC_ADDRESS" json:"grpc_address"`           //GRPCAddress - grpc listner addrtess
}

// NewConfig - создание новой минимальной конфигурации, чтение переменных окружения и флагов коммандной строки
func New() *Config {
	s := Config{
		BaseURL:         "http://127.0.0.1:8080",
		GRPCAddress:     "127.0.0.1:3200",
		ServerAddress:   "127.0.0.1:8080",
		FileStoragePath: "",
		Database:        "",
		Crypto:          false,
	}
	err := s.readEnv()
	if err != nil {
		log.Fatal(err)
	}
	s.readCli()
	if s.Config != "" {
		s.readFile()
	}
	resultconfig := fmt.Sprintf("Result config:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\nENABLE_HTTPS=%v\nCONFIG=%s\nTRUSTED_SUBNET=%s\n", s.BaseURL, s.ServerAddress, s.FileStoragePath, s.Database, s.Crypto, s.Config, s.TrustedSubnet)
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
	if c.Crypto {
		cfg.Crypto = c.Crypto
	}
	if c.Config != "" {
		cfg.Config = c.Config
	}
	if c.TrustedSubnet != "" {
		cfg.TrustedSubnet = c.TrustedSubnet
	}
	if c.GRPCAddress != "" {
		cfg.GRPCAddress = c.GRPCAddress
	}
	parsed := fmt.Sprintf("Evironment parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\nENABLE_HTTPS=%v\nCONFIG=%s\nTRUSTED_SUBNET=%s\n", c.BaseURL, c.ServerAddress, c.FileStoragePath, c.Database, c.Crypto, c.Config, c.TrustedSubnet)
	log.Println(parsed)
	return nil
}

// flags - map for flag iterate
var flags = map[string]string{
	"b": "BASE_URL",
	"a": "SERVER_ADDRESS",
	"f": "FILE_STORAGE_PATH",
	"d": "DATABASE_DSN",
	"s": "ENABLE_HTTPS",
	"c": "CONFIG",
	"t": "TRUSTED_SUBNET",
	"g": "GRPCAddress",
}

// command line flags
var (
	baseURL     = flag.String("b", "", flags["b"])
	srvAddr     = flag.String("a", "", flags["a"])
	grpcAddr    = flag.String("g", "", flags["g"])
	filePath    = flag.String("f", "", flags["f"])
	dbPath      = flag.String("d", "", flags["d"])
	crypt       = flag.Bool("s", false, flags["s"])
	confFile    = flag.String("c", "", flags["c"])
	trustSubnet = flag.String("t", "", flags["t"])
)

func (cfg *Config) readFile() {
	config, err := os.ReadFile(cfg.Config)
	if err != nil {
		log.Fatalln("configuration file path or name invalid", err)
	}
	c := Config{}
	err = json.Unmarshal(config, &c)
	if err != nil {
		log.Fatalln("configuration is invalid", err)
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
	if c.Crypto {
		cfg.Crypto = c.Crypto
	}
	if c.TrustedSubnet != "" {
		cfg.TrustedSubnet = c.TrustedSubnet
	}
	if c.GRPCAddress != "" {
		cfg.GRPCAddress = c.GRPCAddress
	}
	parsed := fmt.Sprintf("File parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\nENABLE_HTTPS=%v\nCONFIG=%s\nTRUSTED_SUBNET=%s\n", c.BaseURL, c.ServerAddress, c.FileStoragePath, c.Database, c.Crypto, c.Config, c.TrustedSubnet)
	log.Println(parsed)
}

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
			case "ENABLE_HTTPS":
				cfg.Crypto = *crypt
			case "CONFIG":
				cfg.Config = *confFile
			case "TRUSTED_SUBNET":
				cfg.TrustedSubnet = *trustSubnet
			case "GRPCAddress":
				cfg.GRPCAddress = *grpcAddr
			}
		}
	}
	parsed := fmt.Sprintf("Flags parsed:\nBASE_URL=%s\nSERVER_ADDRESS=%s\nFILE_STORAGE_PATH=%s\nDATABASE_DSN=%s\nENABLE_HTTPS=%v\nCONFIG=%s\nTRUSTED_SUBNET=%s\n", *baseURL, *srvAddr, *filePath, *dbPath, *crypt, *confFile, *trustSubnet)
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

func (cfg *Config) NewListner(done <-chan os.Signal, stop chan struct{}, wg *sync.WaitGroup, handler http.Handler) error {
	listenerErr := make(chan error)
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: handler,
	}
	if cfg.Crypto {
		go func(errCh chan error) {
			err := GenerateCerts()
			if err != nil {
				log.Fatalln("certificate error", err)
			}
			if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}(listenerErr)
		log.Println("TLS Server Started")
	} else {
		go func(errCh chan error) {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}(listenerErr)
		log.Println("Server Started")
	}
	select {
	case sig := <-done:
		log.Println("recives os signal", sig)
		close(stop)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60*time.Second))
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println("Server Stopped with", err)
			os.Exit(1)
		} else {
			wg.Wait()
			log.Print("Server Stopped Gracefully")
		}
	case err := <-listenerErr:
		return err
	}
	log.Println("Listner stopped")
	return nil
}

// GenerateCerts - generate cert and key file in PEM format
func GenerateCerts() error {
	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(mrand.Int63()),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(crand.Reader, 4096)
	if err != nil {
		return err
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(crand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.WriteFile("cert.pem", certPEM.Bytes(), 0664)
	if err != nil {
		return err
	}
	err = os.WriteFile("key.pem", privateKeyPEM.Bytes(), 0600)
	if err != nil {
		return err
	}
	return nil
}
