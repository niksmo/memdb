package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/niksmo/memdb/internal/client"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	app := client.NewApp(ctx, os.Stdin, os.Stdout, os.Args)
	app.Run(ctx)
}
