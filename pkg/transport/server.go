package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
	"time"
)

type Handler interface {
	Handle(ctx context.Context, p []byte) ([]byte, error)
}

type Config struct {
	Address        string
	IdleTimeout    time.Duration
	MaxConnections int
	MaxMessageSize int
}

type Server struct {
	logger  *slog.Logger
	cfg     Config
	handler Handler

	lis       net.Listener
	connCount atomic.Int32
}

func NewListener(l *slog.Logger, e Handler, c Config) (*Server, error) {
	lis, err := net.Listen("tcp", c.Address)
	if err != nil {
		return nil, err
	}

	s := &Server{
		logger:  l,
		cfg:     c,
		handler: e,
		lis:     lis,
	}

	return s, nil
}

func (s *Server) Close() error {
	return s.lis.Close()
}

func (s *Server) Listen(ctx context.Context) error {
	errCh := s.listen(ctx)

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) listen(ctx context.Context) chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)

		for {
			conn, err := s.lis.Accept()
			if errors.Is(err, net.ErrClosed) { // memdb is closed
				return
			}

			if nerr, ok := errors.AsType[net.Error](err); ok {
				if nerr.Temporary() {
					s.logger.Warn("accept connection", slog.Any("error", nerr))
					continue
				}

				if nerr.Timeout() {
					s.logger.Warn("accept connection", slog.Any("error", nerr))
					continue
				}
			}

			if err != nil {
				ch <- err
				return
			}

			s.connCount.Add(1)
			go s.handleRawConn(ctx, conn)
		}
	}()

	return ch
}

func (s *Server) handleRawConn(ctx context.Context, conn net.Conn) {
	defer s.recover()

	remoteAddr := conn.RemoteAddr().String()
	connLogger := s.logger.With(slog.String("remote_addr", remoteAddr))

	defer s.closeConn(connLogger, conn)

	if s.isConnLimitExceeded(connLogger) {
		return
	}

	maxMsgSize := s.cfg.MaxMessageSize
	r := newReader(conn, maxMsgSize)
	wr := newWriter(conn)

	for {
		if !setReadTimeout(connLogger, conn, s.cfg.IdleTimeout) {
			return
		}

		req, err := r.ReadBytes()
		if err != nil {
			s.handleReadErr(connLogger, wr, err)
			return
		}

		loggingReq := string(req)

		handleStart := time.Now()
		resp, err := s.handler.Handle(ctx, req)
		if err != nil {
			connLogger.Error("failed to handle statement",
				slog.Any("error", err), slog.String("request", loggingReq))

			resp = []byte(err.Error())
		}
		loggingResp := string(resp)

		if !setWriteTimeout(connLogger, conn, s.cfg.IdleTimeout) {
			return
		}

		if _, err := wr.Write(resp); err != nil {
			connLogger.Error("failed to write response",
				slog.Any("error", err), slog.String("request", loggingReq))
			return
		}

		responseTime := time.Since(handleStart).Milliseconds()

		connLogger.Info("inbound request",
			slog.Int64("response_time_ms", responseTime),
			slog.String("request", loggingReq),
			slog.String("response", loggingResp),
		)
	}
}

func (s *Server) isConnLimitExceeded(logger *slog.Logger) bool {
	limit := s.cfg.MaxConnections
	n := int(s.connCount.Load())

	if n > limit {
		logger.Warn("connections limit has been reached", slog.Int("limit", limit))
		return true
	}
	return false
}

func (s *Server) handleReadErr(logger *slog.Logger, wr io.Writer, err error) {
	if errors.Is(err, io.EOF) {
		return
	}

	if errors.Is(err, errMsgSize) {
		msg := fmt.Sprintf("data size exceeds limit %d", s.cfg.MaxMessageSize)

		logger.Warn("data size exceeds limit", slog.Int("limit", s.cfg.MaxMessageSize))

		if _, err := wr.Write([]byte(msg)); err != nil {
			logger.Error("failed to write message", slog.Any("error", err))
		}
		return
	}

	logger.Error("read message", slog.Any("error", err))
}

func (s *Server) closeConn(logger *slog.Logger, conn net.Conn) {
	defer s.connCount.Add(-1)

	if err := conn.Close(); err != nil {
		logger.Error("failed to close connection", slog.Any("error", err))
		return
	}
	logger.Info("connection closed")
}

func (s *Server) recover() {
	if r := recover(); r != nil {
		s.logger.Error("panic recovered", slog.Any("cause", r))
	}
}
