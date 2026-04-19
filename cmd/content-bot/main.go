package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-content-bot/pkg/app/bootstrap"
	"go-content-bot/pkg/buildinfo"
	"go-content-bot/pkg/clihelp"
	"go-content-bot/pkg/config"
)

func main() {
	command := "run"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if isHelpCommand(command) {
		fmt.Print(renderHelp())
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch command {
	case "migrate":
		if err := bootstrap.RunMigrationsOnly(context.Background()); err != nil {
			log.Fatal(err)
		}
		return
	case "version":
		encoded, err := json.MarshalIndent(buildinfo.Current("content-bot.exe"), "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
		return
	case "show-runtime":
		details, err := config.LoadDetailsFromPaths()
		if err != nil {
			log.Fatal(err)
		}
		encoded, err := json.MarshalIndent(details.Config.RuntimeSummary(details.LoadedPaths), "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
		return
	case "check-connections":
		results, err := bootstrap.CheckConnectionsOnly(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		encoded, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
		return
	case "run":
		cfg, err := config.Load()
		if err != nil {
			log.Fatal(err)
		}
		if cfg.App.AutoMigrate {
			if err := bootstrap.RunMigrationsOnly(context.Background()); err != nil {
				log.Fatal(err)
			}
		}
	}

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
	default:
		log.Fatalf("unknown command: %s", command)
	}
}

func runUnified(ctx context.Context, app *bootstrap.App) error {
	if app == nil {
		return errors.New("app is required")
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

func isHelpCommand(command string) bool {
	switch command {
	case "help", "-h", "--help":
		return true
	default:
		return false
	}
}

func renderHelp() string {
	return clihelp.Render(clihelp.Document{
		Binary:         "content-bot.exe",
		Description:    "main runtime process",
		DefaultCommand: "run",
		Commands: []clihelp.Command{
			{Name: "run", Description: "run API and worker using config flags"},
			{Name: "api", Description: "run only the HTTP API"},
			{Name: "worker", Description: "run only the worker"},
			{Name: "migrate", Description: "apply SQL migrations"},
			{Name: "version", Description: "print build and version info"},
			{Name: "show-runtime", Description: "print effective runtime config"},
			{Name: "check-connections", Description: "check database and provider connectivity"},
			{Name: "help", Description: "show this help"},
		},
	})
}
