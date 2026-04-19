package main

import (
	"context"
	"log"

	"go-content-bot/pkg/app/bootstrap"
)

func main() {
	app, err := bootstrap.New()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("migrate shutdown error: %v", err)
		}
	}()

	if err := app.RunMigrations(context.Background()); err != nil {
		log.Fatal(err)
	}
}
