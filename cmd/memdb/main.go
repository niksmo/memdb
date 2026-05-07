package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/niksmo/memdb/internal/memdb"
)

func main() {
	logger := mainLogger()
	logger.Println("memdb executable path", os.Args[0], "PID", os.Getpid())

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	app := memdb.NewApp(ctx, os.Stdout, os.Args)
	app.Run(ctx)
}

func mainLogger() *log.Logger {
	return log.New(os.Stdout, "", 0)
}
