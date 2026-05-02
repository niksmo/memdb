package compute_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/compute"
	"github.com/niksmo/memdb/internal/memdb/core/models"
)

type parserMock struct {
	parseFn func(stmt []byte) (models.Request, error)

	called bool
}

func (m *parserMock) Parse(stmt []byte) (models.Request, error) {
	m.called = true

	return m.parseFn(stmt)
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
		parseFn: func(stmt []byte) (models.Request, error) {
			return models.Request{}, assert.AnError
		},
	}

	c := compute.New(mock)

	_, err := c.Do(context.Background(), []byte("GET key"))

	require.Error(t, err)
	require.ErrorIs(t, err, assert.AnError)
	require.True(t, mock.called)
}

func TestCompute_Do_Success(t *testing.T) {
	expected := models.Request{
		Cmd:   models.CommandGet,
		Key:   "key",
		Value: "value",
	}

	mock := &parserMock{
		parseFn: func(stmt []byte) (models.Request, error) {
			return expected, nil
		},
	}

	c := compute.New(mock)

	stmt := []byte("GET key")
	req, err := c.Do(context.Background(), stmt)

	require.NoError(t, err)
	require.Equal(t, expected, req)

	require.True(t, mock.called)
}
