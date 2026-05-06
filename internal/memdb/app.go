package memdb

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/niksmo/memdb/internal/memdb/config"
	"github.com/niksmo/memdb/internal/memdb/core"
	"github.com/niksmo/memdb/internal/memdb/core/compute"
	"github.com/niksmo/memdb/internal/memdb/core/compute/parser"
	"github.com/niksmo/memdb/internal/memdb/core/storage"
	"github.com/niksmo/memdb/internal/memdb/core/storage/engine"
	"github.com/niksmo/memdb/pkg/closer"
	"github.com/niksmo/memdb/pkg/logger"
	"github.com/niksmo/memdb/pkg/transport"
)

type Closer interface {
	Add(fn func())
	CloseAll(ctx context.Context) error
}

type Listener interface {
	Listen(context.Context) error
	Close() error
}

type App struct {
	cfg       config.Config
	appLogger *slog.Logger
	wr        io.Writer
	closer    Closer

	logger *slog.Logger
	srv    Listener
}

func NewApp(wr io.Writer, args []string) *App {
	app := &App{
		wr:        wr,
		appLogger: logger.New(wr, "info"),
		closer:    closer.New(),
	}

	app.initConfig(args)
	app.mustInitLogger()
	app.mustInitListener()

	return app
}

func (a *App) Run(ctx context.Context) {
	a.appLogger.Warn("memdb running")
	a.logger.Warn("got configuration", slog.Any("engine", a.cfg.Engine))
	a.logger.Warn("got configuration", slog.Any("network", a.cfg.Network))
	a.logger.Warn("got configuration", slog.Any("logging", a.cfg.Logging))
	a.logger.Warn("bootstrapped successfully")

	if err := a.srv.Listen(ctx); err != nil {
		a.logger.Error("listen tcp", slog.Any("error", err))
	}

	a.close()
}

func (a *App) initConfig(args []string) {
	a.cfg = config.Load(args)
	fmt.Fprintf(a.wr, "\nmemdb configuration\n%s\n", a.cfg)
}

func (a *App) mustInitLogger() {
	output := a.cfg.Logging.Output
	if output == "" {
		a.logger = logger.New(a.wr, a.cfg.Logging.Level)
		return
	}
	dir := filepath.Dir(output)
	if !filepath.IsAbs(dir) {
		panic("log output destination must be defined using absolute path")
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		panic(fmt.Errorf("create log output path: %v", err))
	}

	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Errorf("open log output path: %v", err))
	}
	a.logger = logger.New(file, a.cfg.Logging.Level)

	a.closer.Add(func() {
		if err := file.Close(); err != nil {
			a.appLogger.Error("close log file", slog.Any("error", err))
		}
	})
}

func (a *App) mustInitListener() {
	executer := core.NewPipeline(
		compute.New(parser.New()),
		storage.New(engine.New()),
	)

	srv, err := transport.NewListener(a.logger, executer, transport.Config{
		Address:        a.cfg.Network.Address,
		MaxMessageSize: a.cfg.Network.MaxMessageSize,
		MaxConnections: a.cfg.Network.MaxConnections,
		IdleTimeout:    a.cfg.Network.IdleTimeout,
	})
	if err != nil {
		panic(fmt.Errorf("create listener: %v", err))
	}

	a.srv = srv

	a.closer.Add(func() {
		if err := srv.Close(); err != nil {
			a.logger.Error("close memdb", slog.Any("error", err))
		}
	})
}

func (a *App) close() {
	closeCtx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if err := a.closer.CloseAll(closeCtx); err != nil {
		a.appLogger.Error("close application", slog.Any("error", err))
		return
	}
	a.appLogger.Warn("memdb closed successfully")
}
