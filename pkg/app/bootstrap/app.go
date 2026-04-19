package bootstrap

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	contentapp "go-content-bot/internal/content/application"
	contentjobs "go-content-bot/internal/content/infrastructure/jobs"
	contentrepo "go-content-bot/internal/content/infrastructure/persistence/repositories"
	contenthttp "go-content-bot/internal/content/presentation/http"
	sourceapp "go-content-bot/internal/source/application"
	sourcejobs "go-content-bot/internal/source/infrastructure/jobs"
	sourcerepo "go-content-bot/internal/source/infrastructure/persistence/repositories"
	sourcehttp "go-content-bot/internal/source/presentation/http"
	systemapp "go-content-bot/internal/system/application"
	deepseekclient "go-content-bot/internal/system/infrastructure/clients/deepseek"
	geminiclient "go-content-bot/internal/system/infrastructure/clients/gemini"
	telegramclient "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	twitterclient "go-content-bot/internal/system/infrastructure/clients/twitter"
	systemjobs "go-content-bot/internal/system/infrastructure/jobs"
	systemrepo "go-content-bot/internal/system/infrastructure/persistence/repositories"
	"go-content-bot/pkg/config"
	postgresdb "go-content-bot/pkg/database/postgres"
	"go-content-bot/pkg/migration"
	obslogger "go-content-bot/pkg/observability/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Config            config.Config
	Logger            *slog.Logger
	GormDB            *gorm.DB
	SQLDB             *sql.DB
	SourceService     *sourceapp.Service
	ContentService    *contentapp.Service
	WorkerService     *contentapp.WorkerService
	SystemService     *systemapp.ConnectivityService
	SettingStore      *systemrepo.SettingRepository
	DeepSeekClient    *deepseekclient.Client
	GeminiClient      *geminiclient.Client
	RewriteClient     contentapp.RewriteTextPort
	TelegramClient    *telegramclient.Client
	TwitterClient     *twitterclient.Client
	SourceRevalidator *sourcejobs.SourceRevalidationAction
}

const telegramRuntimeSettingKey = "telegram_runtime"

type telegramRuntimeSetting struct {
	BotToken       string                  `json:"bot_token"`
	PublishTargets []config.TelegramTarget `json:"publish_targets"`
	IngestTargets  []config.TelegramTarget `json:"ingest_targets"`
	AdminUserIDs   []int64                 `json:"admin_user_ids"`
}

