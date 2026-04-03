package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEngine_SetAndGet(t *testing.T) {
	t.Parallel()

	e := New()

	e.Set("key1", "value1")

	val, err := e.Get("key1")
	require.NoError(t, err)
	require.Equal(t, "value1", val)
}

func TestEngine_Get_NotFound(t *testing.T) {
	t.Parallel()

	e := New()

	val, err := e.Get("missing")
	require.ErrorIs(t, err, ErrKeyNotFound)
	require.Empty(t, val)
}

func TestEngine_Set_Overwrite(t *testing.T) {
	t.Parallel()

	e := New()

	e.Set("key1", "value1")
	e.Set("key1", "value2")

	val, err := e.Get("key1")
	require.NoError(t, err)
	require.Equal(t, "value2", val)
}

func TestEngine_Del_Success(t *testing.T) {
	t.Parallel()

	e := New()

	e.Set("key1", "value1")

	err := e.Del("key1")
	require.NoError(t, err)

	_, err = e.Get("key1")
	require.ErrorIs(t, err, ErrKeyNotFound)
}

func TestEngine_Del_NotFound(t *testing.T) {
	t.Parallel()

	e := New()

	err := e.Del("missing")

	require.ErrorIs(t, err, ErrKeyNotFound)
}

func TestEngine_StateIsolation(t *testing.T) {
	t.Parallel()

	e := New()

	e.Set("k1", "v1")
	e.Set("k2", "v2")

	err := e.Del("k1")
	require.NoError(t, err)

	val, err := e.Get("k2")
	require.NoError(t, err)
	require.Equal(t, "v2", val)
}
