package jobs

import (
	"context"
	"log/slog"
	"testing"
	"time"

	contentapp "go-content-bot/internal/content/application"
	contentdomain "go-content-bot/internal/content/domain"
)

type twitterPublishRepoStub struct {
	findNextPublishedReadyForTwitterFn func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error)
	saveFn                             func(ctx context.Context, item contentdomain.ContentItem) error
}

func (s twitterPublishRepoStub) CreatePending(ctx context.Context, item contentdomain.ContentItem) (contentdomain.ContentItem, error) {
	return contentdomain.ContentItem{}, nil
}

func (s twitterPublishRepoStub) SkipStalePending(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return 0, nil
}

func (s twitterPublishRepoStub) SkipStaleRewritten(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return 0, nil
}

func (s twitterPublishRepoStub) ClaimNextPending(ctx context.Context) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s twitterPublishRepoStub) ClaimNextReadyForPublish(ctx context.Context) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s twitterPublishRepoStub) FindNextPublishedReadyForTwitter(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error) {
	return s.findNextPublishedReadyForTwitterFn(ctx, publishAfter, sourceTypes, sourceTags, sourceTopics, topicKeywords)
}

func (s twitterPublishRepoStub) FindNextPending(ctx context.Context) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s twitterPublishRepoStub) FindByID(ctx context.Context, id string) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s twitterPublishRepoStub) Save(ctx context.Context, item contentdomain.ContentItem) error {
	return s.saveFn(ctx, item)
}

func (s twitterPublishRepoStub) ListByStatuses(ctx context.Context, statuses []contentdomain.Status, limit int) ([]contentdomain.ContentItem, error) {
	return nil, nil
}

func (s twitterPublishRepoStub) ListRecent(ctx context.Context, limit int) ([]contentdomain.ContentItem, error) {
	return nil, nil
}

type twitterPublishClientStub struct {
	canPublishVI bool
	canPublishEN bool
	publishVIFn  func(ctx context.Context, text string) (string, error)
	publishENFn  func(ctx context.Context, text string) (string, error)
}

func (s twitterPublishClientStub) CanPublishVI() bool { return s.canPublishVI }
func (s twitterPublishClientStub) CanPublishEN() bool { return s.canPublishEN }

func (s twitterPublishClientStub) PublishTweetVI(ctx context.Context, text string) (string, error) {
	return s.publishVIFn(ctx, text)
}

func (s twitterPublishClientStub) PublishTweetEN(ctx context.Context, text string) (string, error) {
	return s.publishENFn(ctx, text)
}

type twitterSettingsStoreStub struct {
	getFn func(ctx context.Context, key string) (string, error)
	setFn func(ctx context.Context, key, value string) error
}

func (s twitterSettingsStoreStub) Get(ctx context.Context, key string) (string, error) {
	if s.getFn == nil {
		return "", nil
	}
	return s.getFn(ctx, key)
}

func (s twitterSettingsStoreStub) Set(ctx context.Context, key, value string) error {
	if s.setFn == nil {
		return nil
	}
	return s.setFn(ctx, key, value)
}

