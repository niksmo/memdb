package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewText(t *testing.T) {
	t.Parallel()

	t.Run("valid level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		log := New(&buf, "warn")

		log.Warn("test message", "key", "value")

		output := buf.String()
		require.Contains(t, output, "test message")
	})

	t.Run("invalid level should backoff on info level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		log := New(&buf, "invalid level")
		output, _ := buf.ReadString('\n')
		require.Contains(t, output, "invalid log level")

		log.Debug("debug message", "key", "value")
		output, _ = buf.ReadString('\n')
		require.Zero(t, output)

		log.Info("info message", "key", "value")
		output, _ = buf.ReadString('\n')
		require.Contains(t, output, "info message")
	})
}
