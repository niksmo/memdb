package internal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		var buf bytes.Buffer
		app := NewApp(&buf, &buf, []string{"log", "warn"})
		require.NotNil(t, app)
	})
}
