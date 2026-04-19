package jobs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	sourcedomain "go-content-bot/internal/source/domain"
	telegrambot "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	twitterclient "go-content-bot/internal/system/infrastructure/clients/twitter"
)

type sourceValidationServiceStub struct {
	activeFn       func(ctx context.Context, limit int) ([]sourcedomain.Source, error)
	inactiveFn     func(ctx context.Context, now time.Time, limit int) ([]sourcedomain.Source, error)
	markCheckedFn  func(ctx context.Context, id string, at time.Time) error
	markInactiveFn func(ctx context.Context, id string, reason string, at time.Time) error
	markActiveFn   func(ctx context.Context, id string, at time.Time) error
}

func (s sourceValidationServiceStub) ListActiveDueForValidation(ctx context.Context, limit int) ([]sourcedomain.Source, error) {
	return s.activeFn(ctx, limit)
}

func (s sourceValidationServiceStub) ListInactiveDueForRecheck(ctx context.Context, now time.Time, limit int) ([]sourcedomain.Source, error) {
	return s.inactiveFn(ctx, now, limit)
}

func (s sourceValidationServiceStub) MarkChecked(ctx context.Context, id string, at time.Time) error {
	return s.markCheckedFn(ctx, id, at)
}

func (s sourceValidationServiceStub) MarkInactive(ctx context.Context, id string, reason string, at time.Time) error {
	return s.markInactiveFn(ctx, id, reason, at)
}

func (s sourceValidationServiceStub) MarkActive(ctx context.Context, id string, at time.Time) error {
	return s.markActiveFn(ctx, id, at)
}

type sourceValidationTelegramStub struct {
	getChatFn func(ctx context.Context, chatID string) (telegrambot.Chat, error)
}

func (s sourceValidationTelegramStub) GetChat(ctx context.Context, chatID string) (telegrambot.Chat, error) {
	return s.getChatFn(ctx, chatID)
}

type sourceValidationTwitterStub struct {
	lookupFn func(ctx context.Context, username string) (twitterclient.User, error)
}

func (s sourceValidationTwitterStub) LookupUserByUsername(ctx context.Context, username string) (twitterclient.User, error) {
	return s.lookupFn(ctx, username)
}

func TestSourceRevalidationActionMarksInvalidActiveSourceInactive(t *testing.T) {
	t.Parallel()

	marked := ""
	action := NewSourceRevalidationAction(
		sourceValidationServiceStub{
			activeFn: func(ctx context.Context, limit int) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{{ID: "src-twitter", Type: sourcedomain.TypeTwitter, Handle: "@missing", Name: "Missing", IsActive: true}}, nil
			},
			inactiveFn:    func(ctx context.Context, now time.Time, limit int) ([]sourcedomain.Source, error) { return nil, nil },
			markCheckedFn: func(ctx context.Context, id string, at time.Time) error { return nil },
			markInactiveFn: func(ctx context.Context, id string, reason string, at time.Time) error {
				marked = id + "|" + reason
				return nil
			},
			markActiveFn: func(ctx context.Context, id string, at time.Time) error { return nil },
		},
		sourceValidationTelegramStub{getChatFn: func(ctx context.Context, chatID string) (telegrambot.Chat, error) { return telegrambot.Chat{}, nil }},
		sourceValidationTwitterStub{lookupFn: func(ctx context.Context, username string) (twitterclient.User, error) {
			return twitterclient.User{}, errors.New("twitter user lookup returned no user")
		}},
		revalidationLogger(),
		true,
	)

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("revalidation crawl: %v", err)
	}
	if marked == "" {
		t.Fatal("expected invalid active source to be marked inactive")
	}
}

func TestSourceRevalidationActionReactivatesInactiveSourceWhenProviderSucceeds(t *testing.T) {
	t.Parallel()

	activated := ""
	action := NewSourceRevalidationAction(
		sourceValidationServiceStub{
			activeFn: func(ctx context.Context, limit int) ([]sourcedomain.Source, error) { return nil, nil },
			inactiveFn: func(ctx context.Context, now time.Time, limit int) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{{ID: "src-telegram", Type: sourcedomain.TypeTelegram, Handle: "@restored", Name: "Restored", IsActive: false}}, nil
			},
			markCheckedFn:  func(ctx context.Context, id string, at time.Time) error { return nil },
			markInactiveFn: func(ctx context.Context, id string, reason string, at time.Time) error { return nil },
			markActiveFn: func(ctx context.Context, id string, at time.Time) error {
				activated = id
				return nil
			},
		},
		sourceValidationTelegramStub{getChatFn: func(ctx context.Context, chatID string) (telegrambot.Chat, error) {
			return telegrambot.Chat{ID: -100123, Username: "restored"}, nil
		}},
		sourceValidationTwitterStub{lookupFn: func(ctx context.Context, username string) (twitterclient.User, error) {
			return twitterclient.User{}, nil
		}},
		revalidationLogger(),
		true,
	)

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("revalidation crawl: %v", err)
	}
	if activated != "src-telegram" {
		t.Fatalf("expected src-telegram to be reactivated, got %q", activated)
	}
}

func revalidationLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
