package adapters

import (
	"context"
	"log/slog"
)

type Sender interface {
	Send(context.Context, []byte) ([]byte, error)
}

type Emitter struct {
	logger *slog.Logger
	sc     Sender
}

func NewEmitter(l *slog.Logger, sc Sender) *Emitter {
	return &Emitter{
		logger: l,
		sc:     sc,
	}
}

func (e *Emitter) Emit(ctx context.Context, p []byte) ([]byte, error) {
	data, err := e.sc.Send(ctx, p)
	if err != nil {
		return nil, err
	}
	return data, nil
}
