package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"go-content-bot/internal/content/application"
	"go-content-bot/pkg/config"
)

const (
	rewriteDuplicateWindowHoursSettingKey        = "rewrite_duplicate_window_hours"
	rewriteDuplicateOriginalThresholdSettingKey  = "rewrite_duplicate_original_threshold"
	rewriteDuplicateRewrittenThresholdSettingKey = "rewrite_duplicate_rewritten_threshold"
	autoPublishSettingKey                        = "auto_publish"
)

type RewriteAction struct {
	service  *application.Service
	rewriter application.RewriteTextPort
	settings queueSettingsStore
	name     string
	logger   *slog.Logger
	enabled  bool
}

func NewRewriteAction(service *application.Service, rewriter application.RewriteTextPort, settings queueSettingsStore, name string, logger *slog.Logger, enabled bool) *RewriteAction {
	return &RewriteAction{
		service:  service,
		rewriter: rewriter,
		settings: settings,
		name:     name,
		logger:   logger,
		enabled:  enabled,
	}
}

func (a *RewriteAction) Name() string {
	if a.name == "" {
		return "rewrite"
	}
	return a.name
}

func (a *RewriteAction) Enabled() bool {
	return a.enabled
}

func (a *RewriteAction) Rewrite(ctx context.Context) error {
	if !a.enabled {
		a.logger.Info("rewrite action skipped", "reason", "rewrite processor disabled", "provider", a.Name())
		return nil
	}
	if skipped, err := skipStalePending(ctx, a.service, a.settings); err != nil {
		return fmt.Errorf("skip stale pending content items: %w", err)
	} else if skipped > 0 {
		a.logger.Info("rewrite stale items skipped", "count", skipped, "provider", a.Name())
	}

	options, err := resolveRewriteOptions(ctx, a.settings)
	if err != nil {
		return fmt.Errorf("resolve rewrite options: %w", err)
	}

	item, err := a.service.ProcessNextPendingWithOptions(ctx, a.rewriter, options)
	if err != nil {
		return fmt.Errorf("process next pending content item: %w", err)
	}
	if item == nil {
		a.logger.Info("rewrite action idle", "reason", "no pending items")
		return nil
	}

	a.logger.Info("rewrite action completed", "content_id", item.ID, "status", string(item.Status), "provider", a.Name())
	return nil
}

func resolveRewriteOptions(ctx context.Context, settings queueSettingsStore) (application.RewriteOptions, error) {
	options := application.RewriteOptions{}
	if settings == nil {
		return options, nil
	}

	windowHours, err := getFloatSetting(ctx, settings, rewriteDuplicateWindowHoursSettingKey)
	if err != nil {
		return options, err
	}
	if windowHours > 0 {
		options.DuplicateWindow = time.Duration(windowHours * float64(time.Hour))
	}

	originalThreshold, err := getFloatSetting(ctx, settings, rewriteDuplicateOriginalThresholdSettingKey)
	if err != nil {
		return options, err
	}
	options.DuplicateOriginalThreshold = originalThreshold

	rewrittenThreshold, err := getFloatSetting(ctx, settings, rewriteDuplicateRewrittenThresholdSettingKey)
	if err != nil {
		return options, err
	}
	options.DuplicateRewrittenThreshold = rewrittenThreshold

	return options, nil
}

func getFloatSetting(ctx context.Context, settings queueSettingsStore, key string) (float64, error) {
	value, err := settings.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

type PublishAction struct {
	service   *application.Service
	publisher application.PublishTextPort
	settings  queueSettingsStore
	logger    *slog.Logger
	enabled   bool
}

func NewPublishAction(service *application.Service, publisher application.PublishTextPort, settings queueSettingsStore, logger *slog.Logger, enabled bool) *PublishAction {
	return &PublishAction{
		service:   service,
		publisher: publisher,
		settings:  settings,
		logger:    logger,
		enabled:   enabled,
	}
}

func (a *PublishAction) Name() string {
	return "telegram_publish"
}

func (a *PublishAction) Enabled() bool {
	return a.enabled
}

func (a *PublishAction) Publish(ctx context.Context) error {
	enabled, err := a.resolveEnabled(ctx)
	if err != nil {
		return err
	}
	if !enabled {
		a.logger.Info("publish action skipped", "reason", "telegram publish disabled")
		return nil
	}
	if skipped, err := skipStaleRewritten(ctx, a.service, a.settings); err != nil {
		return fmt.Errorf("skip stale rewritten content items: %w", err)
	} else if skipped > 0 {
		a.logger.Info("publish stale items skipped", "count", skipped)
	}

	item, err := a.service.PublishNextReady(ctx, a.publisher)
	if err != nil {
		return fmt.Errorf("publish next rewritten content item: %w", err)
	}
	if item == nil {
		a.logger.Info("publish action idle", "reason", "no rewritten items")
		return nil
	}

	a.logger.Info("publish action completed", "content_id", item.ID, "status", string(item.Status))
	return nil
}

func (a *PublishAction) resolveEnabled(ctx context.Context) (bool, error) {
	enabled := a.enabled
	if a.settings == nil {
		return enabled, nil
	}
	value, err := a.settings.Get(ctx, autoPublishSettingKey)
	if err != nil {
		return false, fmt.Errorf("read %s setting: %w", autoPublishSettingKey, err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return enabled, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s setting: %w", autoPublishSettingKey, err)
	}
	return parsed, nil
}

type TelegramPublisher struct {
	client interface {
		SendMessage(ctx context.Context, chatID string, threadID *int64, text string) (string, error)
	}
	targets []config.TelegramTarget
}

func NewTelegramPublisher(client interface {
	SendMessage(ctx context.Context, chatID string, threadID *int64, text string) (string, error)
}, targets []config.TelegramTarget) *TelegramPublisher {
	return &TelegramPublisher{
		client:  client,
		targets: targets,
	}
}

func (p *TelegramPublisher) PublishText(ctx context.Context, text string) (string, error) {
	if len(p.targets) == 0 {
		return "", errors.New("telegram publish targets are not configured")
	}

	type publishResult struct {
		ChatID    string `json:"chat_id"`
		ThreadID  *int64 `json:"thread_id,omitempty"`
		MessageID string `json:"message_id"`
	}

	results := make([]publishResult, 0, len(p.targets))
	for _, target := range p.targets {
		messageID, err := p.client.SendMessage(ctx, target.ChatID, target.ThreadID, text)
		if err != nil {
			return "", err
		}
		results = append(results, publishResult{
			ChatID:    target.ChatID,
			ThreadID:  target.ThreadID,
			MessageID: messageID,
		})
	}

	encoded, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("marshal telegram publish results: %w", err)
	}
	return string(encoded), nil
}
