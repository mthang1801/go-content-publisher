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
	twitterclient "go-content-bot/internal/system/infrastructure/clients/twitter"
)

type twitterCrawlClientStub struct {
	lookupUserFn func(ctx context.Context, username string) (twitterclient.User, error)
	getTweetsFn  func(ctx context.Context, userID string, sinceID string, maxResults int) ([]twitterclient.Tweet, error)
}

func (s twitterCrawlClientStub) LookupUserByUsername(ctx context.Context, username string) (twitterclient.User, error) {
	return s.lookupUserFn(ctx, username)
}

func (s twitterCrawlClientStub) GetUserTweets(ctx context.Context, userID string, sinceID string, maxResults int) ([]twitterclient.Tweet, error) {
	return s.getTweetsFn(ctx, userID, sinceID, maxResults)
}

func TestTwitterCrawlActionIngestsNewTweetsAndStoresCheckpoint(t *testing.T) {
	t.Parallel()

	stored := map[string]string{
		"tw_uid_openai":   "2244994945",
		"tw_since_openai": "100",
	}
	enqueued := make([]contentapp.EnqueueSourceInput, 0)
	createdAt := time.Date(2026, 4, 18, 16, 0, 0, 0, time.UTC)

	action := NewTwitterCrawlAction(
		twitterSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-twitter", Type: sourcedomain.TypeTwitter, Handle: "@OpenAI", Name: "OpenAI"},
					{ID: "src-telegram", Type: sourcedomain.TypeTelegram, Handle: "@coding", Name: "Coding"},
				}, nil
			},
			touchCrawlFn: func(ctx context.Context, id string, at time.Time) error { return nil },
		},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				enqueued = append(enqueued, input)
				return contentdomain.ContentItem{ID: "item-" + input.ExternalID}, nil
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				return stored[key], nil
			},
			setFn: func(ctx context.Context, key, value string) error {
				stored[key] = value
				return nil
			},
		},
		twitterCrawlClientStub{
			lookupUserFn: func(ctx context.Context, username string) (twitterclient.User, error) {
				t.Fatal("lookup should not be called when user id is already cached")
				return twitterclient.User{}, nil
			},
			getTweetsFn: func(ctx context.Context, userID string, sinceID string, maxResults int) ([]twitterclient.Tweet, error) {
				if userID != "2244994945" {
					t.Fatalf("expected cached user id 2244994945, got %s", userID)
				}
				if sinceID != "100" {
					t.Fatalf("expected since id 100, got %s", sinceID)
				}
				if maxResults != 10 {
					t.Fatalf("expected max results 10, got %d", maxResults)
				}
				return []twitterclient.Tweet{
					{ID: "101", Text: "short"},
					{ID: "102", Text: "OpenAI ships a meaningful platform update for developers.", CreatedAt: &createdAt},
				}, nil
			},
		},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("twitter crawl: %v", err)
	}
	if len(enqueued) != 1 {
		t.Fatalf("expected 1 enqueued tweet, got %d", len(enqueued))
	}
	if enqueued[0].SourceID != "src-twitter" {
		t.Fatalf("expected twitter source id, got %s", enqueued[0].SourceID)
	}
	if enqueued[0].ExternalID != "102" {
		t.Fatalf("expected tweet id 102, got %s", enqueued[0].ExternalID)
	}
	if enqueued[0].AuthorName != "@openai" {
		t.Fatalf("expected author @openai, got %s", enqueued[0].AuthorName)
	}
	if enqueued[0].SourceURL == nil || *enqueued[0].SourceURL != "https://x.com/openai/status/102" {
		t.Fatalf("expected source url to be set, got %#v", enqueued[0].SourceURL)
	}
	if stored["tw_since_openai"] != "102" {
		t.Fatalf("expected latest since id 102, got %s", stored["tw_since_openai"])
	}
}

