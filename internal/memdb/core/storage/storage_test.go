package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
	"github.com/niksmo/memdb/internal/memdb/core/storage/engine"
)

type mockEngine struct {
	setFn func(key, payload string)
	getFn func(key string) (string, error)
	delFn func(key string)
}

func (m *mockEngine) Set(key, payload string) { m.setFn(key, payload) }

func (m *mockEngine) Get(key string) (string, error) { return m.getFn(key) }

func (m *mockEngine) Del(key string) { m.delFn(key) }

func TestStorage_Process_Set(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		setFn: func(key, val string) {},
	}
	s := New(Options{Engine: e})
	op := domain.Operation{
		Code:    domain.OpSet,
		Key:     "k1",
		Payload: "v1",
	}

	resp, err := s.Process(context.Background(), op)
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

	s := New(Options{Engine: e})
	op := domain.Operation{
		Code:    domain.OpGet,
		Key:     "k1",
		Payload: "",
	}

	data, err := s.Process(context.Background(), op)
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), data)
}

func TestStorage_Process_Get_NotFound(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		getFn: func(key string) (string, error) {
			return "", engine.NotFound
		},
	}

	s := New(Options{Engine: e})

	op := domain.Operation{
		Code: domain.OpGet,
		Key:  "missing",
	}

	_, err := s.Process(context.Background(), op)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestStorage_Process_Del(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		delFn: func(key string) {},
	}

	s := New(Options{Engine: e})

	op := domain.Operation{
		Code: domain.OpDel,
		Key:  "k1",
	}

	data, err := s.Process(context.Background(), op)
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), data)
}

func TestStorage_Process_Del_NotFound(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		delFn: func(key string) {},
	}

	s := New(Options{Engine: e})

	op := domain.Operation{
		Code: domain.OpDel,
		Key:  "missing",
	}
	_, err := s.Process(context.Background(), op)
	require.NoError(t, err)
}

func TestStorage_Process_UnknownCommand(t *testing.T) {
	t.Parallel()

	e := &mockEngine{}

	s := New(Options{Engine: e})

	op := domain.Operation{}

	_, err := s.Process(context.Background(), op)
	require.Error(t, err)
}

func TestStorage_Process_CtxCanceled(t *testing.T) {
	t.Parallel()

	e := &mockEngine{
		setFn: func(key, val string) {},
	}

	s := New(Options{Engine: e})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	op := domain.Operation{Code: domain.OpSet}

	_, err := s.Process(ctx, op)
	require.NoError(t, err) // context errors are ignored by the storage implementation.
}
