package storage

import (
	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

type Database interface {
	Write(map[string]models.WebData) error
	ReadByCookie(string) (map[string]models.WebData, error)
	ReadByTag(string) (map[string]string, error)
	TagByURL(string) (string, error)
	Close() error
	Ping() error
}

//SetStorage - отпределения типа хранилища и его применение
func SetStorage(cfg *config.Config) (Database, error) {
	if cfg.Database != "" {
		stor, err := NewPostgreSQL(cfg.Database)
		if err != nil {
			return nil, err
		}
		return stor, nil
	}
	if cfg.FileStoragePath != "" {
		stor := NewFileDB(cfg.FileStoragePath)
		return stor, nil
	}
	stor := NewMemDB()
	return stor, nil
}
