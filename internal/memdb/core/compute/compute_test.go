package compute_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/compute"
	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

type parserMock struct {
	parseFn func(payload []byte) (domain.Operation, error)

	called bool
}

func (m *parserMock) Parse(payload []byte) (domain.Operation, error) {
	m.called = true

	return m.parseFn(payload)
}

func TestCompute_Do_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock := &parserMock{}

	c := compute.New(mock)

	_, err := c.Do(ctx, []byte("GET key"))

	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, mock.called)
}

func TestCompute_Do_ParserError(t *testing.T) {
	t.Parallel()

	mock := &parserMock{
		parseFn: func(payload []byte) (domain.Operation, error) {
			return domain.Operation{}, assert.AnError
		},
	}

	c := compute.New(mock)

	_, err := c.Do(context.Background(), []byte("GET key"))

	require.Error(t, err)
	require.ErrorIs(t, err, assert.AnError)
	require.True(t, mock.called)
}

func TestCompute_Do_Success(t *testing.T) {
	expected := domain.Operation{
		Code:    domain.OpGet,
		Key:     "key",
		Payload: "value",
	}

	mock := &parserMock{
		parseFn: func(payload []byte) (domain.Operation, error) {
			return expected, nil
		},
	}

	c := compute.New(mock)

	payload := []byte("GET key")
	operation, err := c.Do(context.Background(), payload)

	require.NoError(t, err)
	require.Equal(t, expected, operation)

	require.True(t, mock.called)
}
