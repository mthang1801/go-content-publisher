package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	sourcedomain "go-content-bot/internal/source/domain"
	telegrambot "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	twitterclient "go-content-bot/internal/system/infrastructure/clients/twitter"
)

const sourceRevalidationBatchLimit = 100

type sourceValidationService interface {
	ListActiveDueForValidation(ctx context.Context, limit int) ([]sourcedomain.Source, error)
	ListInactiveDueForRecheck(ctx context.Context, now time.Time, limit int) ([]sourcedomain.Source, error)
	MarkChecked(ctx context.Context, id string, at time.Time) error
	MarkInactive(ctx context.Context, id string, reason string, at time.Time) error
	MarkActive(ctx context.Context, id string, at time.Time) error
}

type sourceValidationTelegramClient interface {
	GetChat(ctx context.Context, chatID string) (telegrambot.Chat, error)
}

type sourceValidationTwitterClient interface {
	LookupUserByUsername(ctx context.Context, username string) (twitterclient.User, error)
}

type SourceRevalidationAction struct {
	service  sourceValidationService
	telegram sourceValidationTelegramClient
	twitter  sourceValidationTwitterClient
	logger   *slog.Logger
	enabled  bool
}

func NewSourceRevalidationAction(
	service sourceValidationService,
	telegram sourceValidationTelegramClient,
	twitter sourceValidationTwitterClient,
	logger *slog.Logger,
	enabled bool,
) *SourceRevalidationAction {
	return &SourceRevalidationAction{
		service:  service,
		telegram: telegram,
		twitter:  twitter,
		logger:   logger,
		enabled:  enabled,
	}
}

func (a *SourceRevalidationAction) Name() string {
	return "source_revalidation"
}

func (a *SourceRevalidationAction) Enabled() bool {
	return a.enabled
}

func (a *SourceRevalidationAction) Crawl(ctx context.Context) error {
	if !a.enabled {
		a.logger.Info("source revalidation skipped", "reason", "source revalidation disabled")
		return nil
	}

	now := time.Now()

	active, err := a.service.ListActiveDueForValidation(ctx, sourceRevalidationBatchLimit)
	if err != nil {
		return fmt.Errorf("list active sources due for validation: %w", err)
	}
	inactive, err := a.service.ListInactiveDueForRecheck(ctx, now, sourceRevalidationBatchLimit)
	if err != nil {
		return fmt.Errorf("list inactive sources due for recheck: %w", err)
	}

	checked := 0
	activated := 0
	inactivated := 0
	var errs []error

	for _, source := range active {
		changed, err := a.validateSource(ctx, source, now)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		checked++
		if changed == "inactive" {
			inactivated++
		}
	}

	for _, source := range inactive {
		changed, err := a.validateSource(ctx, source, now)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		checked++
		if changed == "active" {
			activated++
		}
	}

	a.logger.Info("source revalidation completed",
		"checked", checked,
		"activated", activated,
		"inactivated", inactivated,
		"active_due", len(active),
		"inactive_due", len(inactive),
	)

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (a *SourceRevalidationAction) validateSource(ctx context.Context, source sourcedomain.Source, now time.Time) (string, error) {
	err := a.checkSource(ctx, source)
	if err == nil {
		if source.IsActive {
			if saveErr := a.service.MarkChecked(ctx, source.ID, now); saveErr != nil {
				return "", fmt.Errorf("mark source checked %s: %w", source.ID, saveErr)
			}
			return "", nil
		}
		if saveErr := a.service.MarkActive(ctx, source.ID, now); saveErr != nil {
			return "", fmt.Errorf("mark source active %s: %w", source.ID, saveErr)
		}
		return "active", nil
	}

	if !shouldDeactivateSource(source.Type, err) {
		return "", fmt.Errorf("validate source %s (%s): %w", source.Handle, source.Type, err)
	}

	if saveErr := a.service.MarkInactive(ctx, source.ID, err.Error(), now); saveErr != nil {
		return "", fmt.Errorf("mark source inactive %s: %w", source.ID, saveErr)
	}
	return "inactive", nil
}

func (a *SourceRevalidationAction) checkSource(ctx context.Context, source sourcedomain.Source) error {
	switch source.Type {
	case sourcedomain.TypeTwitter:
		if a.twitter == nil {
			return fmt.Errorf("twitter validator is not configured")
		}
		_, err := a.twitter.LookupUserByUsername(ctx, source.Handle)
		return err
	case sourcedomain.TypeTelegram:
		if a.telegram == nil {
			return fmt.Errorf("telegram validator is not configured")
		}
		handle, _ := splitTelegramValidationHandle(source.Handle)
		_, err := a.telegram.GetChat(ctx, handle)
		return err
	default:
		return fmt.Errorf("unsupported source type %s", source.Type)
	}
}

func shouldDeactivateSource(sourceType sourcedomain.Type, err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	switch sourceType {
	case sourcedomain.TypeTwitter:
		return strings.Contains(message, "returned no user") || strings.Contains(message, "not found error")
	case sourcedomain.TypeTelegram:
		return strings.Contains(message, "chat not found")
	default:
		return false
	}
}

func splitTelegramValidationHandle(handle string) (string, *int64) {
	parts := strings.SplitN(strings.TrimSpace(handle), "#", 2)
	base := strings.TrimSpace(parts[0])
	return base, nil
}
