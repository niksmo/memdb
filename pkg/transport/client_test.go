package transport

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TODO
func TestClient(t *testing.T) {
	t.Skip()
	t.Parallel()

	options := []ClientOption{
		WithMaxResponseSize(1024),
		WithReadTimeout(30 * time.Second),
		WithWriteTimeout(30 * time.Second),
	}
	cli, err := NewClient(t.Context(), slog.Default(), "127.0.0.1:8000", options...)
	require.NoError(t, err)
	require.NotNil(t, cli)
}