type runtimeValueGetter interface {
	Get(ctx context.Context, key string) (string, error)
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := obslogger.New(cfg.App.LogLevel, cfg.App.LogMaxSizeMB)
	if logFilePath := obslogger.LogFilePath(); strings.TrimSpace(logFilePath) != "" {
		logger.Info("file logging enabled", "path", logFilePath)
	}

	gormDB, sqlDB, err := postgresdb.Open(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	sourceService := sourceapp.NewService(sourcerepo.NewSourceRepository(gormDB))
	contentService := contentapp.NewService(contentrepo.NewContentRepository(gormDB))
	settingRepo := systemrepo.NewSettingRepository(gormDB)
	if err := applyRuntimeSettingOverrides(settingRepo, &cfg); err != nil {
		return nil, fmt.Errorf("apply runtime setting overrides: %w", err)
	}
	telegram := telegramclient.New(cfg.Telegram.BotToken)
	twitter := twitterclient.New(
		cfg.Twitter.BearerToken,
		twitterclient.OAuthCredentials{
			APIKey:       cfg.Twitter.VIAPIKey,
			APISecret:    cfg.Twitter.VIAPISecret,
			AccessToken:  cfg.Twitter.VIAccessToken,
			AccessSecret: cfg.Twitter.VIAccessSecret,
		},
		twitterclient.OAuthCredentials{
			APIKey:       cfg.Twitter.ENAPIKey,
			APISecret:    cfg.Twitter.ENAPISecret,
			AccessToken:  cfg.Twitter.ENAccessToken,
			AccessSecret: cfg.Twitter.ENAccessSecret,
		},
	)
	deepseek := deepseekclient.New(cfg.DeepSeek.APIKey, cfg.DeepSeek.Model)
	gemini := geminiclient.New(cfg.Gemini.APIKey, cfg.Gemini.Model)
	systemService := systemapp.NewConnectivityService(
		sqlDB,
		telegram,
		deepseek,
		gemini,
	)
	rewriter, rewriteName, err := selectRewriteProvider(cfg, deepseek, gemini)
	if err != nil {
		return nil, err
	}
	rewriter, rewriteName = applyRewriteFallback(cfg, rewriter, rewriteName, deepseek, gemini)
	twitterCrawlAction := contentjobs.NewTwitterCrawlAction(
		sourceService,
		contentService,
		settingRepo,
		twitter,
		logger,
		cfg.Feature.EnableTwitterCrawler,
	)
	sourceRevalidationAction := sourcejobs.NewSourceRevalidationAction(
		sourceService,
		telegram,
		twitter,
		logger,
		cfg.Feature.EnableTelegramCrawler || cfg.Feature.EnableTwitterCrawler,
	)
	rewriteAction := contentjobs.NewRewriteAction(contentService, rewriter, settingRepo, rewriteName, logger, cfg.Feature.EnableRewriteProcessor)
	publishAction := contentjobs.NewPublishAction(
		contentService,
		contentjobs.NewTelegramPublisher(telegram, cfg.Telegram.PublishTargets),
		settingRepo,
		logger,
		cfg.Feature.AutoPublish,
	)
	telegramAdminHandler := newTelegramAdminHandler(
		cfg,
		contentService,
		sourceService,
		settingRepo,
		telegram,
		systemrepo.NewLogRepository(gormDB),
		sourceRevalidationAction,
		twitterCrawlAction,
	)
	telegramIngestAction := contentjobs.NewTelegramIngestAction(
		sourceService,
		sourceService,
		contentService,
		settingRepo,
		telegram,
		cfg.Telegram.IngestTargets,
		logger,
		cfg.Feature.EnableTelegramCrawler,
	).WithAdminHandler(telegramAdminHandler)
	crawlAction := contentjobs.NewCompositeCrawlAction("crawl", logger, sourceRevalidationAction, telegramIngestAction, twitterCrawlAction)
	twitterPublishAction := contentjobs.NewTwitterPublishAction(
		contentService,
		twitter,
		settingRepo,
		logger,
		cfg.Feature.EnableTwitterPublishVI || cfg.Feature.EnableTwitterPublishEN,
		cfg.Twitter.PublishAfter,
	)
	workerService := contentapp.NewWorkerService(
		logger,
		cfg.Scheduler,
		cfg.Feature,
		crawlAction,
		rewriteAction,
		publishAction,
		twitterPublishAction,
	)

	return &App{
		Config:            cfg,
		Logger:            logger,
		GormDB:            gormDB,
		SQLDB:             sqlDB,
		SourceService:     sourceService,
		ContentService:    contentService,
		WorkerService:     workerService,
		SystemService:     systemService,
		SettingStore:      settingRepo,
		DeepSeekClient:    deepseek,
		GeminiClient:      gemini,
		RewriteClient:     rewriter,
		TelegramClient:    telegram,
		TwitterClient:     twitter,
		SourceRevalidator: sourceRevalidationAction,
	}, nil
}

func newTelegramAdminHandler(
	cfg config.Config,
	contentService *contentapp.Service,
	sourceService *sourceapp.Service,
	settingRepo *systemrepo.SettingRepository,
	telegram *telegramclient.Client,
	logRepo *systemrepo.LogRepository,
	sourceRevalidationAction *sourcejobs.SourceRevalidationAction,
	twitterCrawlAction *contentjobs.TwitterCrawlAction,
) *systemjobs.TelegramAdminHandler {
	if strings.TrimSpace(cfg.Telegram.BotToken) == "" {
		return nil
	}
	return systemjobs.NewTelegramAdminHandler(
		cfg.Telegram.AdminUserIDs,
		contentService,
		sourceService,
		settingRepo,
		telegram,
		logRepo,
		func(ctx context.Context) error {
			var errs []error
			if sourceRevalidationAction != nil && sourceRevalidationAction.Enabled() {
				if err := sourceRevalidationAction.Crawl(ctx); err != nil {
					errs = append(errs, fmt.Errorf("source revalidation: %w", err))
				}
			}
			if twitterCrawlAction != nil && twitterCrawlAction.Enabled() {
				if err := twitterCrawlAction.Crawl(ctx); err != nil {
					errs = append(errs, fmt.Errorf("twitter crawl: %w", err))
				}
			}
			return errors.Join(errs...)
		},
	)
}

func applyRuntimeSettingOverrides(settingRepo *systemrepo.SettingRepository, cfg *config.Config) error {
	if settingRepo == nil || cfg == nil {
		return nil
	}

	ctx := context.Background()

	if err := applyTelegramRuntimeSettingOverride(ctx, settingRepo, &cfg.Telegram); err != nil {
		return err
	}
	if err := applyBoolSettingOverride(ctx, settingRepo, "auto_publish", &cfg.Feature.AutoPublish); err != nil {
		return err
	}
	if err := applyBoolSettingOverride(ctx, settingRepo, "enable_telegram_crawler", &cfg.Feature.EnableTelegramCrawler); err != nil {
		return err
	}
	if err := applyBoolSettingOverride(ctx, settingRepo, "enable_twitter_crawler", &cfg.Feature.EnableTwitterCrawler); err != nil {
		return err
	}
	if err := applyBoolSettingOverride(ctx, settingRepo, "enable_rewrite_processor", &cfg.Feature.EnableRewriteProcessor); err != nil {
		return err
	}
	if err := applyBoolSettingOverride(ctx, settingRepo, "enable_twitter_publish_vi", &cfg.Feature.EnableTwitterPublishVI); err != nil {
		return err
	}
	if err := applyBoolSettingOverride(ctx, settingRepo, "enable_twitter_publish_en", &cfg.Feature.EnableTwitterPublishEN); err != nil {
		return err
	}
	if err := applySecondsSettingOverride(ctx, settingRepo, "crawl_interval_seconds", &cfg.Scheduler.CrawlInterval); err != nil {
		return err
	}
	if err := applySecondsSettingOverride(ctx, settingRepo, "process_interval_seconds", &cfg.Scheduler.ProcessInterval); err != nil {
		return err
	}
	if err := applySecondsSettingOverride(ctx, settingRepo, "publish_interval_seconds", &cfg.Scheduler.PublishInterval); err != nil {
		return err
	}
	if err := applySecondsSettingOverride(ctx, settingRepo, "twitter_publish_interval_seconds", &cfg.Scheduler.TwitterPublishInterval); err != nil {
		return err
	}

	if value, err := settingRepo.Get(ctx, "rewrite_provider"); err != nil {
		return err
	} else if strings.TrimSpace(value) != "" {
		cfg.Rewrite.Provider = strings.TrimSpace(value)
	}

	return nil
}

func applyTelegramRuntimeSettingOverride(ctx context.Context, settingRepo *systemrepo.SettingRepository, target *config.TelegramConfig) error {
	if settingRepo == nil || target == nil {
		return nil
	}
	record, err := settingRepo.GetRecord(ctx, telegramRuntimeSettingKey)
	if err != nil {
		return err
	}
	if len(record.JSONValue) == 0 {
		return nil
	}
	return applyTelegramRuntimeJSON(record.JSONValue, target)
}

func applyTelegramRuntimeJSON(raw []byte, target *config.TelegramConfig) error {
	if target == nil || len(raw) == 0 {
		return nil
	}

	var setting telegramRuntimeSetting
	if err := json.Unmarshal(raw, &setting); err != nil {
		return fmt.Errorf("parse %s: %w", telegramRuntimeSettingKey, err)
	}

	publishTargets, err := normalizeTelegramTargets(setting.PublishTargets)
	if err != nil {
		return fmt.Errorf("parse %s publish_targets: %w", telegramRuntimeSettingKey, err)
	}
	ingestTargets, err := normalizeTelegramTargets(setting.IngestTargets)
	if err != nil {
		return fmt.Errorf("parse %s ingest_targets: %w", telegramRuntimeSettingKey, err)
	}

	target.BotToken = strings.TrimSpace(setting.BotToken)
	target.PublishTargets = publishTargets
	target.IngestTargets = ingestTargets
	target.AdminUserIDs = append([]int64(nil), setting.AdminUserIDs...)
	target.TargetChannel = ""

	return nil
}

func normalizeTelegramTargets(targets []config.TelegramTarget) ([]config.TelegramTarget, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	normalized := make([]config.TelegramTarget, len(targets))
	for i, target := range targets {
		target.ChatID = strings.TrimSpace(target.ChatID)
		if target.ChatID == "" {
			return nil, fmt.Errorf("target %d chat_id is required", i)
		}
		normalized[i] = target
	}
	return normalized, nil
}

func applyBoolSettingOverride(ctx context.Context, settingRepo runtimeValueGetter, key string, target *bool) error {
	if settingRepo == nil || target == nil {
		return nil
	}
	value, err := settingRepo.Get(ctx, key)
	if err != nil {
		return err
	}
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fmt.Errorf("parse %s: %w", key, err)
	}
	*target = parsed
	return nil
}

