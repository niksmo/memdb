package parser_test

import (
	"context"
	"testing"

	"github.com/niksmo/memdb/internal/core/compute/parser"
	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/core/models"
)

func TestParser_Parse_EmptyStatement(t *testing.T) {
	t.Parallel()

	p := parser.New()

	_, err := p.Parse(context.Background(), []byte(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "statement is empty")
}

func TestParser_Parse_InvalidUTF8(t *testing.T) {
	t.Parallel()

	p := parser.New()

	// invalid UTF-8 byte sequence
	stmt := []byte{0xff, 0xfe, 0xfd}

	_, err := p.Parse(context.Background(), stmt)
	require.Error(t, err)
	require.Contains(t, err.Error(), "statement should be the valid encoded string")
}

func TestParser_Parse_InvalidArgsCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		stmt []byte
	}{
		{
			name: "only one argument",
			stmt: []byte("GET"),
		},
		{
			name: "more then three arguments",
			stmt: []byte("GET myKey1 myValue1 myKey2 myValue2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := parser.New()

			_, err := p.Parse(context.Background(), tt.stmt)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid number of statement arguments")

		})
	}

}

func TestParser_Parse_InvalidCommand(t *testing.T) {
	t.Parallel()

	p := parser.New()

	_, err := p.Parse(context.Background(), []byte("UNKNOWN key"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid statement command")
}

func TestParser_Parse_InvalidRequestFromModel(t *testing.T) {
	t.Parallel()

	p := parser.New()

	// assuming models.NewRequest will fail for invalid combination
	// adjust depending on your domain rules
	_, err := p.Parse(context.Background(), []byte("GET key value"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid statement:")
}

func TestParser_Parse_SetCommand_Success(t *testing.T) {
	t.Parallel()

	p := parser.New()

	req, err := p.Parse(context.Background(), []byte("SET myKey myValue"))

	require.NoError(t, err)
	require.Equal(t, models.CommandSet, req.Cmd())
	require.Equal(t, "myKey", req.Key())
	require.Equal(t, "myValue", req.Value())
}

func TestParser_Parse_GetCommand_Success(t *testing.T) {
	t.Parallel()

	p := parser.New()

	req, err := p.Parse(context.Background(), []byte("GET myKey"))

	require.NoError(t, err)

	require.Equal(t, models.CommandGet, req.Cmd())
	require.Equal(t, "myKey", req.Key())
	require.Equal(t, "", req.Value())
}

func TestParser_Parse_DelCommand_Success(t *testing.T) {
	t.Parallel()

	p := parser.New()

	req, err := p.Parse(context.Background(), []byte("DEL myKey"))

	require.NoError(t, err)

	require.Equal(t, models.CommandDel, req.Cmd())
	require.Equal(t, "myKey", req.Key())
	require.Equal(t, "", req.Value())
}
