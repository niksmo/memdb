package core

import (
	"context"
	"errors"
	"testing"

	"github.com/niksmo/memdb/internal/core/models"
	"github.com/niksmo/memdb/internal/core/storage"
	"github.com/stretchr/testify/require"
)

type mockCompute struct {
	doFn func(ctx context.Context, stmt []byte) (models.Request, error)
}

func (m *mockCompute) Do(ctx context.Context, stmt []byte) (models.Request, error) {
	return m.doFn(ctx, stmt)
}

type mockStorage struct {
	processFn func(ctx context.Context, req models.Request) ([]byte, error)
}

func (m *mockStorage) Process(ctx context.Context, req models.Request) ([]byte, error) {
	return m.processFn(ctx, req)
}

func TestPipeline_Exec_Success(t *testing.T) {
	t.Parallel()

	req, _ := models.NewRequest(models.CommandSet, "k1", "v1")

	c := &mockCompute{
		doFn: func(ctx context.Context, stmt []byte) (models.Request, error) {
			return req, nil
		},
	}

	s := &mockStorage{
		processFn: func(ctx context.Context, r models.Request) ([]byte, error) {
			return []byte("OK"), nil
		},
	}

	p := NewPipeline(c, s)
	resp, err := p.Exec(context.Background(), []byte("SET k1 v1"))
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), resp)
}

func TestPipeline_Exec_ComputeError(t *testing.T) {
	t.Parallel()

	computeErr := errors.New("compute failed")

	c := &mockCompute{
		doFn: func(ctx context.Context, stmt []byte) (models.Request, error) {
			return models.Request{}, computeErr
		},
	}
	s := &mockStorage{}

	p := NewPipeline(c, s)
	_, err := p.Exec(context.Background(), []byte("any"))
	require.ErrorIs(t, err, computeErr)
}

func TestPipeline_Exec_StorageError(t *testing.T) {
	t.Parallel()

	req, _ := models.NewRequest(models.CommandSet, "k1", "v1")

	storageErr := errors.New("storage failed")

	c := &mockCompute{
		doFn: func(ctx context.Context, stmt []byte) (models.Request, error) {
			return req, nil
		},
	}
	s := &mockStorage{
		processFn: func(ctx context.Context, r models.Request) ([]byte, error) {
			return nil, storageErr
		},
	}

	p := NewPipeline(c, s)
	_, err := p.Exec(context.Background(), []byte("SET k1 v1"))
	require.ErrorIs(t, err, storageErr)
}

func TestPipeline_Exec_StorageUnknownCommand(t *testing.T) {
	req := models.Request{}

	c := &mockCompute{
		doFn: func(ctx context.Context, stmt []byte) (models.Request, error) {
			return req, nil
		},
	}
	s := &mockStorage{
		processFn: func(ctx context.Context, r models.Request) ([]byte, error) {
			return nil, storage.ErrUnknownCommand
		},
	}

	p := NewPipeline(c, s)
	_, err := p.Exec(context.Background(), []byte("FOO k1"))
	require.ErrorIs(t, err, ErrInternal)
}
