package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	contentapp "go-content-bot/internal/content/application"
	contentdomain "go-content-bot/internal/content/domain"
	sourcedomain "go-content-bot/internal/source/domain"
	telegrambot "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	"go-content-bot/pkg/config"
)

const telegramUpdateOffsetSettingKey = "telegram_bot_update_offset"

type telegramSourceLister interface {
	ListActive(ctx context.Context) ([]sourcedomain.Source, error)
}

type telegramSourceToucher interface {
	TouchCrawl(ctx context.Context, id string, at time.Time) error
}

type telegramContentEnqueuer interface {
	EnqueueFromSource(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error)
}

type telegramSettingsStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

type telegramUpdatesClient interface {
	GetUpdates(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error)
}

type telegramAdminCommandHandler interface {
	Handle(ctx context.Context, message telegrambot.Message) (bool, error)
}

type TelegramIngestAction struct {
	sourceService telegramSourceLister
	sourceToucher telegramSourceToucher
	content       telegramContentEnqueuer
	settings      telegramSettingsStore
	client        telegramUpdatesClient
	adminHandler  telegramAdminCommandHandler
	targets       []config.TelegramTarget
	logger        *slog.Logger
	enabled       bool
}

func NewTelegramIngestAction(
	sourceService telegramSourceLister,
	sourceToucher telegramSourceToucher,
	content telegramContentEnqueuer,
	settings telegramSettingsStore,
	client telegramUpdatesClient,
	targets []config.TelegramTarget,
	logger *slog.Logger,
	enabled bool,
) *TelegramIngestAction {
	return &TelegramIngestAction{
		sourceService: sourceService,
		sourceToucher: sourceToucher,
		content:       content,
		settings:      settings,
		client:        client,
		targets:       targets,
		logger:        logger,
		enabled:       enabled,
	}
}

func (a *TelegramIngestAction) WithAdminHandler(handler telegramAdminCommandHandler) *TelegramIngestAction {
	a.adminHandler = handler
	return a
}

func (a *TelegramIngestAction) Name() string {
	return "telegram_ingest"
}

func (a *TelegramIngestAction) Enabled() bool {
	return a.enabled || a.adminHandler != nil
}

func (a *TelegramIngestAction) Crawl(ctx context.Context) error {
	if !a.enabled && a.adminHandler == nil {
		a.logger.Info("telegram ingest skipped", "reason", "telegram crawler and admin handler disabled")
		return nil
	}

	if len(a.targets) == 0 {
		a.logger.Info("telegram ingest idle", "reason", "no configured ingest targets")
		return nil
	}

	var telegramSources []sourcedomain.Source
	if a.enabled {
		sources, err := a.sourceService.ListActive(ctx)
		if err != nil {
			return fmt.Errorf("list active sources: %w", err)
		}

		telegramSources = filterTelegramSources(sources)
		if len(telegramSources) == 0 && a.adminHandler == nil {
			a.logger.Info("telegram ingest idle", "reason", "no active telegram sources")
			return nil
		}
	}

	offset, err := a.readOffset(ctx)
	if err != nil {
		return err
	}

	updates, err := a.client.GetUpdates(ctx, offset, 100)
	if err != nil {
		return fmt.Errorf("get telegram updates: %w", err)
	}

	nextOffset := offset
	ingested := 0
	for _, update := range updates {
		if update.UpdateID >= nextOffset {
			nextOffset = update.UpdateID + 1
		}

		message := selectTelegramMessage(update)
		if message == nil {
			continue
		}
		if !matchesTelegramTarget(a.targets, *message) {
			continue
		}

		if a.adminHandler != nil {
			handled, err := a.adminHandler.Handle(ctx, *message)
			if err != nil {
				return fmt.Errorf("handle telegram admin command: %w", err)
			}
			if handled {
				continue
			}
		}

		if !a.enabled {
			continue
		}

		source, ok := matchTelegramSource(telegramSources, *message)
		if !ok {
			continue
		}

		text := extractTelegramText(*message)
		if text == "" {
			continue
		}

		_, err := a.content.EnqueueFromSource(ctx, contentapp.EnqueueSourceInput{
			SourceID:     source.ID,
			ExternalID:   buildTelegramExternalID(*message),
			OriginalText: text,
			AuthorName:   resolveAuthorName(source, message.Chat),
			SourceURL:    buildTelegramSourceURL(message.Chat.Username, message.MessageID),
			CrawledAt:    time.Now(),
		})
		if err != nil {
			if errors.Is(err, contentdomain.ErrContentAlreadyExists) {
				if a.sourceToucher != nil {
					if touchErr := a.sourceToucher.TouchCrawl(ctx, source.ID, time.Now()); touchErr != nil {
						return fmt.Errorf("touch telegram source crawl %s: %w", source.ID, touchErr)
					}
				}
				a.logger.Info("telegram ingest duplicate skipped", "source_id", source.ID, "external_id", buildTelegramExternalID(*message))
				continue
			}
			return fmt.Errorf("enqueue telegram content from source %s: %w", source.ID, err)
		}

		if a.sourceToucher != nil {
			if err := a.sourceToucher.TouchCrawl(ctx, source.ID, time.Now()); err != nil {
				return fmt.Errorf("touch telegram source crawl %s: %w", source.ID, err)
			}
		}
		ingested++
	}

	if nextOffset > offset {
		if err := a.settings.Set(ctx, telegramUpdateOffsetSettingKey, strconv.FormatInt(nextOffset, 10)); err != nil {
			return fmt.Errorf("store telegram update offset: %w", err)
		}
	}

	if ingested == 0 {
		a.logger.Info("telegram ingest completed", "ingested", 0, "updates", len(updates), "offset", nextOffset)
		return nil
	}

	a.logger.Info("telegram ingest completed", "ingested", ingested, "updates", len(updates), "offset", nextOffset)
	return nil
}

