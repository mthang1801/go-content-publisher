package jobs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	contentapp "go-content-bot/internal/content/application"
	contentdomain "go-content-bot/internal/content/domain"
	sourcedomain "go-content-bot/internal/source/domain"
	telegrambot "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	systemrepo "go-content-bot/internal/system/infrastructure/persistence/repositories"
)

const (
	adminQueueLimit       = 10
	adminRecentLimit      = 10
	adminSourceLimit      = 25
	adminLogsLimit        = 10
	adminRetryFailedLimit = 100
)

type TelegramAdminHandler struct {
	adminUserIDs map[int64]struct{}
	content      telegramAdminContentService
	sources      telegramAdminSourceService
	settings     telegramAdminSettingsStore
	sender       telegramAdminSender
	logs         telegramAdminLogLister
	crawlNow     func(context.Context) error
}

type telegramAdminContentService interface {
	ListQueue(ctx context.Context, limit int) ([]contentdomain.ContentItem, error)
	ListRecent(ctx context.Context, limit int) ([]contentdomain.ContentItem, error)
	RetryFailed(ctx context.Context, limit int) (int, error)
	SkipByIDPrefix(ctx context.Context, prefix, reason string) (contentdomain.ContentItem, error)
	EnqueueManual(ctx context.Context, input contentapp.EnqueueManualInput) (contentdomain.ContentItem, error)
}

type telegramAdminSourceService interface {
	ListAll(ctx context.Context) ([]sourcedomain.Source, error)
	Create(ctx context.Context, source sourcedomain.Source) (sourcedomain.Source, error)
	MarkActive(ctx context.Context, id string, at time.Time) error
	MarkInactive(ctx context.Context, id string, reason string, at time.Time) error
}

type telegramAdminSettingsStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

type telegramAdminSender interface {
	SendMessage(ctx context.Context, chatID string, threadID *int64, text string) (string, error)
}

type telegramAdminLogLister interface {
	ListRecent(ctx context.Context, limit int) ([]systemrepo.LogRecord, error)
}

func NewTelegramAdminHandler(
	adminUserIDs []int64,
	content telegramAdminContentService,
	sources telegramAdminSourceService,
	settings telegramAdminSettingsStore,
	sender telegramAdminSender,
	logs telegramAdminLogLister,
	crawlNow func(context.Context) error,
) *TelegramAdminHandler {
	allowed := make(map[int64]struct{}, len(adminUserIDs))
	for _, id := range adminUserIDs {
		if id != 0 {
			allowed[id] = struct{}{}
		}
	}
	return &TelegramAdminHandler{
		adminUserIDs: allowed,
		content:      content,
		sources:      sources,
		settings:     settings,
		sender:       sender,
		logs:         logs,
		crawlNow:     crawlNow,
	}
}

func (h *TelegramAdminHandler) Handle(ctx context.Context, message telegrambot.Message) (bool, error) {
	text := strings.TrimSpace(message.Text)
	if text == "" {
		text = strings.TrimSpace(message.Caption)
	}
	if text == "" {
		return false, nil
	}
	if !strings.HasPrefix(text, "/") {
		if !h.isAuthorized(message.From) {
			return false, nil
		}
		reply, err := h.addManual(ctx, text, message.From)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	}

	if !h.isAuthorized(message.From) {
		return true, h.reply(ctx, message, "Not authorized.")
	}

	command, args := parseAdminCommand(text)
	switch command {
	case "/add":
		if len(args) == 0 {
			return true, h.reply(ctx, message, "Usage: /add <text>")
		}
		reply, err := h.addManual(ctx, strings.TrimSpace(strings.Join(args, " ")), message.From)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/start", "/help":
		return true, h.reply(ctx, message, adminUsage())
	case "/status":
		reply, err := h.status(ctx)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/queue":
		reply, err := h.queue(ctx)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/recent":
		reply, err := h.recent(ctx)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/sources":
		reply, err := h.sourcesReport(ctx)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/addsource":
		reply, err := h.addSource(ctx, args)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/removesource":
		reply, err := h.removeSource(ctx, args)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/retry":
		retried, err := h.content.RetryFailed(ctx, adminRetryFailedLimit)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, fmt.Sprintf("Retried %d failed item(s).", retried))
	case "/skip":
		if len(args) < 1 {
			return true, h.reply(ctx, message, "Usage: /skip <content-id-prefix>")
		}
		item, err := h.content.SkipByIDPrefix(ctx, args[0], "skipped by Telegram admin command")
		if err != nil {
			return true, h.reply(ctx, message, "Skip failed: "+err.Error())
		}
		return true, h.reply(ctx, message, fmt.Sprintf("Skipped %s.", shortID(item.ID)))
	case "/pause":
		if err := h.settings.Set(ctx, "auto_publish", "false"); err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, "auto_publish=false")
	case "/resume":
		if err := h.settings.Set(ctx, "auto_publish", "true"); err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, "auto_publish=true")
	case "/logs":
		reply, err := h.logsReport(ctx)
		if err != nil {
			return true, err
		}
		return true, h.reply(ctx, message, reply)
	case "/crawlnow":
		if h.crawlNow == nil {
			return true, h.reply(ctx, message, "crawlnow is not configured in this runtime.")
		}
		if err := h.crawlNow(ctx); err != nil {
			return true, h.reply(ctx, message, "Crawl failed: "+err.Error())
		}
		return true, h.reply(ctx, message, "Crawl completed.")
	default:
		return true, h.reply(ctx, message, "Unknown command. Use /help.")
	}
}

