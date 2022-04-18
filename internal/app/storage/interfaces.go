package storage

import (
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
)

type Database interface {
	OpenFile(string) error
	Write(helpers.Data) error
	Read() (helpers.Data, error)
	Close() error
}
