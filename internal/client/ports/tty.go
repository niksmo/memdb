package ports

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

type Emitter interface {
	Emit(ctx context.Context, stmt []byte) ([]byte, error)
}

type TTY struct {
	logger  *slog.Logger
	r       io.Reader
	w       io.Writer
	emitter Emitter
}

func NewTTY(
	l *slog.Logger,
	r io.Reader,
	w io.Writer,
	e Emitter,
) *TTY {
	return &TTY{
		logger:  l,
		r:       r,
		w:       w,
		emitter: e,
	}
}

func (t *TTY) Run(ctx context.Context) error {
	inputCh := t.listenInput()

	for {
		t.writePrompt()

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case buf, ok := <-inputCh:
			if !ok {
				return nil
			}

			if err := t.process(ctx, buf); err != nil {
				return err
			}
		}
	}
}

func (t *TTY) writePrompt() {
	_, _ = fmt.Fprintf(t.w, "memdb ❯ ")
}

func (t *TTY) writeSuccess(res []byte) {
	_, _ = fmt.Fprintf(t.w, "%s\n", res)
}

func (t *TTY) listenInput() <-chan []byte {
	ch := make(chan []byte)

	go func() {
		defer close(ch)

		reader := bufio.NewReader(t.r)

		for {
			buf, err := reader.ReadBytes('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				t.logger.Error("input", slog.Any("error", err))
				continue
			}

			ch <- bytes.TrimSpace(buf)
		}
	}()

	return ch
}

func (t *TTY) process(ctx context.Context, buf []byte) error {
	res, err := t.emitter.Emit(ctx, buf)
	if errors.Is(err, context.Canceled) {
		return err
	}

	if err != nil {
		t.logger.Error("query execution", slog.Any("error", err))
		return nil
	}

	t.writeSuccess(res)

	return nil
}
