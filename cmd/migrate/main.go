package main

import (
	"context"
	"log"

	"go-content-bot/pkg/app/bootstrap"
)

func main() {
	if err := bootstrap.RunMigrationsOnly(context.Background()); err != nil {
		log.Fatal(err)
	}
}
