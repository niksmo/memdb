package memdb

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		var buf bytes.Buffer
		app := NewApp(&buf, os.Args[:1])
		require.NotNil(t, app)
	})
}
