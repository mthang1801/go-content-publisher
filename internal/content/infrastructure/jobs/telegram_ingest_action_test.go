package jobs

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	contentapp "go-content-bot/internal/content/application"
	contentdomain "go-content-bot/internal/content/domain"
	sourcedomain "go-content-bot/internal/source/domain"
	telegrambot "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	"go-content-bot/pkg/config"
)

type telegramIngestSourceServiceStub struct {
	listActiveFn func(ctx context.Context) ([]sourcedomain.Source, error)
	touchCrawlFn func(ctx context.Context, id string, at time.Time) error
}

func (s telegramIngestSourceServiceStub) ListActive(ctx context.Context) ([]sourcedomain.Source, error) {
	return s.listActiveFn(ctx)
}

func (s telegramIngestSourceServiceStub) TouchCrawl(ctx context.Context, id string, at time.Time) error {
	if s.touchCrawlFn == nil {
		return nil
	}
	return s.touchCrawlFn(ctx, id, at)
}

type telegramIngestContentServiceStub struct {
	enqueueFromSourceFn func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error)
}

func (s telegramIngestContentServiceStub) EnqueueFromSource(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
	return s.enqueueFromSourceFn(ctx, input)
}

type telegramSettingsStoreStub struct {
	getFn func(ctx context.Context, key string) (string, error)
	setFn func(ctx context.Context, key, value string) error
}

func (s telegramSettingsStoreStub) Get(ctx context.Context, key string) (string, error) {
	return s.getFn(ctx, key)
}

func (s telegramSettingsStoreStub) Set(ctx context.Context, key, value string) error {
	return s.setFn(ctx, key, value)
}

type telegramUpdatesClientStub struct {
	getUpdatesFn func(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error)
}

func (s telegramUpdatesClientStub) GetUpdates(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error) {
	return s.getUpdatesFn(ctx, offset, limit)
}

type telegramAdminCommandHandlerStub struct {
	handleFn func(ctx context.Context, message telegrambot.Message) (bool, error)
}

func (s telegramAdminCommandHandlerStub) Handle(ctx context.Context, message telegrambot.Message) (bool, error) {
	return s.handleFn(ctx, message)
}

func TestTelegramIngestActionCrawlsMatchingSourcesAndUpdatesOffset(t *testing.T) {
	t.Parallel()

	enqueued := make([]contentapp.EnqueueSourceInput, 0)
	stored := map[string]string{}
	threadID := int64(5)

	action := NewTelegramIngestAction(
		telegramIngestSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-telegram", Type: sourcedomain.TypeTelegram, Handle: "@coding", Name: "Coding"},
					{ID: "src-twitter", Type: sourcedomain.TypeTwitter, Handle: "@other", Name: "Other"},
				}, nil
			},
		},
		telegramIngestSourceServiceStub{},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				enqueued = append(enqueued, input)
				return contentdomain.ContentItem{ID: "item-" + input.ExternalID}, nil
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				if key != telegramUpdateOffsetSettingKey {
					t.Fatalf("unexpected get key %s", key)
				}
				return "5", nil
			},
			setFn: func(ctx context.Context, key, value string) error {
				stored[key] = value
				return nil
			},
		},
		telegramUpdatesClientStub{
			getUpdatesFn: func(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error) {
				if offset != 5 {
					t.Fatalf("expected offset 5, got %d", offset)
				}
				if limit != 100 {
					t.Fatalf("expected limit 100, got %d", limit)
				}
				return []telegrambot.Update{
					{
						UpdateID: 5,
						Message: &telegrambot.Message{
							MessageID: 8,
							Text:      "ignore unmatched",
							Chat: telegrambot.Chat{
								ID:       -1001001,
								Username: "other",
								Title:    "Other",
							},
						},
					},
					{
						UpdateID: 6,
						ChannelPost: &telegrambot.Message{
							MessageID:       9,
							MessageThreadID: &threadID,
							Text:            "telegram source message",
							Chat: telegrambot.Chat{
								ID:       -1002451344189,
								Username: "coding",
								Title:    "Coding",
							},
						},
					},
					{
						UpdateID: 7,
						Message: &telegrambot.Message{
							MessageID:       10,
							MessageThreadID: &threadID,
							Text:            "   ",
							Chat: telegrambot.Chat{
								ID:       -1002451344189,
								Username: "coding",
								Title:    "Coding",
							},
						},
					},
				}, nil
			},
		},
		[]config.TelegramTarget{{ChatID: "-1002451344189", ThreadID: &threadID}},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("crawl: %v", err)
	}
	if len(enqueued) != 1 {
		t.Fatalf("expected 1 enqueued item, got %d", len(enqueued))
	}
	if enqueued[0].SourceID != "src-telegram" {
		t.Fatalf("expected telegram source id, got %s", enqueued[0].SourceID)
	}
	if enqueued[0].ExternalID != "tg:-1002451344189:5:9" {
		t.Fatalf("expected telegram external id, got %s", enqueued[0].ExternalID)
	}
	if enqueued[0].AuthorName != "Coding" {
		t.Fatalf("expected Coding author, got %s", enqueued[0].AuthorName)
	}
	if stored[telegramUpdateOffsetSettingKey] != "8" {
		t.Fatalf("expected stored offset 8, got %q", stored[telegramUpdateOffsetSettingKey])
	}
}

