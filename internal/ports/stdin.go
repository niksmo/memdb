package ports

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
)

type StdinHandler struct {
	log *slog.Logger
	r   io.Reader
	w   io.Writer
}

func NewStdinHandler(l *slog.Logger, r io.Reader, w io.Writer) *StdinHandler {
	return &StdinHandler{log: l, r: r, w: w}
}

func (h *StdinHandler) Run(_ context.Context) error {
	scanner := bufio.NewScanner(h.r)

	h.writePrompt()
	for scanner.Scan() {
		input := scanner.Text()

		fmt.Println("🆗 ❮", input)

		h.writePrompt()
	}

	if err := scanner.Err(); err != nil {
		h.log.Error("invalid input", "error", err)
	}
	return nil
}

func (h *StdinHandler) writePrompt() {
	_, _ = fmt.Fprintf(h.w, "memdb ❯ ")
}
