package storage

import (
	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
)

type Data interface {
	Write(models.ClientData) error
	ReadByCookie(string) (models.ClientData, error)
	ReadByTag(string) (models.ShortData, error)
	TagByURL(string, string) (string, error)
	Close() error
	Ping() error
	Cleaner(<-chan models.DelWorker, int)
}

//GetStorage - отпределения типа хранилища и его применение
func GetStorage(cfg *config.Config) (Data, error) {
	if cfg.Database != "" {
		stor, err := NewPostgreSQL(cfg.Database)
		if err != nil {
			return nil, err
		}
		return stor, nil
	}
	if cfg.FileStoragePath != "" {
		stor := NewFile(cfg.FileStoragePath)
		return stor, nil
	}
	stor := newRAM()
	return stor, nil
}
