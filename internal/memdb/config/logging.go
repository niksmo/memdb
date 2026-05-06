package config

type logging struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

func defaultLogging() logging {
	return logging{
		Level:  "INFO",
		Output: "", // stdout
	}
}

type rawLogging struct {
	Level  *string `mapstructure:"level"`
	Output *string `mapstructure:"output"`
}

func toLogging(raw *rawLogging) logging {
	l := defaultLogging()
	if raw == nil {
		return l
	}

	if raw.Level != nil {
		l.Level = *raw.Level
	}

	if raw.Output != nil {
		l.Output = *raw.Output
	}

	return l
}