func applySecondsSettingOverride(ctx context.Context, settingRepo runtimeValueGetter, key string, target *time.Duration) error {
	if settingRepo == nil || target == nil {
		return nil
	}
	value, err := settingRepo.Get(ctx, key)
	if err != nil {
		return err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("parse %s: %w", key, err)
	}
	if seconds <= 0 {
		return fmt.Errorf("parse %s: seconds must be positive", key)
	}
	*target = time.Duration(seconds) * time.Second
	return nil
}

func (a *App) RunAPI(ctx context.Context) error {
	gin.DefaultWriter = obslogger.Output()
	gin.DefaultErrorWriter = obslogger.Output()

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	sourcehttp.RegisterRoutes(api, sourcehttp.NewHandler(a.SourceService))
	contenthttp.RegisterRoutes(api, contenthttp.NewHandler(a.ContentService))

	server := &http.Server{
		Addr:         httpListenAddress(a.Config.HTTP.Host, a.Config.HTTP.Port),
		Handler:      router,
		ReadTimeout:  a.Config.HTTP.ReadTimeout,
		WriteTimeout: a.Config.HTTP.WriteTimeout,
	}

	listener, err := openHTTPListener(server, a.Config.HTTP.Host, a.Config.HTTP.Port)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		addr, port, healthzURL := describeHTTPListener(listener, a.Config.HTTP.Host)
		a.Logger.Info("api listening", "addr", addr, "port", port, "healthz", healthzURL)
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.Logger.Info("api shutting down")
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) RunWorker(ctx context.Context) error {
	return a.WorkerService.Run(ctx)
}

