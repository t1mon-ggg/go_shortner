package storage

var Database interface {
	Write(map[string]string) error
	Read() (map[string]string, error)
	Close() error
}
