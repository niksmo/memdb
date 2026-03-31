package internal

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/niksmo/memdb/internal/config"
	"github.com/niksmo/memdb/pkg/logger"
)

type App struct {
	logger *slog.Logger
}

func NewApp(stdout io.Writer, args []string) *App {
	app := &App{}
	cfg := config.FromFlags(args)

	app.initLogger(stdout, cfg.LogLevel)

	return &App{}
}

func (app *App) Run(ctx context.Context) {
	fmt.Println("memdb bootstrapped successfully 🚀")
	<-ctx.Done()
	fmt.Println("memdb closed successfully ✅")
}

func (app *App) initLogger(w io.Writer, level string) {
	app.logger = logger.NewText(w, level)
}
