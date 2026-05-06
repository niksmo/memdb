package config

import "time"

type network struct {
	Address        string        `yaml:"address"`
	MaxConnections int           `yaml:"max_connections"`
	MaxMessageSize int           `yaml:"max_message_size"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
}

func defaultNetwork() network {
	return network{
		Address:        "127.0.0.1:8000",
		MaxConnections: 100,
		MaxMessageSize: 4 << 10,
		IdleTimeout:    5 * time.Minute,
	}
}

type rawNetwork struct {
	Address        *string        `mapstructure:"address"`
	MaxConnections *int           `mapstructure:"max_connections"`
	MaxMessageSize *int           `mapstructure:"max_message_size"`
	IdleTimeout    *time.Duration `mapstructure:"idle_timeout"`
}

func toNetwork(raw *rawNetwork) network {
	n := defaultNetwork()
	if raw == nil {
		return n
	}

	if raw.Address != nil {
		n.Address = *raw.Address
	}

	if raw.MaxConnections != nil {
		n.MaxConnections = *raw.MaxConnections
	}

	if raw.MaxMessageSize != nil {
		n.MaxMessageSize = *raw.MaxMessageSize
	}

	if raw.IdleTimeout != nil {
		n.IdleTimeout = *raw.IdleTimeout
	}

	return n
}
