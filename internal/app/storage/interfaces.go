package storage

import (
	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

type Database interface {
	Write(models.ClientData) error
	ReadByCookie(string) (models.ClientData, error)
	ReadByTag(string) (models.ShortData, error)
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
