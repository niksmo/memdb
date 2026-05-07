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
		ctx := t.Context()
		wr := new(bytes.Buffer)
		app := NewApp(ctx, wr, os.Args[:1])
		require.NotNil(t, app)
	})
}
