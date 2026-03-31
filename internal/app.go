package internal

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/niksmo/memdb/internal/config"
	"github.com/niksmo/memdb/internal/ports"
	"github.com/niksmo/memdb/pkg/logger"
)

type Runner interface {
	Run(ctx context.Context) error
}

type App struct {
	logger  *slog.Logger
	handler Runner
}

func NewApp(stdin io.Reader, stdout io.Writer, args []string) *App {
	app := &App{}
	cfg := config.FromFlags(args)

	app.initLogger(stdout, cfg.LogLevel)
	app.initHandler(stdin, stdout)

	return app
}

func (app *App) Run(ctx context.Context) {
	fmt.Println("🚀 bootstrapped successfully")

	go func() {
		_ = app.handler.Run(ctx)
	}()

	<-ctx.Done()

	fmt.Println()
	fmt.Println("✅closed successfully")
}

func (app *App) initLogger(w io.Writer, level string) {
	app.logger = logger.NewText(w, level)
}

func (app *App) initHandler(r io.Reader, w io.Writer) {
	app.handler = ports.NewStdinHandler(app.logger, r, w)
}
