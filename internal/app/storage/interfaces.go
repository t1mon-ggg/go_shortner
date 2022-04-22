package storage

import (
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

type Database interface {
	Write(helpers.Data) error
	ReadByCookie(string) (helpers.Data, error)
	ReadByTag(string) (map[string]string, error)
	Close() error
	Ping() error
}
