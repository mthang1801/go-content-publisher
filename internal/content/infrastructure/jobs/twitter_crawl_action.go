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
	twitterclient "go-content-bot/internal/system/infrastructure/clients/twitter"
)

type twitterUserLookupClient interface {
	LookupUserByUsername(ctx context.Context, username string) (twitterclient.User, error)
	GetUserTweets(ctx context.Context, userID string, sinceID string, maxResults int) ([]twitterclient.Tweet, error)
}

type twitterCrawlSourceService interface {
	ListActive(ctx context.Context) ([]sourcedomain.Source, error)
	TouchCrawl(ctx context.Context, id string, at time.Time) error
	MarkInactive(ctx context.Context, id string, reason string, at time.Time) error
}

type TwitterCrawlAction struct {
	sourceService twitterCrawlSourceService
	content       telegramContentEnqueuer
	settings      telegramSettingsStore
	client        twitterUserLookupClient
	logger        *slog.Logger
	enabled       bool
}

func NewTwitterCrawlAction(
	sourceService twitterCrawlSourceService,
	content telegramContentEnqueuer,
	settings telegramSettingsStore,
	client twitterUserLookupClient,
	logger *slog.Logger,
	enabled bool,
) *TwitterCrawlAction {
	return &TwitterCrawlAction{
		sourceService: sourceService,
		content:       content,
		settings:      settings,
		client:        client,
		logger:        logger,
		enabled:       enabled,
	}
}

func (a *TwitterCrawlAction) Name() string {
	return "twitter_crawl"
}

func (a *TwitterCrawlAction) Enabled() bool {
	return a.enabled
}

func (a *TwitterCrawlAction) Crawl(ctx context.Context) error {
	if !a.enabled {
		a.logger.Info("twitter crawl skipped", "reason", "twitter crawler disabled")
		return nil
	}

	sources, err := a.sourceService.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active sources: %w", err)
	}

	twitterSources := filterTwitterSources(sources)
	if len(twitterSources) == 0 {
		a.logger.Info("twitter crawl idle", "reason", "no active twitter sources")
		return nil
	}

	totalNew := 0
	for _, source := range twitterSources {
		count, err := a.crawlSource(ctx, source)
		if err != nil {
			return err
		}
		totalNew += count
	}

	a.logger.Info("twitter crawl completed", "ingested", totalNew, "sources", len(twitterSources))
	return nil
}

func (a *TwitterCrawlAction) crawlSource(ctx context.Context, source sourcedomain.Source) (int, error) {
	handle := normalizeTwitterHandle(source.Handle)
	if handle == "" {
		return 0, nil
	}
	now := time.Now()
	if err := a.sourceService.TouchCrawl(ctx, source.ID, now); err != nil {
		return 0, fmt.Errorf("touch twitter source crawl %s: %w", source.ID, err)
	}

	userIDKey := "tw_uid_" + handle
	userID, err := a.settings.Get(ctx, userIDKey)
	if err != nil {
		return 0, fmt.Errorf("read twitter user id cache for %s: %w", handle, err)
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		user, err := a.client.LookupUserByUsername(ctx, handle)
		if err != nil {
			if shouldDeactivateTwitterLookup(err) {
				if markErr := a.sourceService.MarkInactive(ctx, source.ID, err.Error(), now); markErr != nil {
					return 0, fmt.Errorf("mark twitter source inactive %s: %w", source.ID, markErr)
				}
				a.logger.Info("twitter source marked inactive", "source_id", source.ID, "handle", source.Handle, "reason", err.Error())
				return 0, nil
			}
			return 0, fmt.Errorf("lookup twitter user %s: %w", handle, err)
		}
		userID = user.ID
		if err := a.settings.Set(ctx, userIDKey, userID); err != nil {
			return 0, fmt.Errorf("cache twitter user id for %s: %w", handle, err)
		}
	}

	sinceIDKey := "tw_since_" + handle
	sinceID, err := a.settings.Get(ctx, sinceIDKey)
	if err != nil {
		return 0, fmt.Errorf("read twitter since id for %s: %w", handle, err)
	}
	sinceID = strings.TrimSpace(sinceID)

	if sinceID == "" {
		tweets, err := a.client.GetUserTweets(ctx, userID, "", 5)
		if err != nil {
			return 0, fmt.Errorf("initialize twitter timeline for %s: %w", handle, err)
		}
		if len(tweets) > 0 {
			if err := a.settings.Set(ctx, sinceIDKey, tweets[0].ID); err != nil {
				return 0, fmt.Errorf("store twitter since id for %s: %w", handle, err)
			}
		}
		return 0, nil
	}

	tweets, err := a.client.GetUserTweets(ctx, userID, sinceID, 10)
	if err != nil {
		return 0, fmt.Errorf("get twitter timeline for %s: %w", handle, err)
	}

	newCount := 0
	newestID := sinceID
	for _, tweet := range tweets {
		text := strings.TrimSpace(tweet.Text)
		if len(text) < 10 {
			continue
		}
		if compareTwitterIDs(tweet.ID, newestID) > 0 {
			newestID = tweet.ID
		}

		createdAt := time.Now()
		if tweet.CreatedAt != nil {
			createdAt = *tweet.CreatedAt
		}
		sourceURL := fmt.Sprintf("https://x.com/%s/status/%s", handle, tweet.ID)

		_, err := a.content.EnqueueFromSource(ctx, contentapp.EnqueueSourceInput{
			SourceID:     source.ID,
			ExternalID:   tweet.ID,
			OriginalText: text,
			AuthorName:   "@" + handle,
			SourceURL:    &sourceURL,
			CrawledAt:    createdAt,
		})
		if err != nil {
			if errors.Is(err, contentdomain.ErrContentAlreadyExists) {
				continue
			}
			return newCount, fmt.Errorf("enqueue twitter content from source %s: %w", source.ID, err)
		}
		newCount++
	}

	if newestID != "" && newestID != sinceID {
		if err := a.settings.Set(ctx, sinceIDKey, newestID); err != nil {
			return newCount, fmt.Errorf("store twitter since id for %s: %w", handle, err)
		}
	}

	return newCount, nil
}

func filterTwitterSources(sources []sourcedomain.Source) []sourcedomain.Source {
	filtered := make([]sourcedomain.Source, 0, len(sources))
	for _, source := range sources {
		if source.Type == sourcedomain.TypeTwitter {
			filtered = append(filtered, source)
		}
	}
	return filtered
}

func normalizeTwitterHandle(handle string) string {
	handle = strings.TrimSpace(strings.ToLower(handle))
	handle = strings.TrimPrefix(handle, "@")
	return handle
}

func compareTwitterIDs(left string, right string) int {
	if strings.TrimSpace(left) == "" && strings.TrimSpace(right) == "" {
		return 0
	}
	if strings.TrimSpace(right) == "" {
		return 1
	}
	leftID, leftErr := strconv.ParseInt(left, 10, 64)
	rightID, rightErr := strconv.ParseInt(right, 10, 64)
	if leftErr == nil && rightErr == nil {
		switch {
		case leftID > rightID:
			return 1
		case leftID < rightID:
			return -1
		default:
			return 0
		}
	}
	return strings.Compare(left, right)
}

func shouldDeactivateTwitterLookup(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "returned no user") || strings.Contains(message, "not found error")
}
