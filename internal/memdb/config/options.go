package config

import (
	"log/slog"
	"os"

	"github.com/niksmo/memdb/pkg/logger"
)

// YamlProvider defines the interface for accessing configuration data sources.
type YamlProvider interface {
	Bytes() ([]byte, error)
}

type options struct {
	yamlProvider YamlProvider
	logger       *slog.Logger
}

func newOptions(opts []Option) *options {
	o := &options{}
	for _, fn := range opts {
		fn(o)
	}

	return o
}

// YamlProvider returns the configured provider or initializes a default
// provider based on the executable path in args[0].
func (o *options) YamlProvider(args []string) YamlProvider {
	if o.yamlProvider == nil {
		return &yamlProvider{args: args, rootPath: args[0]}
	}
	return o.yamlProvider
}

// Logger returns the configured logger or a default stdout logger
// if none was provided.
func (o *options) Logger() *slog.Logger {
	if o.logger == nil {
		return logger.New(os.Stdout, "info")
	}
	return o.logger
}

// Option defines a functional configuration gate for the Load function.
type Option func(*options)

// WithFileProvider returns an Option that overrides the default file
// discovery logic with a custom implementation.
func WithFileProvider(fp YamlProvider) Option {
	return func(o *options) {
		o.yamlProvider = fp
	}
}

// WithLogger returns an Option that sets a custom logger for the
// configuration loading process.
func WithLogger(l *slog.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}
