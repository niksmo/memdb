package config

import (
	"bytes"
	"log/slog"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type rawConfig struct {
	Engine  *rawEngine  `mapstructure:"engine"`
	Network *rawNetwork `mapstructure:"network"`
	Logging *rawLogging `mapstructure:"logging"`
}

func toConfig(raw rawConfig) Config {
	return Config{
		Engine:  toEngine(raw.Engine),
		Network: toNetwork(raw.Network),
		Logging: toLogging(raw.Logging),
	}
}

// Config is the configuration object for the application.
type Config struct {
	Engine  engine  `yaml:"engine"`
	Network network `yaml:"network"`
	Logging logging `yaml:"logging"`
}

func defaultConfig() Config {
	return Config{
		Engine:  defaultEngine(),
		Network: defaultNetwork(),
		Logging: defaultLogging(),
	}
}

// String returns a YAML representation of the configuration.
func (c Config) String() string {
	p, _ := yaml.Marshal(c)
	return string(p)
}

// Load initializes the configuration.
// The args parameter typically receives os.Args and is used by the YamlProvider
// to locate configuration files via command-line flags.
//
// If a configuration file cannot be opened or parsed, Load logs the error
// and returns the default configuration.
func Load(args []string, opts ...Option) Config {
	o := newOptions(opts)
	log := o.Logger()
	p := o.YamlProvider(args)

	yamlData, err := p.Bytes()
	if err != nil {
		log.Error("open configuration file", slog.Any("error", err))
		log.Info("default configuration loaded")
		return defaultConfig()
	}

	v := viper.New()
	v.SetConfigType("yml")

	r := bytes.NewReader(yamlData)
	if err := v.ReadConfig(r); err != nil {
		log.Error("read configuration file", slog.Any("error", err))
	}

	var rc rawConfig
	_ = v.Unmarshal(&rc)

	return toConfig(rc)
}
