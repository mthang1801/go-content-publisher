package bootstrap

import (
	"context"
	"fmt"

	"go-content-bot/pkg/config"
	postgresdb "go-content-bot/pkg/database/postgres"
	"go-content-bot/pkg/migration"
)

func RunMigrationsOnly(ctx context.Context, paths ...string) error {
	cfg, err := config.Load(paths...)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	_, sqlDB, err := postgresdb.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer sqlDB.Close()

	if err := migration.New("db/migrations").Up(ctx, sqlDB); err != nil {
		return err
	}
	return nil
}
