package models_test

import (
	"testing"

	"github.com/niksmo/memdb/internal/core/models"
	"github.com/stretchr/testify/require"
)

func TestNewRequest_Set_Success(t *testing.T) {
	t.Parallel()

	req, err := models.NewRequest(models.CommandSet, "key", "value")

	require.NoError(t, err)

	require.Equal(t, models.CommandSet, req.Cmd())
	require.Equal(t, "key", req.Key())
	require.Equal(t, "value", req.Value())
}

func TestNewRequest_Get_Success(t *testing.T) {
	t.Parallel()

	req, err := models.NewRequest(models.CommandGet, "key", "")

	require.NoError(t, err)

	require.Equal(t, models.CommandGet, req.Cmd())
	require.Equal(t, "key", req.Key())
	require.Equal(t, "", req.Value())
}

func TestNewRequest_Del_Success(t *testing.T) {
	t.Parallel()

	req, err := models.NewRequest(models.CommandDel, "key", "")

	require.NoError(t, err)

	require.Equal(t, models.CommandDel, req.Cmd())
	require.Equal(t, "key", req.Key())
	require.Equal(t, "", req.Value())
}

func TestNewRequest_EmptyKey(t *testing.T) {
	t.Parallel()

	_, err := models.NewRequest(models.CommandSet, "", "value")

	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrEmptyKey)
}

func TestNewRequest_Set_EmptyValue(t *testing.T) {
	t.Parallel()

	_, err := models.NewRequest(models.CommandSet, "key", "")

	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrEmptyValue)
	require.Contains(t, err.Error(), "value is empty")
}

func TestNewRequest_Get_WithValue(t *testing.T) {
	t.Parallel()

	_, err := models.NewRequest(models.CommandGet, "key", "value")

	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrMustEmptyValue)
	require.Contains(t, err.Error(), "value must be empty")
}

func TestNewRequest_Del_WithValue(t *testing.T) {
	t.Parallel()

	_, err := models.NewRequest(models.CommandDel, "key", "value")

	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrMustEmptyValue)
	require.Contains(t, err.Error(), "value must be empty")
}

func TestNewRequest_UnknownCommand(t *testing.T) {
	t.Parallel()

	_, err := models.NewRequest(models.CommandUnknown, "key", "")

	require.Error(t, err)
	require.ErrorIs(t, err, models.ErrInvalidCommand)
}
