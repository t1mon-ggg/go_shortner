package config

type Config interface {
	Read() error
}
