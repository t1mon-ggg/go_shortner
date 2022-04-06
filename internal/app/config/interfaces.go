package config

type Config interface {
	//NewConfig() *OsVars
	ReadEnv() error
	ReadCli()
}
