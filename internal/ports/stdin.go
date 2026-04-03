package ports

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

type Executer interface {
	Exec(ctx context.Context, stmt []byte) ([]byte, error)
}

type StdinHandler struct {
	log      *slog.Logger
	reader   io.Reader
	writer   io.Writer
	executer Executer
}

func NewStdinHandler(
	l *slog.Logger,
	r io.Reader,
	w io.Writer,
	e Executer,
) *StdinHandler {
	return &StdinHandler{log: l, reader: r, writer: w, executer: e}
}

func (h *StdinHandler) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(h.reader)

	h.writePrompt()

	for scanner.Scan() {
		stmt := scanner.Bytes()

		res, err := h.executer.Exec(ctx, stmt)
		if err != nil {
			h.writeError(err)
			continue
		}

		h.writeSuccess(res)
	}

	if err := scanner.Err(); err != nil {
		h.log.Error("failed to scan input", "error", err)
	}

	return nil
}

func (h *StdinHandler) validate(stmt []string) error {
	nArgs := len(stmt)
	if !(nArgs == 2 || nArgs == 3) {
		return errors.New("invalid statement")
	}
	return nil
}

func (h *StdinHandler) writePrompt() {
	_, _ = fmt.Fprintf(h.writer, "memdb ❯ ")
}

func (h *StdinHandler) writeError(err error) {
	_, _ = fmt.Fprintf(h.writer, "❌ %v\n", err)
	h.writePrompt()
}

func (h *StdinHandler) writeSuccess(res []byte) {
	_, _ = fmt.Fprintf(h.writer, "✅ %s\n", res)
	h.writePrompt()
}
