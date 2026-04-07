package storage

import (
	"context"
	"testing"

	"github.com/niksmo/memdb/internal/core/models"
	"github.com/niksmo/memdb/internal/core/storage/engine"
	"github.com/stretchr/testify/require"
)

type mockEngine struct {
	setFn func(key, val string)
	getFn func(key string) (string, error)
	delFn func(key string)
}

func (m *mockEngine) Set(key, val string) { m.setFn(key, val) }

func (m *mockEngine) Get(key string) (string, error) { return m.getFn(key) }

func (m *mockEngine) Del(key string) { m.delFn(key) }

func TestStorage_Process_Set(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		setFn: func(key, val string) {},
	}
	s := New(e)
	req := models.Request{
		Cmd:   models.CommandSet,
		Key:   "k1",
		Value: "v1",
	}

	resp, err := s.Process(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), resp)
}

func TestStorage_Process_Get(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		getFn: func(key string) (string, error) {
			if key == "k1" {
				return "v1", nil
			}
			return "", nil
		},
	}

	s := New(e)
	req := models.Request{
		Cmd:   models.CommandGet,
		Key:   "k1",
		Value: "",
	}

	resp, err := s.Process(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), resp)
}

func TestStorage_Process_Get_NotFound(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		getFn: func(key string) (string, error) {
			return "", engine.ErrKeyNotFound
		},
	}

	s := New(e)

	req := models.Request{
		Cmd: models.CommandGet,
		Key: "missing",
	}

	_, err := s.Process(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestStorage_Process_Del(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		delFn: func(key string) {},
	}

	s := New(e)

	req := models.Request{
		Cmd: models.CommandDel,
		Key: "k1",
	}

	resp, err := s.Process(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), resp)
}

func TestStorage_Process_Del_NotFound(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		delFn: func(key string) {},
	}

	s := New(e)

	req := models.Request{
		Cmd: models.CommandDel,
		Key: "missing",
	}
	_, err := s.Process(context.Background(), req)
	require.NoError(t, err)
}

func TestStorage_Process_UnknownCommand(t *testing.T) {
	t.Parallel()

	e := &mockEngine{}

	s := New(e)

	req := models.Request{}

	_, err := s.Process(context.Background(), req)
	require.Error(t, err)
}

func TestStorage_Process_CtxCanceled(t *testing.T) {
	t.Parallel()

	e := &mockEngine{}

	s := New(e)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := models.Request{}

	_, err := s.Process(ctx, req)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}
