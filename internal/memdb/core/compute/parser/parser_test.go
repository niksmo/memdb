package parser_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/compute/parser"
	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

func TestParser_Parse_EmptyStatement(t *testing.T) {
	t.Parallel()

	p := parser.New()

	_, err := p.Parse([]byte(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "payload is empty")
}

func TestParser_Parse_InvalidArgsCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "only one argument",
			payload: []byte("GET"),
		},
		{
			name:    "more than three arguments",
			payload: []byte("GET myKey1 myValue1 myKey2 myValue2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := parser.New()

			_, err := p.Parse(tt.payload)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid number of arguments")

		})
	}

}

func TestParser_Parse_InvalidCommand(t *testing.T) {
	t.Parallel()

	p := parser.New()

	_, err := p.Parse([]byte("UNKNOWN key"))
	require.Error(t, err)
}

func TestParser_Parse_InvalidRequestFromModel(t *testing.T) {
	t.Parallel()

	p := parser.New()

	_, err := p.Parse([]byte("GET key value"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid arguments count")
}

func TestParser_Parse_SetCommand_Success(t *testing.T) {
	t.Parallel()

	p := parser.New()

	req, err := p.Parse([]byte("SET myKey myValue"))

	require.NoError(t, err)
	require.Equal(t, domain.OpSet, req.Code)
	require.Equal(t, "myKey", req.Key)
	require.Equal(t, "myValue", req.Payload)
}

func TestParser_Parse_GetCommand_Success(t *testing.T) {
	t.Parallel()

	p := parser.New()

	req, err := p.Parse([]byte("GET myKey"))

	require.NoError(t, err)

	require.Equal(t, domain.OpGet, req.Code)
	require.Equal(t, "myKey", req.Key)
	require.Equal(t, "", req.Payload)
}

func TestParser_Parse_DelCommand_Success(t *testing.T) {
	t.Parallel()

	p := parser.New()

	req, err := p.Parse([]byte("DEL myKey"))

	require.NoError(t, err)

	require.Equal(t, domain.OpDel, req.Code)
	require.Equal(t, "myKey", req.Key)
	require.Equal(t, "", req.Payload)
}
