package models_test

import (
	"testing"

	"github.com/niksmo/memdb/internal/core/models"
	"github.com/stretchr/testify/require"
)

func TestParseCommand_Regular(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg      string
		expected models.Command
	}{
		{
			name:     "SET command",
			arg:      "SET",
			expected: models.CommandSet,
		},
		{
			name:     "GET command",
			arg:      "GET",
			expected: models.CommandGet,
		},
		{
			name:     "DEL command",
			arg:      "DEL",
			expected: models.CommandDel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, err := models.ParseCommand(tt.arg)

			require.NoError(t, err)
			require.Equal(t, tt.expected, cmd)
		})
	}
}

func TestParseCommand_CaseInsensitive(t *testing.T) {
	t.Parallel()

	cmd, err := models.ParseCommand("gEt")

	require.NoError(t, err)
	require.Equal(t, models.CommandGet, cmd)
}

func TestParseCommand_Invalid(t *testing.T) {
	t.Parallel()

	cmd, err := models.ParseCommand("UNKNOWN")

	require.Error(t, err)
	require.Equal(t, models.CommandUnknown, cmd)
	require.Equal(t, `invalid command "UNKNOWN"`, err.Error())
}

func TestParseCommand_Empty(t *testing.T) {
	t.Parallel()

	cmd, err := models.ParseCommand("abracadabra")

	require.Error(t, err)
	require.Equal(t, models.CommandUnknown, cmd)
	require.Equal(t, `invalid command "abracadabra"`, err.Error())
}
