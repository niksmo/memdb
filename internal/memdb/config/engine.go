package config

type engine struct {
	Type string `yaml:"type"`
}

func defaultEngine() engine {
	return engine{Type: "in_memory"}
}

type rawEngine struct {
	Type *string `mapstructure:"type"`
}

func toEngine(raw *rawEngine) engine {
	e := defaultEngine()
	if raw == nil {
		return e
	}

	if raw.Type != nil {
		e.Type = *raw.Type
	}

	return e
}