func TestTwitterPublishActionPublishesConfiguredAccounts(t *testing.T) {
	t.Parallel()

	saved := make([]contentdomain.ContentItem, 0)
	service := contentapp.NewService(twitterPublishRepoStub{
		findNextPublishedReadyForTwitterFn: func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error) {
			if publishAfter != nil {
				t.Fatalf("expected nil publish cutoff, got %v", publishAfter)
			}
			return &contentdomain.ContentItem{
				ID:              "item-twitter-publish",
				OriginalText:    "original",
				RewrittenText:   stringPtr("Ban tin *co link* https://example.com ve OpenAI."),
				RewrittenTextEn: stringPtr("English rewrite for OpenAI update."),
				TweetTextVI:     stringPtr("Ban tin cap nhat OpenAI danh cho nha phat trien @source https://example.com"),
				AuthorName:      "author",
				CrawledAt:       time.Now(),
				Status:          contentdomain.StatusPublished,
			}, nil
		},
		saveFn: func(ctx context.Context, item contentdomain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	})

	action := NewTwitterPublishAction(
		service,
		twitterPublishClientStub{
			canPublishVI: true,
			canPublishEN: true,
			publishVIFn: func(ctx context.Context, text string) (string, error) {
				if text != "Ban tin cap nhat OpenAI danh cho nha phat trien" {
					t.Fatalf("expected cleaned vietnamese tweet text, got %q", text)
				}
				return "tweet-vi-1", nil
			},
			publishENFn: func(ctx context.Context, text string) (string, error) {
				if text != "English rewrite for OpenAI update." {
					t.Fatalf("expected english tweet text, got %q", text)
				}
				return "tweet-en-1", nil
			},
		},
		twitterSettingsStoreStub{},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
		nil,
	)

	if err := action.Publish(context.Background()); err != nil {
		t.Fatalf("twitter publish action: %v", err)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 saved item, got %d", len(saved))
	}
	if saved[0].TweetViID == nil || *saved[0].TweetViID != "tweet-vi-1" {
		t.Fatalf("expected tweet vi id to be saved, got %#v", saved[0].TweetViID)
	}
	if saved[0].TweetEnID == nil || *saved[0].TweetEnID != "tweet-en-1" {
		t.Fatalf("expected tweet en id to be saved, got %#v", saved[0].TweetEnID)
	}
}

func TestTwitterPublishActionSkipsWhenTextTooShortOrAccountDisabled(t *testing.T) {
	t.Parallel()

	service := contentapp.NewService(twitterPublishRepoStub{
		findNextPublishedReadyForTwitterFn: func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error) {
			if publishAfter != nil {
				t.Fatalf("expected nil publish cutoff, got %v", publishAfter)
			}
			return &contentdomain.ContentItem{
				ID:            "item-twitter-skip",
				OriginalText:  "original",
				RewrittenText: stringPtr("Ngan"),
				AuthorName:    "author",
				CrawledAt:     time.Now(),
				Status:        contentdomain.StatusPublished,
			}, nil
		},
		saveFn: func(ctx context.Context, item contentdomain.ContentItem) error { return nil },
	})

	action := NewTwitterPublishAction(
		service,
		twitterPublishClientStub{
			canPublishVI: false,
			canPublishEN: true,
			publishVIFn: func(ctx context.Context, text string) (string, error) {
				t.Fatal("vi publish should not be called when account is disabled")
				return "", nil
			},
			publishENFn: func(ctx context.Context, text string) (string, error) {
				t.Fatal("en publish should not be called when text is too short")
				return "", nil
			},
		},
		twitterSettingsStoreStub{},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
		nil,
	)

	if err := action.Publish(context.Background()); err != nil {
		t.Fatalf("twitter publish action skip: %v", err)
	}
}

func TestTwitterPublishActionUsesSettingOverride(t *testing.T) {
	t.Parallel()

	expected := time.Date(2026, 4, 19, 6, 3, 28, 0, time.FixedZone("+07", 7*60*60))
	service := contentapp.NewService(twitterPublishRepoStub{
		findNextPublishedReadyForTwitterFn: func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error) {
			if publishAfter == nil || !publishAfter.Equal(expected) {
				t.Fatalf("expected publish cutoff %v, got %v", expected, publishAfter)
			}
			if len(sourceTypes) != 2 || sourceTypes[0] != "twitter" || sourceTypes[1] != "telegram" {
				t.Fatalf("unexpected source types %#v", sourceTypes)
			}
			if len(sourceTopics) != 2 || sourceTopics[0] != "economy" || sourceTopics[1] != "crypto" {
				t.Fatalf("unexpected source topics %#v", sourceTopics)
			}
			if len(sourceTags) != 0 {
				t.Fatalf("expected source tags to be ignored when topics exist, got %#v", sourceTags)
			}
			if len(topicKeywords) != 2 || topicKeywords[0] != "macro" || topicKeywords[1] != "fed" {
				t.Fatalf("unexpected topic keywords %#v", topicKeywords)
			}
			return nil, nil
		},
		saveFn: func(ctx context.Context, item contentdomain.ContentItem) error { return nil },
	})

	action := NewTwitterPublishAction(
		service,
		twitterPublishClientStub{},
		twitterSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				switch key {
				case twitterPublishAfterSettingKey:
					return expected.Format(time.RFC3339), nil
				case twitterPublishSourceTypesSettingKey:
					return "twitter,telegram", nil
				case twitterPublishSourceTagsSettingKey:
					return "markets,macro", nil
				case twitterPublishSourceTopicsSettingKey:
					return "economy,crypto", nil
				case twitterPublishTopicKeywordsSettingKey:
					return "macro,fed", nil
				default:
					t.Fatalf("unexpected setting key %q", key)
					return "", nil
				}
			},
		},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
		nil,
	)

	if err := action.Publish(context.Background()); err != nil {
		t.Fatalf("twitter publish action with setting override: %v", err)
	}
}

