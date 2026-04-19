package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"go-content-bot/pkg/app/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("worker shutdown error: %v", err)
		}
	}()

	if err := app.RunWorker(ctx); err != nil {
		log.Fatal(err)
	}
}
