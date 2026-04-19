package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-content-bot/pkg/app/bootstrap"
)

func main() {
	command := "run"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("content-bot shutdown error: %v", err)
		}
	}()

	switch command {
	case "run":
		if err := runUnified(ctx, app); err != nil {
			log.Fatal(err)
		}
	case "api":
		if err := app.RunAPI(ctx); err != nil {
			log.Fatal(err)
		}
	case "worker":
		if err := app.RunWorker(ctx); err != nil {
			log.Fatal(err)
		}
	case "migrate":
		if err := app.RunMigrations(context.Background()); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command: %s", command)
	}
}

func runUnified(ctx context.Context, app *bootstrap.App) error {
	if app == nil {
		return errors.New("app is required")
	}
	if app.Config.App.AutoMigrate {
		if err := app.RunMigrations(context.Background()); err != nil {
			return err
		}
	}

	runAPI := app.Config.App.RunAPI
	runWorker := app.Config.App.RunWorker
	if !runAPI && !runWorker {
		return errors.New("APP_RUN_API and APP_RUN_WORKER cannot both be false")
	}

	errCh := make(chan error, 2)
	running := 0
	if runAPI {
		running++
		go func() {
			errCh <- app.RunAPI(ctx)
		}()
	}
	if runWorker {
		running++
		go func() {
			errCh <- app.RunWorker(ctx)
		}()
	}

	for i := 0; i < running; i++ {
		err := <-errCh
		if err == nil || errors.Is(err, context.Canceled) {
			continue
		}
		return err
	}
	return nil
}