func (h *TelegramAdminHandler) addManual(ctx context.Context, text string, from *telegrambot.User) (string, error) {
	item, err := h.content.EnqueueManual(ctx, contentapp.EnqueueManualInput{
		Text:   text,
		Author: manualAuthorName(from),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Added manual content %s: %s", shortID(item.ID), truncate(item.OriginalText, 80)), nil
}

func manualAuthorName(user *telegrambot.User) string {
	if user == nil {
		return "Manual"
	}
	if value := strings.TrimSpace(user.FirstName); value != "" {
		return value
	}
	if value := strings.TrimSpace(user.Username); value != "" {
		return value
	}
	return "Manual"
}

func (h *TelegramAdminHandler) isAuthorized(user *telegrambot.User) bool {
	if user == nil {
		return false
	}
	if len(h.adminUserIDs) == 0 {
		return true
	}
	_, ok := h.adminUserIDs[user.ID]
	return ok
}

func (h *TelegramAdminHandler) reply(ctx context.Context, message telegrambot.Message, text string) error {
	if h.sender == nil {
		return errors.New("telegram admin sender is not configured")
	}
	_, err := h.sender.SendMessage(ctx, strconv.FormatInt(message.Chat.ID, 10), message.MessageThreadID, text)
	return err
}

func (h *TelegramAdminHandler) status(ctx context.Context) (string, error) {
	queue, err := h.content.ListQueue(ctx, 200)
	if err != nil {
		return "", err
	}
	sources, err := h.sources.ListAll(ctx)
	if err != nil {
		return "", err
	}
	autoPublish, err := h.settings.Get(ctx, "auto_publish")
	if err != nil {
		return "", err
	}

	statusCounts := map[contentdomain.Status]int{}
	for _, item := range queue {
		statusCounts[item.Status]++
	}

	activeSources := 0
	for _, source := range sources {
		if source.IsActive {
			activeSources++
		}
	}

	return fmt.Sprintf(
		"Status\n"+
			"auto_publish: %s\n"+
			"sources: %d active / %d total\n"+
			"queue: pending=%d processing=%d rewritten=%d publishing=%d failed=%d",
		valueOrUnset(autoPublish),
		activeSources,
		len(sources),
		statusCounts[contentdomain.StatusPending],
		statusCounts[contentdomain.StatusProcessing],
		statusCounts[contentdomain.StatusRewritten],
		statusCounts[contentdomain.StatusPublishing],
		statusCounts[contentdomain.StatusFailed],
	), nil
}

func (h *TelegramAdminHandler) queue(ctx context.Context) (string, error) {
	items, err := h.content.ListQueue(ctx, adminQueueLimit)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "Queue is empty.", nil
	}
	lines := []string{"Queue"}
	for _, item := range items {
		lines = append(lines, formatContentLine(item))
	}
	return strings.Join(lines, "\n"), nil
}

func (h *TelegramAdminHandler) recent(ctx context.Context) (string, error) {
	items, err := h.content.ListRecent(ctx, adminRecentLimit)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "No recent content.", nil
	}
	lines := []string{"Recent"}
	for _, item := range items {
		lines = append(lines, formatContentLine(item))
	}
	return strings.Join(lines, "\n"), nil
}

func (h *TelegramAdminHandler) sourcesReport(ctx context.Context) (string, error) {
	sources, err := h.sources.ListAll(ctx)
	if err != nil {
		return "", err
	}
	if len(sources) == 0 {
		return "No sources configured.", nil
	}
	lines := []string{"Sources"}
	for i, source := range sources {
		if i >= adminSourceLimit {
			lines = append(lines, fmt.Sprintf("...and %d more", len(sources)-i))
			break
		}
		state := "inactive"
		if source.IsActive {
			state = "active"
		}
		lines = append(lines, fmt.Sprintf("%s %s %s (%s)", state, source.Type, source.Handle, truncate(source.Name, 40)))
	}
	return strings.Join(lines, "\n"), nil
}