func (a *TelegramIngestAction) readOffset(ctx context.Context) (int64, error) {
	value, err := a.settings.Get(ctx, telegramUpdateOffsetSettingKey)
	if err != nil {
		return 0, fmt.Errorf("read telegram update offset: %w", err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	offset, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse telegram update offset %q: %w", value, err)
	}
	return offset, nil
}

func filterTelegramSources(sources []sourcedomain.Source) []sourcedomain.Source {
	filtered := make([]sourcedomain.Source, 0, len(sources))
	for _, source := range sources {
		if source.Type == sourcedomain.TypeTelegram {
			filtered = append(filtered, source)
		}
	}
	return filtered
}

func selectTelegramMessage(update telegrambot.Update) *telegrambot.Message {
	if update.ChannelPost != nil {
		return update.ChannelPost
	}
	return update.Message
}

func matchTelegramSource(sources []sourcedomain.Source, message telegrambot.Message) (sourcedomain.Source, bool) {
	chatID := strconv.FormatInt(message.Chat.ID, 10)
	username := normalizeTelegramHandle(message.Chat.Username)

	for _, source := range sources {
		handle, threadID := splitTelegramSourceHandle(source.Handle)
		if handle == "" {
			continue
		}
		if threadID != nil {
			if message.MessageThreadID == nil || *message.MessageThreadID != *threadID {
				continue
			}
		}
		if username != "" && handle == username {
			return source, true
		}
		if handle == chatID {
			return source, true
		}
	}
	return sourcedomain.Source{}, false
}

func splitTelegramSourceHandle(handle string) (string, *int64) {
	parts := strings.SplitN(strings.TrimSpace(handle), "#", 2)
	base := normalizeTelegramHandle(parts[0])
	if len(parts) == 1 || strings.TrimSpace(parts[1]) == "" {
		return base, nil
	}

	threadID, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return base, nil
	}
	return base, &threadID
}

func matchesTelegramTarget(targets []config.TelegramTarget, message telegrambot.Message) bool {
	chatID := strconv.FormatInt(message.Chat.ID, 10)
	for _, target := range targets {
		if strings.TrimSpace(target.ChatID) != chatID {
			continue
		}
		if target.ThreadID == nil {
			return true
		}
		if message.MessageThreadID != nil && *message.MessageThreadID == *target.ThreadID {
			return true
		}
	}
	return false
}

func normalizeTelegramHandle(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "@")
	return value
}

func extractTelegramText(message telegrambot.Message) string {
	text := strings.TrimSpace(message.Text)
	if text != "" {
		return text
	}
	return strings.TrimSpace(message.Caption)
}

func resolveAuthorName(source sourcedomain.Source, chat telegrambot.Chat) string {
	if strings.TrimSpace(source.Name) != "" {
		return source.Name
	}
	if strings.TrimSpace(chat.Title) != "" {
		return chat.Title
	}
	if strings.TrimSpace(chat.Username) != "" {
		return chat.Username
	}
	return "Telegram"
}

func buildTelegramSourceURL(username string, messageID int64) *string {
	clean := normalizeTelegramHandle(username)
	if clean == "" {
		return nil
	}
	url := fmt.Sprintf("https://t.me/%s/%d", clean, messageID)
	return &url
}

func buildTelegramExternalID(message telegrambot.Message) string {
	if message.MessageThreadID != nil {
		return fmt.Sprintf("tg:%d:%d:%d", message.Chat.ID, *message.MessageThreadID, message.MessageID)
	}
	return fmt.Sprintf("tg:%d:%d", message.Chat.ID, message.MessageID)
}
