package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

func TestParseCommand_Regular(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg      string
		expected domain.OpCode
	}{
		{
			name:     "SET command",
			arg:      "SET",
			expected: domain.OpSet,
		},
		{
			name:     "GET command",
			arg:      "GET",
			expected: domain.OpGet,
		},
		{
			name:     "DEL command",
			arg:      "DEL",
			expected: domain.OpDel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, err := domain.ParseOpCode(tt.arg)

			require.NoError(t, err)
			require.Equal(t, tt.expected, cmd)
		})
	}
}

func TestParseCommand_CaseInsensitive(t *testing.T) {
	t.Parallel()

	cmd, err := domain.ParseOpCode("gEt")

	require.NoError(t, err)
	require.Equal(t, domain.OpGet, cmd)
}

func TestParseCommand_Invalid(t *testing.T) {
	t.Parallel()

	cmd, err := domain.ParseOpCode("UNKNOWN")

	require.Error(t, err)
	require.Equal(t, domain.OpUnknown, cmd)
	require.Equal(t, `unsupported operation: "UNKNOWN"`, err.Error())
}

func TestParseCommand_Empty(t *testing.T) {
	t.Parallel()

	cmd, err := domain.ParseOpCode("abracadabra")

	require.Error(t, err)
	require.Equal(t, domain.OpUnknown, cmd)
	require.Equal(t, `unsupported operation: "abracadabra"`, err.Error())
}
