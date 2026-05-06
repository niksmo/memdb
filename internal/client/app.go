package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/niksmo/memdb/internal/client/adapters"
	"github.com/niksmo/memdb/internal/client/config"
	"github.com/niksmo/memdb/internal/client/ports"
	"github.com/niksmo/memdb/pkg/closer"
	"github.com/niksmo/memdb/pkg/logger"
	"github.com/niksmo/memdb/pkg/transport"
)

const connEstablishTimeout = 30 * time.Second

type Closer interface {
	Add(func())
	CloseAll(context.Context) error
}

type Runner interface {
	Run(ctx context.Context) error
}

type App struct {
	logger *slog.Logger
	tty    Runner
	r      io.Reader
	w      io.Writer
	closer Closer
}

func NewApp(ctx context.Context, r io.Reader, w io.Writer, args []string) *App {
	a := &App{
		r:      r,
		w:      w,
		closer: closer.New(),
	}

	cfg := config.FromFlags(args)
	a.logger = logger.New(w, cfg.LogLevel)

	a.mustInitTTY(ctx, cfg.ServerAddr)

	return a
}

func (a *App) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)

		if err := a.tty.Run(ctx); err != nil {
			a.logger.Error("runtime", slog.Any("error", err))
		}
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}

	closeCtx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if err := a.closer.CloseAll(closeCtx); err != nil {
		a.logger.Error("close client application", slog.Any("error", err))
		return
	}

	<-done

	a.logger.Info("memdb client closed successfully")
}

func (a *App) mustInitTTY(ctx context.Context, addr string) {
	ctx, cancel := context.WithTimeout(ctx, connEstablishTimeout)
	defer cancel()

	client, err := transport.NewClient(ctx, a.logger, addr)
	if err != nil {
		panic(fmt.Errorf("create transport client: %v", err))
	}

	a.closer.Add(func() {
		if err := client.Close(); err != nil {
			a.logger.Error("close transport client", slog.Any("error", err))
		}
	})

	emitter := adapters.NewEmitter(a.logger, client)

	a.tty = ports.NewTTY(a.logger, a.r, a.w, emitter)
}
