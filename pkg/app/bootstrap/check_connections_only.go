package bootstrap

import (
	"context"
	"fmt"

	systemapp "go-content-bot/internal/system/application"
	deepseekclient "go-content-bot/internal/system/infrastructure/clients/deepseek"
	geminiclient "go-content-bot/internal/system/infrastructure/clients/gemini"
	telegramclient "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	"go-content-bot/pkg/config"
	postgresdb "go-content-bot/pkg/database/postgres"
)

func CheckConnectionsOnly(ctx context.Context, paths ...string) (systemapp.ConnectivityResults, error) {
	cfg, err := config.Load(paths...)
	if err != nil {
		return systemapp.ConnectivityResults{}, fmt.Errorf("load config: %w", err)
	}

	_, sqlDB, err := postgresdb.Open(cfg.Database)
	if err != nil {
		return systemapp.ConnectivityResults{}, fmt.Errorf("connect postgres: %w", err)
	}
	defer sqlDB.Close()

	service := systemapp.NewConnectivityService(
		sqlDB,
		telegramclient.New(cfg.Telegram.BotToken),
		deepseekclient.New(cfg.DeepSeek.APIKey, cfg.DeepSeek.Model),
		geminiclient.New(cfg.Gemini.APIKey, cfg.Gemini.Model),
	)
	return service.CheckAll(ctx)
}
