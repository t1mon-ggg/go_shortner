package storage

type Database interface {
	OpenFile(string) error
	Write(map[string]string) error
	Read() (map[string]string, error)
	Close() error
}
