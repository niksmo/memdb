package config

import "flag"

type Config struct {
	LogLevel string
}

func FromFlags(osArgs []string) Config {
	if len(osArgs) == 0 {
		return defaultConfig()
	}

	flagSet := flag.NewFlagSet("memdb", flag.ExitOnError)

	logLevel := flagSet.String("log", "info", "debug | info | warn | error")

	err := flagSet.Parse(osArgs[1:])
	if err != nil {
		return Config{}
	}

	return Config{LogLevel: *logLevel}
}

func defaultConfig() Config {
	return Config{LogLevel: "info"}
}