func (h *TelegramAdminHandler) addSource(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "Usage: /addsource <twitter|telegram> <handle> [name]", nil
	}
	sourceType := sourcedomain.Type(strings.ToLower(strings.TrimSpace(args[0])))
	handle := strings.TrimSpace(args[1])
	name := handle
	if len(args) > 2 {
		name = strings.TrimSpace(strings.Join(args[2:], " "))
	}
	source, err := h.sources.Create(ctx, sourcedomain.Source{
		Type:   sourceType,
		Handle: handle,
		Name:   name,
	})
	if err != nil {
		if errors.Is(err, sourcedomain.ErrInvalidSource) {
			return "Invalid source. Type must be twitter or telegram, and handle/name are required.", nil
		}
		if errors.Is(err, sourcedomain.ErrSourceAlreadyExists) {
			return h.reactivateExistingSource(ctx, sourceType, handle)
		}
		return "", err
	}
	return fmt.Sprintf("Added %s %s.", source.Type, source.Handle), nil
}

func (h *TelegramAdminHandler) reactivateExistingSource(ctx context.Context, sourceType sourcedomain.Type, handle string) (string, error) {
	sources, err := h.sources.ListAll(ctx)
	if err != nil {
		return "", err
	}
	for _, source := range sources {
		if source.Type == sourceType && equivalentSourceHandle(source.Handle, handle) {
			if source.IsActive {
				return "Source already exists.", nil
			}
			if err := h.sources.MarkActive(ctx, source.ID, time.Now()); err != nil {
				return "", err
			}
			return fmt.Sprintf("Reactivated %s %s.", source.Type, source.Handle), nil
		}
	}
	return "Source already exists.", nil
}

func (h *TelegramAdminHandler) removeSource(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "Usage: /removesource [twitter|telegram] <handle>", nil
	}

	var sourceType *sourcedomain.Type
	handleArg := args[0]
	if len(args) >= 2 {
		parsed := sourcedomain.Type(strings.ToLower(strings.TrimSpace(args[0])))
		sourceType = &parsed
		handleArg = args[1]
	}

	sources, err := h.sources.ListAll(ctx)
	if err != nil {
		return "", err
	}

	var matches []sourcedomain.Source
	for _, source := range sources {
		if sourceType != nil && source.Type != *sourceType {
			continue
		}
		if equivalentSourceHandle(source.Handle, handleArg) {
			matches = append(matches, source)
		}
	}
	if len(matches) == 0 {
		return "Source not found.", nil
	}
	if len(matches) > 1 {
		return "Multiple sources match. Use /removesource <twitter|telegram> <handle>.", nil
	}

	source := matches[0]
	if err := h.sources.MarkInactive(ctx, source.ID, "deactivated by Telegram admin command", time.Now()); err != nil {
		return "", err
	}
	return fmt.Sprintf("Deactivated %s %s.", source.Type, source.Handle), nil
}

func (h *TelegramAdminHandler) logsReport(ctx context.Context) (string, error) {
	if h.logs == nil {
		return "Logs persistence is not configured.", nil
	}
	logs, err := h.logs.ListRecent(ctx, adminLogsLimit)
	if err != nil {
		return "", err
	}
	if len(logs) == 0 {
		return "No persisted logs.", nil
	}
	lines := []string{"Logs"}
	for _, row := range logs {
		lines = append(lines, fmt.Sprintf("%s %s %s: %s", row.CreatedAt.Format(time.RFC3339), row.Level, row.Module, truncate(row.Message, 80)))
	}
	return strings.Join(lines, "\n"), nil
}

func parseAdminCommand(text string) (string, []string) {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) == 0 {
		return "", nil
	}
	command := strings.ToLower(parts[0])
	if at := strings.Index(command, "@"); at >= 0 {
		command = command[:at]
	}
	return command, parts[1:]
}

func adminUsage() string {
	return strings.Join([]string{
		"Commands:",
		"/add <text>",
		"/status",
		"/queue",
		"/recent",
		"/sources",
		"/addsource <twitter|telegram> <handle> [name]",
		"/removesource [twitter|telegram] <handle>",
		"/retry",
		"/skip <content-id-prefix>",
		"/pause",
		"/resume",
		"/logs",
		"/crawlnow",
		"Any non-command text from an authorized admin is queued as manual content.",
	}, "\n")
}

func formatContentLine(item contentdomain.ContentItem) string {
	return fmt.Sprintf("%s %s %s", shortID(item.ID), item.Status, truncate(item.OriginalText, 70))
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func truncate(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func valueOrUnset(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "<unset>"
	}
	return value
}

func equivalentSourceHandle(left, right string) bool {
	return normalizeSourceHandle(left) == normalizeSourceHandle(right)
}

func normalizeSourceHandle(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.TrimPrefix(value, "@")
}
