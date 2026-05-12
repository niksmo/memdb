package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

type mockCompute struct {
	doFn func(ctx context.Context, payload []byte) (domain.Operation, error)
}

func (m *mockCompute) Do(ctx context.Context, payload []byte) (domain.Operation, error) {
	return m.doFn(ctx, payload)
}

type mockStorage struct {
	processFn func(ctx context.Context, op domain.Operation) ([]byte, error)
}

func (m *mockStorage) Process(ctx context.Context, operation domain.Operation) ([]byte, error) {
	return m.processFn(ctx, operation)
}

func TestPipeline_Exec_Success(t *testing.T) {
	t.Parallel()

	operation := domain.Operation{
		Code:    domain.OpSet,
		Key:     "k1",
		Payload: "v1",
	}

	c := &mockCompute{
		doFn: func(ctx context.Context, payload []byte) (domain.Operation, error) {
			return operation, nil
		},
	}

	s := &mockStorage{
		processFn: func(ctx context.Context, r domain.Operation) ([]byte, error) {
			return []byte("OK"), nil
		},
	}

	p := NewPipeline(c, s)
	data, err := p.Handle(context.Background(), []byte("SET k1 v1"))
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), data)
}

func TestPipeline_Exec_ComputeError(t *testing.T) {
	t.Parallel()

	computeErr := errors.New("compute failed")

	c := &mockCompute{
		doFn: func(ctx context.Context, payload []byte) (domain.Operation, error) {
			return domain.Operation{}, computeErr
		},
	}
	s := &mockStorage{}

	p := NewPipeline(c, s)
	_, err := p.Handle(context.Background(), []byte("any"))
	require.ErrorIs(t, err, computeErr)
}

func TestPipeline_Exec_StorageError(t *testing.T) {
	t.Parallel()

	operation := domain.Operation{
		Code:    domain.OpSet,
		Key:     "k1",
		Payload: "v1",
	}

	storageErr := errors.New("storage failed")

	c := &mockCompute{
		doFn: func(ctx context.Context, payload []byte) (domain.Operation, error) {
			return operation, nil
		},
	}
	s := &mockStorage{
		processFn: func(ctx context.Context, r domain.Operation) ([]byte, error) {
			return nil, storageErr
		},
	}

	p := NewPipeline(c, s)
	_, err := p.Handle(context.Background(), []byte("SET k1 v1"))
	require.ErrorIs(t, err, storageErr)
}

func TestPipeline_Exec_StorageUnknownCommand(t *testing.T) {
	operation := domain.Operation{}

	c := &mockCompute{
		doFn: func(ctx context.Context, payload []byte) (domain.Operation, error) {
			return operation, nil
		},
	}
	s := &mockStorage{
		processFn: func(ctx context.Context, r domain.Operation) ([]byte, error) {
			return nil, assert.AnError
		},
	}

	p := NewPipeline(c, s)
	_, err := p.Handle(context.Background(), []byte("FOO k1"))
	require.ErrorIs(t, err, assert.AnError)
}
