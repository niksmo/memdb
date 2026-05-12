package memdb

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/niksmo/memdb/internal/memdb/config"
	"github.com/niksmo/memdb/internal/memdb/core"
	"github.com/niksmo/memdb/internal/memdb/core/compute"
	"github.com/niksmo/memdb/internal/memdb/core/compute/parser"
	"github.com/niksmo/memdb/internal/memdb/core/domain"
	"github.com/niksmo/memdb/internal/memdb/core/storage"
	"github.com/niksmo/memdb/internal/memdb/core/storage/engine"
	"github.com/niksmo/memdb/internal/memdb/core/storage/wal"
	"github.com/niksmo/memdb/pkg/closer"
	"github.com/niksmo/memdb/pkg/logger"
	"github.com/niksmo/memdb/pkg/transport"
)

type Closer interface {
	Add(fn func())
	CloseAll(ctx context.Context) error
}

type Handler interface {
	Handle(context.Context, []byte) ([]byte, error)
}

type Storage interface {
	Run() error
	Process(context.Context, domain.Operation) ([]byte, error)
	Load(context.Context) error
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

	logger  *slog.Logger
	storage Storage
	handler Handler
	srv     Listener
}

func NewApp(ctx context.Context, wr io.Writer, args []string) *App {
	app := &App{
		wr:        wr,
		appLogger: logger.New(wr, "info"),
		closer:    closer.New(),
	}

	app.initConfig(args)
	app.mustInitLogger()
	app.mustInitStorage(ctx)
	app.mustInitListener()

	return app
}

func (a *App) Run(ctx context.Context) {
	a.appLogger.Warn("memdb running")
	a.logger.Warn("got configuration", slog.Any("engine", a.cfg.Engine))
	a.logger.Warn("got configuration", slog.Any("network", a.cfg.Network))
	a.logger.Warn("got configuration", slog.Any("logging", a.cfg.Logging))
	a.logger.Warn("got configuration", slog.Any("wal", a.cfg.WAL))
	a.logger.Warn("bootstrapped successfully")

	storageDone := make(chan struct{})
	go func() {
		if err := a.storage.Run(); err != nil {
			a.logger.Error("storage run", slog.Any("error", err))
			close(storageDone)
		}
	}()

	listenerDone := make(chan struct{})
	go func() {
		defer close(listenerDone)
		if err := a.srv.Listen(ctx); err != nil {
			a.logger.Error("listen tcp", slog.Any("error", err))
		}
	}()

	select {
	case <-storageDone:
	case <-listenerDone:
	case <-ctx.Done():
		a.logger.Warn("memdb canceled", slog.Any("cause", context.Cause(ctx)))
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

	fw, err := logger.NewFileWriter(output)
	if err != nil {
		panic(err)
	}

	a.closer.Add(func() {
		if err := fw.Close(); err != nil {
			a.appLogger.Error("close log file", slog.Any("error", err))
		}
	})

	a.logger = logger.New(fw, a.cfg.Logging.Level)
}

func (a *App) mustInitStorage(ctx context.Context) {
	var (
		dir                  = a.cfg.WAL.DataDirectory
		maxSegmentSize       = a.cfg.WAL.MaxSegmentSize
		flushingBatchSize    = a.cfg.WAL.FlushingBatchSize
		flushingBatchTimeout = a.cfg.WAL.FlushingBatchTimeout
	)

	fileManager, err := wal.NewJSONFileManager(dir, maxSegmentSize)
	if err != nil {
		panic(fmt.Errorf("create wal file manager: %v", err))
	}

	writeAheadLog, err := wal.New(wal.Options{
		Logger:       a.logger,
		BatchSize:    flushingBatchSize,
		FlushingTick: flushingBatchTimeout,
		FileManager:  fileManager,
	})

	if err != nil {
		panic(fmt.Errorf("create wal: %v", err))
	}

	s := storage.New(storage.Options{
		Engine:     engine.New(),
		WAL:        writeAheadLog,
		WALEnabled: a.cfg.WAL.Enabled,
	})

	a.closer.Add(func() {
		if err := s.Close(); err != nil {
			a.logger.Error("storage close", slog.Any("error", err))
		}
	})

	if err := s.Load(ctx); err != nil {
		a.logger.Error("storage load data", slog.Any("error", err))
		return
	}

	a.storage = s
}

func (a *App) mustInitListener() {
	c := compute.New(parser.New())
	handler := core.NewPipeline(c, a.storage)

	srv, err := transport.NewListener(a.logger, handler, transport.Config{
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
			a.logger.Error("net listener close", slog.Any("error", err))
		}
	})
}

func (a *App) close() {
	closeCtx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if err := a.closer.CloseAll(closeCtx); err != nil {
		a.appLogger.Error("memdb close", slog.Any("error", err))
		return
	}
	a.appLogger.Warn("memdb closed successfully")
}
