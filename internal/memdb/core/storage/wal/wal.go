package wal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

type FileManager interface {
	WriteOps([]domain.Operation) error
	ReadOps(context.Context, func(domain.Operation) error) error
	Close() error
}

var errClosed = errors.New("closed")

type Implementation struct {
	logger       *slog.Logger
	batchSize    int
	flushingTick time.Duration
	fm           FileManager

	batch  []domain.Operation
	waitQ  []chan<- struct{}
	ticker *time.Ticker

	mu      sync.RWMutex
	eventCh chan event
	closed  bool

	closeDone chan struct{}
}

func New(opt Options) (*Implementation, error) {
	if err := opt.validate(); err != nil {
		return nil, err
	}

	imp := &Implementation{
		logger:       opt.Logger.With(slog.String("component", "wal")),
		batchSize:    opt.BatchSize,
		flushingTick: opt.FlushingTick,
		fm:           opt.FileManager,

		eventCh: make(chan event),
		batch:   make([]domain.Operation, 0, opt.BatchSize),
		waitQ:   make([]chan<- struct{}, 0, opt.BatchSize),
		ticker:  time.NewTicker(opt.FlushingTick),

		closeDone: make(chan struct{}),
	}

	return imp, nil
}

func (imp *Implementation) LoadAll(ctx context.Context, fn func(opCode domain.OpCode, key, payload string)) error {
	const op = "WAL.LoadAll"

	err := imp.fm.ReadOps(ctx, func(operation domain.Operation) error {
		fn(operation.Code, operation.Key, operation.Payload)
		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (imp *Implementation) WriteLog(ctx context.Context, opCode domain.OpCode, key, payload string) error {
	const op = "WAL.WriteLog"

	imp.mu.RLock()
	if imp.closed {
		imp.mu.RUnlock()
		return fmt.Errorf("%s: failed to write log: wal is closed", op)
	}

	evt, done := newEvent(opCode, key, payload)

	select {
	case <-ctx.Done(): // check on done below for exclude mu.RUnlock() duplication in both cases
	case imp.eventCh <- evt:
	}
	imp.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return context.Cause(ctx)
	}

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-done:
		return nil
	}
}

func (imp *Implementation) Run() error {
	const op = "WAL.Run"

	defer close(imp.closeDone)

	defer imp.ticker.Stop()

	for {
		if err := imp.handleEvent(); err != nil {
			if errors.Is(err, errClosed) {
				return nil
			}
			return fmt.Errorf("%s: %w", op, err)
		}
	}
}

func (imp *Implementation) Close() error {
	const op = "WAL.Close"

	defer func() {
		if err := imp.fm.Close(); err != nil {
			imp.logger.Error("file manager close", slog.Any("error", err))
		}
	}()

	imp.mu.Lock()
	if imp.closed {
		imp.mu.Unlock()
		return fmt.Errorf("%s: wal is already closed", op)
	}

	close(imp.eventCh)
	imp.closed = true
	imp.mu.Unlock()

	<-imp.closeDone

	return nil
}

func (imp *Implementation) handleEvent() error {
	select {
	case evt, ok := <-imp.eventCh:
		if !ok { // flush on close
			if err := imp.flush(); err != nil {
				return err
			}
			return errClosed
		}

		imp.batch = append(imp.batch, evt.operation)
		imp.waitQ = append(imp.waitQ, evt.done)

		if len(imp.batch) != imp.batchSize {
			return nil
		}

		if err := imp.flush(); err != nil {
			return err
		}

		imp.ticker.Reset(imp.flushingTick)
		return nil

	case <-imp.ticker.C:
		return imp.flush()
	}
}

func (imp *Implementation) flush() error {
	if len(imp.batch) == 0 {
		return nil
	}

	if err := imp.fm.WriteOps(imp.batch); err != nil {
		return err
	}

	for _, done := range imp.waitQ {
		close(done)
	}

	imp.batch = imp.batch[:0]
	imp.waitQ = imp.waitQ[:0]

	return nil
}