func TestTwitterCrawlActionInitializesCheckpointWithoutIngestingOldTweets(t *testing.T) {
	t.Parallel()

	stored := map[string]string{}
	action := NewTwitterCrawlAction(
		twitterSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-twitter", Type: sourcedomain.TypeTwitter, Handle: "@OpenAI", Name: "OpenAI"},
				}, nil
			},
			touchCrawlFn: func(ctx context.Context, id string, at time.Time) error { return nil },
		},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				t.Fatal("no tweets should be enqueued on first initialization crawl")
				return contentdomain.ContentItem{}, nil
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				return stored[key], nil
			},
			setFn: func(ctx context.Context, key, value string) error {
				stored[key] = value
				return nil
			},
		},
		twitterCrawlClientStub{
			lookupUserFn: func(ctx context.Context, username string) (twitterclient.User, error) {
				return twitterclient.User{ID: "2244994945", Username: "OpenAI"}, nil
			},
			getTweetsFn: func(ctx context.Context, userID string, sinceID string, maxResults int) ([]twitterclient.Tweet, error) {
				if sinceID != "" {
					t.Fatalf("expected empty since id on initialization, got %s", sinceID)
				}
				return []twitterclient.Tweet{
					{ID: "999", Text: "Most recent tweet"},
				}, nil
			},
		},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("twitter crawl initialization: %v", err)
	}
	if stored["tw_uid_openai"] != "2244994945" {
		t.Fatalf("expected cached user id, got %s", stored["tw_uid_openai"])
	}
	if stored["tw_since_openai"] != "999" {
		t.Fatalf("expected initialized since id 999, got %s", stored["tw_since_openai"])
	}
}

func TestTwitterCrawlActionMarksSourceInactiveWhenLookupReturnsNoUser(t *testing.T) {
	t.Parallel()

	marked := ""
	action := NewTwitterCrawlAction(
		twitterSourceServiceStub{
			listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
				return []sourcedomain.Source{
					{ID: "src-twitter", Type: sourcedomain.TypeTwitter, Handle: "@missing", Name: "Missing"},
				}, nil
			},
			touchCrawlFn: func(ctx context.Context, id string, at time.Time) error { return nil },
		},
		telegramIngestContentServiceStub{
			enqueueFromSourceFn: func(ctx context.Context, input contentapp.EnqueueSourceInput) (contentdomain.ContentItem, error) {
				t.Fatal("missing source should not enqueue content")
				return contentdomain.ContentItem{}, nil
			},
		},
		telegramSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) { return "", nil },
			setFn: func(ctx context.Context, key, value string) error { return nil },
		},
		twitterCrawlClientStub{
			lookupUserFn: func(ctx context.Context, username string) (twitterclient.User, error) {
				return twitterclient.User{}, errors.New("twitter user lookup returned no user")
			},
			getTweetsFn: func(ctx context.Context, userID string, sinceID string, maxResults int) ([]twitterclient.Tweet, error) {
				t.Fatal("timeline should not be called when lookup fails")
				return nil, nil
			},
		},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)
	action.sourceService = twitterSourceServiceStub{
		listActiveFn: func(ctx context.Context) ([]sourcedomain.Source, error) {
			return []sourcedomain.Source{{ID: "src-twitter", Type: sourcedomain.TypeTwitter, Handle: "@missing", Name: "Missing"}}, nil
		},
		touchCrawlFn: func(ctx context.Context, id string, at time.Time) error { return nil },
		markInactiveFn: func(ctx context.Context, id string, reason string, at time.Time) error {
			marked = id + "|" + reason
			return nil
		},
	}

	if err := action.Crawl(context.Background()); err != nil {
		t.Fatalf("crawl should swallow missing twitter source as inactive mark: %v", err)
	}
	if marked == "" {
		t.Fatal("expected source to be marked inactive")
	}
}

type twitterSourceServiceStub struct {
	listActiveFn   func(ctx context.Context) ([]sourcedomain.Source, error)
	touchCrawlFn   func(ctx context.Context, id string, at time.Time) error
	markInactiveFn func(ctx context.Context, id string, reason string, at time.Time) error
}

func (s twitterSourceServiceStub) ListActive(ctx context.Context) ([]sourcedomain.Source, error) {
	return s.listActiveFn(ctx)
}

func (s twitterSourceServiceStub) TouchCrawl(ctx context.Context, id string, at time.Time) error {
	if s.touchCrawlFn == nil {
		return nil
	}
	return s.touchCrawlFn(ctx, id, at)
}

func (s twitterSourceServiceStub) MarkInactive(ctx context.Context, id string, reason string, at time.Time) error {
	if s.markInactiveFn == nil {
		return nil
	}
	return s.markInactiveFn(ctx, id, reason, at)
}
