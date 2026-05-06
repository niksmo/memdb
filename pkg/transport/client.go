package transport

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"
)

const (
	defaultMaxResponseSize = 4 << 10
	defaultReadTimeout     = 5 * time.Minute
	defaultWriteTimeout    = 5 * time.Minute
)

type response struct {
	p   []byte
	err error
}

type Client struct {
	logger       *slog.Logger
	conn         net.Conn
	r            *reader
	w            *writer
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewClient(ctx context.Context, l *slog.Logger, addr string, opts ...ClientOption) (*Client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	options, err := newClientOptions(opts)
	if err != nil {
		l.Error("got invalid client options", slog.Any("error", err))
	}

	c := &Client{
		logger:       l,
		conn:         conn,
		w:            newWriter(conn),
		r:            newReader(conn, options.MaxResponseSize()),
		writeTimeout: options.WriteTimeout(),
		readTimeout:  options.ReadTimeout(),
	}

	return c, nil
}

func (c *Client) Send(ctx context.Context, p []byte) ([]byte, error) {
	if err := c.write(ctx, p); err != nil {
		return nil, err
	}

	resp, err := c.read(ctx)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) write(ctx context.Context, p []byte) error {
	writeErrCh := make(chan error)

	go func() {
		defer close(writeErrCh)

		if !setWriteTimeout(c.logger, c.conn, c.writeTimeout) {
			writeErrCh <- errors.New("set write timeout failed")
			return
		}

		if _, err := c.w.Write(p); err != nil {
			writeErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case err := <-writeErrCh:
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *Client) read(ctx context.Context) ([]byte, error) {
	readRespCh := make(chan response)

	go func() {
		defer close(readRespCh)

		if !setReadTimeout(c.logger, c.conn, c.readTimeout) {
			readRespCh <- response{err: errors.New("set read timeout failed")}
			return
		}

		data, err := c.r.ReadBytes()
		if err != nil {
			readRespCh <- response{err: err}
			return
		}

		readRespCh <- response{p: data}
	}()

	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	case resp := <-readRespCh:
		if err := resp.err; err != nil {
			return nil, err
		}
		return resp.p, nil
	}
}
