package transport

import (
	"errors"
	"time"
)

type clientOptions struct {
	maxResponseSize int
	readTimeout     time.Duration
	writeTimeout    time.Duration
}

func newClientOptions(opts []ClientOption) (*clientOptions, error) {
	o := &clientOptions{}

	var errs []error
	for _, fn := range opts {
		if err := fn(o); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return o, errors.Join(errs...)
	}

	return o, nil
}

func (o *clientOptions) MaxResponseSize() int {
	if o.maxResponseSize <= 0 {
		return defaultMaxResponseSize
	}
	return o.maxResponseSize
}

func (o *clientOptions) ReadTimeout() time.Duration {
	if o.readTimeout <= 0 {
		return defaultReadTimeout
	}
	return o.readTimeout
}

func (o *clientOptions) WriteTimeout() time.Duration {
	if o.writeTimeout <= 0 {
		return defaultWriteTimeout
	}
	return o.writeTimeout
}

type ClientOption func(o *clientOptions) error

func WithMaxResponseSize(size int) ClientOption {
	return func(o *clientOptions) error {
		if size <= 0 {
			return errors.New("size must be more than zero, apply default")
		}
		o.maxResponseSize = size
		return nil
	}
}

func WithReadTimeout(t time.Duration) ClientOption {
	return func(o *clientOptions) error {
		if t <= 0 {
			return errors.New("read timeout must be more than zero, apply default")
		}
		o.readTimeout = t
		return nil
	}
}

func WithWriteTimeout(t time.Duration) ClientOption {
	return func(o *clientOptions) error {
		if t <= 0 {
			return errors.New("write timeout must be more than zero, apply default")
		}
		o.writeTimeout = t
		return nil
	}
}