func (a *App) RunMigrations(ctx context.Context) error {
	a.Logger.Info("running migrations", "dir", "db/migrations")
	return migration.New("db/migrations").Up(ctx, a.SQLDB)
}

func (a *App) RunConnectionChecks(ctx context.Context) (systemapp.ConnectivityResults, error) {
	return a.SystemService.CheckAll(ctx)
}

func (a *App) RunGetSetting(ctx context.Context, key string) (systemrepo.SettingRecord, error) {
	return a.SettingStore.GetRecord(ctx, key)
}

func (a *App) RunSetSetting(ctx context.Context, key, value string) error {
	return a.SettingStore.Set(ctx, key, value)
}

func (a *App) RunSetSettingJSON(ctx context.Context, key string, jsonValue []byte) error {
	return a.SettingStore.SetJSON(ctx, key, jsonValue)
}

func (a *App) RunSourceRevalidationOnce(ctx context.Context) error {
	if a.SourceRevalidator == nil {
		return nil
	}
	return a.SourceRevalidator.Crawl(ctx)
}

type ContentJobResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type TelegramTargetProbeResult struct {
	ChatID             string `json:"chat_id"`
	ThreadID           *int64 `json:"thread_id,omitempty"`
	GetChatOK          bool   `json:"get_chat_ok"`
	GetChatDescription string `json:"get_chat_description,omitempty"`
	SendOK             bool   `json:"send_ok"`
	SendDescription    string `json:"send_description,omitempty"`
	MessageID          string `json:"message_id,omitempty"`
}

func (a *App) RunProcessNext(ctx context.Context) (*ContentJobResult, error) {
	if _, err := contentjobs.SkipStalePendingForRun(ctx, a.ContentService, a.SettingStore); err != nil {
		return nil, err
	}
	item, err := a.ContentService.ProcessNextPending(ctx, a.RewriteClient)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	return &ContentJobResult{
		ID:     item.ID,
		Status: string(item.Status),
	}, nil
}

func (a *App) RunPublishNext(ctx context.Context) (*ContentJobResult, error) {
	if _, err := contentjobs.SkipStaleRewrittenForRun(ctx, a.ContentService, a.SettingStore); err != nil {
		return nil, err
	}
	item, err := a.ContentService.PublishNextReady(
		ctx,
		contentjobs.NewTelegramPublisher(a.TelegramClient, a.Config.Telegram.PublishTargets),
	)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	return &ContentJobResult{
		ID:     item.ID,
		Status: string(item.Status),
	}, nil
}

func (a *App) RunTelegramIngestOnce(ctx context.Context) error {
	sourceRevalidationAction := sourcejobs.NewSourceRevalidationAction(
		a.SourceService,
		a.TelegramClient,
		a.TwitterClient,
		a.Logger,
		a.Config.Feature.EnableTelegramCrawler || a.Config.Feature.EnableTwitterCrawler,
	)
	twitterCrawlAction := contentjobs.NewTwitterCrawlAction(
		a.SourceService,
		a.ContentService,
		systemrepo.NewSettingRepository(a.GormDB),
		a.TwitterClient,
		a.Logger,
		a.Config.Feature.EnableTwitterCrawler,
	)
	adminHandler := newTelegramAdminHandler(
		a.Config,
		a.ContentService,
		a.SourceService,
		systemrepo.NewSettingRepository(a.GormDB),
		a.TelegramClient,
		systemrepo.NewLogRepository(a.GormDB),
		sourceRevalidationAction,
		twitterCrawlAction,
	)
	action := contentjobs.NewTelegramIngestAction(
		a.SourceService,
		a.SourceService,
		a.ContentService,
		systemrepo.NewSettingRepository(a.GormDB),
		a.TelegramClient,
		a.Config.Telegram.IngestTargets,
		a.Logger,
		a.Config.Feature.EnableTelegramCrawler,
	).WithAdminHandler(adminHandler)
	return action.Crawl(ctx)
}

