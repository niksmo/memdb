package logger

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func New(wr io.Writer, level string) *slog.Logger {
	logLevel := parseStringLevel(wr, level)

	logHandlerOpts := &slog.HandlerOptions{
		Level:       logLevel,
		ReplaceAttr: convertTimeInUnixMicro,
	}

	handler := slog.NewTextHandler(wr, logHandlerOpts)
	l := slog.New(handler)

	return l
}

func NewObservable(level string) (logger *slog.Logger, observer *strings.Builder) {
	observer = new(strings.Builder)

	return New(observer, level), observer
}

func parseStringLevel(wr io.Writer, level string) slog.Level {
	var logLevel slog.Level

	err := logLevel.UnmarshalText([]byte(level))
	if err != nil {
		fmt.Fprintf(wr, "invalid log level `%s`, expected: debug | info | warn | error\n", level)
		logLevel = slog.LevelInfo
	}

	return logLevel
}

func convertTimeInUnixMicro(_ []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		a.Value = slog.Int64Value(a.Value.Time().UTC().UnixMicro())
	}
	return a
}
