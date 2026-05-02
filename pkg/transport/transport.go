package transport

import (
	"errors"
	"log/slog"
	"net"
	"time"
)

const delim = '\n'

var errMsgSize = errors.New("limit exhausted")

func setReadTimeout(l *slog.Logger, conn net.Conn, delay time.Duration) bool {
	if err := conn.SetReadDeadline(time.Now().Add(delay)); err != nil {
		l.Error("set connection read timeout", slog.Any("error", err))
		return false
	}
	return true
}

func setWriteTimeout(l *slog.Logger, conn net.Conn, delay time.Duration) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(delay)); err != nil {
		l.Error("set connection write timeout", slog.Any("error", err))
		return false
	}
	return true
}
