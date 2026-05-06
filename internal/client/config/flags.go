package config

import "flag"

const (
	logLevelName    = "l"
	logLevelDefault = "info"
	logLevelUsage   = `set the logging level ("debug", "info", "warn", "error", "fatal") (default "info")`

	serverAddrName    = "a"
	serverAddrDefault = "127.0.0.1:8000"
	serverAddrUsage   = `set the memdb address (default "127.0.0.1:8000")`
)

type Config struct {
	LogLevel   string
	ServerAddr string
}

func FromFlags(osArgs []string) Config {
	flagSet := flag.NewFlagSet("memdb", flag.ExitOnError)
	logLevel := flagSet.String(logLevelName, logLevelDefault, logLevelUsage)
	serverAddr := flagSet.String(serverAddrName, serverAddrDefault, serverAddrUsage)

	err := flagSet.Parse(osArgs[1:])
	if err != nil {
		return Config{}
	}

	return Config{
		LogLevel:   *logLevel,
		ServerAddr: *serverAddr,
	}
}