func TestTwitterPublishActionBootstrapsSettingFromDefault(t *testing.T) {
	t.Parallel()

	defaultCutoff := time.Date(2026, 4, 19, 6, 3, 28, 0, time.FixedZone("+07", 7*60*60))
	var persisted string
	service := contentapp.NewService(twitterPublishRepoStub{
		findNextPublishedReadyForTwitterFn: func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error) {
			if publishAfter == nil || !publishAfter.Equal(defaultCutoff) {
				t.Fatalf("expected publish cutoff %v, got %v", defaultCutoff, publishAfter)
			}
			return nil, nil
		},
		saveFn: func(ctx context.Context, item contentdomain.ContentItem) error { return nil },
	})

	action := NewTwitterPublishAction(
		service,
		twitterPublishClientStub{},
		twitterSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				return "", nil
			},
			setFn: func(ctx context.Context, key, value string) error {
				if key != twitterPublishAfterSettingKey {
					t.Fatalf("unexpected setting key %q", key)
				}
				persisted = value
				return nil
			},
		},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
		&defaultCutoff,
	)

	if err := action.Publish(context.Background()); err != nil {
		t.Fatalf("twitter publish action bootstrap setting: %v", err)
	}
	if persisted != defaultCutoff.Format(time.RFC3339) {
		t.Fatalf("expected persisted cutoff %q, got %q", defaultCutoff.Format(time.RFC3339), persisted)
	}
}

func TestTwitterPublishActionUsesSourceTagsWhenTopicsEmpty(t *testing.T) {
	t.Parallel()

	service := contentapp.NewService(twitterPublishRepoStub{
		findNextPublishedReadyForTwitterFn: func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error) {
			if len(sourceTopics) != 0 {
				t.Fatalf("expected no source topics, got %#v", sourceTopics)
			}
			if len(sourceTags) != 2 || sourceTags[0] != "markets" || sourceTags[1] != "macro" {
				t.Fatalf("unexpected source tags %#v", sourceTags)
			}
			return nil, nil
		},
		saveFn: func(ctx context.Context, item contentdomain.ContentItem) error { return nil },
	})

	action := NewTwitterPublishAction(
		service,
		twitterPublishClientStub{},
		twitterSettingsStoreStub{
			getFn: func(ctx context.Context, key string) (string, error) {
				switch key {
				case twitterPublishAfterSettingKey:
					return "", nil
				case twitterPublishSourceTypesSettingKey:
					return "twitter", nil
				case twitterPublishSourceTagsSettingKey:
					return "markets,macro", nil
				case twitterPublishSourceTopicsSettingKey:
					return "", nil
				case twitterPublishTopicKeywordsSettingKey:
					return "", nil
				default:
					t.Fatalf("unexpected setting key %q", key)
					return "", nil
				}
			},
		},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
		nil,
	)

	if err := action.Publish(context.Background()); err != nil {
		t.Fatalf("twitter publish action with source tag fallback: %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
