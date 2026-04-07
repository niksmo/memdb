package logger

import (
	"fmt"
	"io"
	"log/slog"
)

func NewText(w io.Writer, level string) *slog.Logger {
	logHandlerOpts := &slog.HandlerOptions{
		Level:       parseStringLevel(level),
		ReplaceAttr: convertTimeInUnixMicro,
	}

	return slog.New(slog.NewTextHandler(w, logHandlerOpts))
}

func parseStringLevel(level string) slog.Level {
	var logLevel slog.Level

	err := logLevel.UnmarshalText([]byte(level))
	if err != nil {
		fmt.Printf("invalid log level `%s`, expected: debug | info | warn | error\n", level)
		logLevel = slog.LevelInfo
	}
	fmt.Printf("Log level is set to %s\n", logLevel)

	return logLevel
}

func convertTimeInUnixMicro(_ []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		a.Value = slog.Int64Value(a.Value.Time().UTC().UnixMicro())
	}
	return a
}