func (a *App) RunTwitterCrawlOnce(ctx context.Context) error {
	action := contentjobs.NewTwitterCrawlAction(
		a.SourceService,
		a.ContentService,
		systemrepo.NewSettingRepository(a.GormDB),
		a.TwitterClient,
		a.Logger,
		a.Config.Feature.EnableTwitterCrawler,
	)
	return action.Crawl(ctx)
}

func (a *App) RunTwitterPublishNext(ctx context.Context) error {
	action := contentjobs.NewTwitterPublishAction(
		a.ContentService,
		a.TwitterClient,
		systemrepo.NewSettingRepository(a.GormDB),
		a.Logger,
		a.Config.Feature.EnableTwitterPublishVI || a.Config.Feature.EnableTwitterPublishEN,
		a.Config.Twitter.PublishAfter,
	)
	return action.Publish(ctx)
}

func (a *App) RunProbeTelegramTargets(ctx context.Context, text string) ([]TelegramTargetProbeResult, error) {
	results := make([]TelegramTargetProbeResult, 0, len(a.Config.Telegram.PublishTargets))
	if len(a.Config.Telegram.PublishTargets) == 0 {
		return results, nil
	}

	message := strings.TrimSpace(text)
	if message == "" {
		message = fmt.Sprintf("[probe] telegram target test at %s", time.Now().Format(time.RFC3339))
	}

	for _, target := range a.Config.Telegram.PublishTargets {
		result := TelegramTargetProbeResult{
			ChatID:   target.ChatID,
			ThreadID: target.ThreadID,
		}

		chat, err := a.TelegramClient.GetChat(ctx, target.ChatID)
		if err != nil {
			result.GetChatDescription = err.Error()
		} else {
			result.GetChatOK = true
			if strings.TrimSpace(chat.Title) != "" {
				result.GetChatDescription = chat.Title
			} else if strings.TrimSpace(chat.Username) != "" {
				result.GetChatDescription = chat.Username
			} else {
				result.GetChatDescription = chat.Type
			}
		}

		messageID, err := a.TelegramClient.SendMessage(ctx, target.ChatID, target.ThreadID, message)
		if err != nil {
			result.SendDescription = err.Error()
		} else {
			result.SendOK = true
			result.MessageID = messageID
		}

		results = append(results, result)
	}

	return results, nil
}

func selectRewriteProvider(
	cfg config.Config,
	deepseek *deepseekclient.Client,
	gemini *geminiclient.Client,
) (contentapp.RewriteTextPort, string, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Rewrite.Provider)) {
	case "deepseek":
		return deepseek, "deepseek_rewrite", nil
	case "gemini":
		return gemini, "gemini_rewrite", nil
	default:
		return nil, "", fmt.Errorf("unsupported rewrite provider: %s", cfg.Rewrite.Provider)
	}
}

func applyRewriteFallback(
	cfg config.Config,
	primary contentapp.RewriteTextPort,
	primaryName string,
	deepseek *deepseekclient.Client,
	gemini *geminiclient.Client,
) (contentapp.RewriteTextPort, string) {
	switch strings.ToLower(strings.TrimSpace(cfg.Rewrite.Provider)) {
	case "gemini":
		if strings.TrimSpace(cfg.DeepSeek.APIKey) == "" {
			return primary, primaryName
		}
		return contentjobs.NewFallbackRewriter("gemini", gemini, "deepseek", deepseek), "gemini_rewrite_with_deepseek_fallback"
	case "deepseek":
		if strings.TrimSpace(cfg.Gemini.APIKey) == "" {
			return primary, primaryName
		}
		return contentjobs.NewFallbackRewriter("deepseek", deepseek, "gemini", gemini), "deepseek_rewrite_with_gemini_fallback"
	default:
		return primary, primaryName
	}
}

func (a *App) Close() error {
	if a.SQLDB == nil {
		return nil
	}
	return a.SQLDB.Close()
}
