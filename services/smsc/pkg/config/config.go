package config

type Configuration struct {
	Smsc   Smsc
	Logger Logger
}

type Smsc struct {
	StartupTimeout int64
}