func TestTelegramIngestActionHandlesAdminCommandWithoutEnqueue(t *testing.T) {
	t.Parallel()

	threadID := int64(5)
	handledCommands := 0
	enqueued := 0

	action := NewTelegramIngestAction(
		telegramIngestSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-telegram", Type: sourcedomain.TypeTelegram, Handle: "-1002451344189", Name: "Coding"},
				}, nil
			},
		},
		telegramIngestSourceServiceStub{},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				enqueued++
				return contentdomain.ContentItem{}, nil
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) { return "", nil },
			setFn: func(ctx context.Context, key, value string) error { return nil },
		},
		telegramUpdatesClientStub{
			getUpdatesFn: func(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error) {
				return []telegrambot.Update{
					{
						UpdateID: 15,
						Message: &telegrambot.Message{
							MessageID:       41,
							MessageThreadID: &threadID,
							Text:            "/status",
							Chat:            telegrambot.Chat{ID: -1002451344189, Title: "Coding"},
						},
					},
				}, nil
			},
		},
		[]config.TelegramTarget{{ChatID: "-1002451344189", ThreadID: &threadID}},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	).WithAdminHandler(telegramAdminCommandHandlerStub{
		handleFn: func(ctx context.Context, message telegrambot.Message) (bool, error) {
			handledCommands++
			return true, nil
		},
	})

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("crawl with admin command: %v", err)
	}
	if handledCommands != 1 {
		t.Fatalf("expected 1 handled command, got %d", handledCommands)
	}
	if enqueued != 0 {
		t.Fatalf("expected admin command not enqueued, got %d enqueued", enqueued)
	}
}

func TestTelegramIngestActionIgnoresDuplicateContentErrors(t *testing.T) {
	t.Parallel()

	setCalls := 0
	threadID := int64(5)

	action := NewTelegramIngestAction(
		telegramIngestSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-telegram", Type: sourcedomain.TypeTelegram, Handle: "-1002451344189", Name: "Coding"},
				}, nil
			},
		},
		telegramIngestSourceServiceStub{},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				return contentdomain.ContentItem{}, contentdomain.ErrContentAlreadyExists
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				return "", nil
			},
			setFn: func(ctx context.Context, key, value string) error {
				setCalls++
				if value != "12" {
					t.Fatalf("expected offset 12, got %s", value)
				}
				return nil
			},
		},
		telegramUpdatesClientStub{
			getUpdatesFn: func(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error) {
				return []telegrambot.Update{
					{
						UpdateID: 11,
						Message: &telegrambot.Message{
							MessageID:       21,
							MessageThreadID: &threadID,
							Text:            "duplicate content",
							Chat: telegrambot.Chat{
								ID:    -1002451344189,
								Title: "Coding",
							},
						},
					},
				}, nil
			},
		},
		[]config.TelegramTarget{{ChatID: "-1002451344189", ThreadID: &threadID}},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("crawl with duplicate: %v", err)
	}
	if setCalls != 1 {
		t.Fatalf("expected settings offset update once, got %d", setCalls)
	}
}

type testWriter struct {
	t *testing.T
}

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}

func TestTelegramIngestActionPropagatesSettingsReadErrors(t *testing.T) {
	t.Parallel()
	threadID := int64(5)

	action := NewTelegramIngestAction(
		telegramIngestSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-telegram", Type: sourcedomain.TypeTelegram, Handle: "@coding", Name: "Coding"},
				}, nil
			},
		},
		telegramIngestSourceServiceStub{},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				return contentdomain.ContentItem{}, nil
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("settings unavailable")
			},
			setFn: func(ctx context.Context, key, value string) error {
				return nil
			},
		},
		telegramUpdatesClientStub{
			getUpdatesFn: func(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error) {
				return nil, nil
			},
		},
		[]config.TelegramTarget{{ChatID: "-1002451344189", ThreadID: &threadID}},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)

	if err := action.Crawl(context.Background()); err == nil {
		t.Fatal("expected settings read error")
	}
}

func TestTelegramIngestActionSkipsMessagesOutsideConfiguredTopic(t *testing.T) {
	t.Parallel()

	threadID := int64(5)
	enqueued := 0

	action := NewTelegramIngestAction(
		telegramIngestSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-telegram", Type: sourcedomain.TypeTelegram, Handle: "-1002451344189", Name: "Coding"},
				}, nil
			},
		},
		telegramIngestSourceServiceStub{},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				enqueued++
				return contentdomain.ContentItem{}, nil
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) { return "", nil },
			setFn: func(ctx context.Context, key, value string) error { return nil },
		},
		telegramUpdatesClientStub{
			getUpdatesFn: func(ctx context.Context, offset int64, limit int) ([]telegrambot.Update, error) {
				otherThreadID := int64(9)
				return []telegrambot.Update{
					{
						UpdateID: 14,
						Message: &telegrambot.Message{
							MessageID:       31,
							MessageThreadID: &otherThreadID,
							Text:            "wrong topic",
							Chat:            telegrambot.Chat{ID: -1002451344189, Title: "Coding"},
						},
					},
				}, nil
			},
		},
		[]config.TelegramTarget{{ChatID: "-1002451344189", ThreadID: &threadID}},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("crawl with non-matching topic: %v", err)
	}
	if enqueued != 0 {
		t.Fatalf("expected no enqueued messages outside configured topic, got %d", enqueued)
	}
}
