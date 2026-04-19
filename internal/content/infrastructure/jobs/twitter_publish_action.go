package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	contentapp "go-content-bot/internal/content/application"
)

type twitterPublishClient interface {
	CanPublishVI() bool
	CanPublishEN() bool
	PublishTweetVI(ctx context.Context, text string) (string, error)
	PublishTweetEN(ctx context.Context, text string) (string, error)
}

type twitterSettingsStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

const twitterPublishAfterSettingKey = "twitter_publish_after"
const twitterPublishSourceTypesSettingKey = "twitter_publish_source_types"
const twitterPublishSourceTagsSettingKey = "twitter_publish_source_tags"
const twitterPublishSourceTopicsSettingKey = "twitter_publish_source_topics"
const twitterPublishTopicKeywordsSettingKey = "twitter_publish_topic_keywords"

type TwitterPublishAction struct {
	service      *contentapp.Service
	client       twitterPublishClient
	settings     twitterSettingsStore
	logger       *slog.Logger
	enabled      bool
	publishAfter *time.Time
}

func NewTwitterPublishAction(service *contentapp.Service, client twitterPublishClient, settings twitterSettingsStore, logger *slog.Logger, enabled bool, publishAfter *time.Time) *TwitterPublishAction {
	return &TwitterPublishAction{
		service:      service,
		client:       client,
		settings:     settings,
		logger:       logger,
		enabled:      enabled,
		publishAfter: publishAfter,
	}
}

func (a *TwitterPublishAction) Name() string {
	return "twitter_publish"
}

func (a *TwitterPublishAction) Enabled() bool {
	return a.enabled
}

func (a *TwitterPublishAction) Publish(ctx context.Context) error {
	if !a.enabled {
		a.logger.Info("twitter publish skipped", "reason", "twitter publish disabled")
		return nil
	}

	publishAfter, sourceTypes, sourceTags, sourceTopics, topicKeywords, err := a.resolvePolicy(ctx)
	if err != nil {
		return fmt.Errorf("resolve twitter publish policy: %w", err)
	}

	item, err := a.service.PublishNextToTwitterWithFilters(ctx, a, publishAfter, sourceTypes, sourceTags, sourceTopics, topicKeywords)
	if err != nil {
		return fmt.Errorf("publish next content item to twitter: %w", err)
	}
	if item == nil {
		a.logger.Info("twitter publish idle", "reason", "no published items waiting for twitter")
		return nil
	}

	a.logger.Info("twitter publish completed",
		"content_id", item.ID,
		"tweet_vi_id", derefTwitterResult(item.TweetViID),
		"tweet_en_id", derefTwitterResult(item.TweetEnID),
	)
	return nil
}

func (a *TwitterPublishAction) resolvePolicy(ctx context.Context) (*time.Time, []string, []string, []string, []string, error) {
	sourceTypes := []string{}
	sourceTags := []string{}
	sourceTopics := []string{}
	topicKeywords := []string{}
	if a.settings != nil {
		value, err := a.settings.Get(ctx, twitterPublishAfterSettingKey)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		value = strings.TrimSpace(value)
		if value != "" {
			parsed, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("parse %s: %w", twitterPublishAfterSettingKey, err)
			}
			if sourceTypes, err = a.resolveStringListSetting(ctx, twitterPublishSourceTypesSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if sourceTopics, err = a.resolveStringListSetting(ctx, twitterPublishSourceTopicsSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if sourceTags, err = a.resolveStringListSetting(ctx, twitterPublishSourceTagsSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if topicKeywords, err = a.resolveStringListSetting(ctx, twitterPublishTopicKeywordsSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			sourceTags, sourceTopics = applySourcePolicyPrecedence(sourceTags, sourceTopics)
			return &parsed, sourceTypes, sourceTags, sourceTopics, topicKeywords, nil
		}
	}

	if a.publishAfter == nil {
		var err error
		if a.settings != nil {
			if sourceTypes, err = a.resolveStringListSetting(ctx, twitterPublishSourceTypesSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if sourceTags, err = a.resolveStringListSetting(ctx, twitterPublishSourceTagsSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if sourceTopics, err = a.resolveStringListSetting(ctx, twitterPublishSourceTopicsSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if topicKeywords, err = a.resolveStringListSetting(ctx, twitterPublishTopicKeywordsSettingKey); err != nil {
				return nil, nil, nil, nil, nil, err
			}
		}
		sourceTags, sourceTopics = applySourcePolicyPrecedence(sourceTags, sourceTopics)
		return nil, sourceTypes, sourceTags, sourceTopics, topicKeywords, nil
	}
	if a.settings != nil {
		if err := a.settings.Set(ctx, twitterPublishAfterSettingKey, a.publishAfter.Format(time.RFC3339)); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		var err error
		if sourceTypes, err = a.resolveStringListSetting(ctx, twitterPublishSourceTypesSettingKey); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if sourceTags, err = a.resolveStringListSetting(ctx, twitterPublishSourceTagsSettingKey); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if sourceTopics, err = a.resolveStringListSetting(ctx, twitterPublishSourceTopicsSettingKey); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if topicKeywords, err = a.resolveStringListSetting(ctx, twitterPublishTopicKeywordsSettingKey); err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}
	sourceTags, sourceTopics = applySourcePolicyPrecedence(sourceTags, sourceTopics)
	return a.publishAfter, sourceTypes, sourceTags, sourceTopics, topicKeywords, nil
}

func (a *TwitterPublishAction) resolveStringListSetting(ctx context.Context, key string) ([]string, error) {
	if a.settings == nil {
		return nil, nil
	}
	value, err := a.settings.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return parseStringList(value), nil
}

func parseStringList(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	return values
}

func applySourcePolicyPrecedence(sourceTags []string, sourceTopics []string) ([]string, []string) {
	if len(sourceTopics) > 0 {
		return nil, sourceTopics
	}
	return sourceTags, nil
}

func (a *TwitterPublishAction) PublishTweets(ctx context.Context, input contentapp.TwitterPublishInput) (contentapp.TwitterPublishResult, error) {
	result := contentapp.TwitterPublishResult{
		TweetViID: "skipped",
		TweetEnID: "skipped",
	}

	if a.client.CanPublishVI() {
		if text := prepareTweetText(input.TweetTextVI); text != "" {
			tweetID, err := a.client.PublishTweetVI(ctx, text)
			if err != nil {
				return contentapp.TwitterPublishResult{}, err
			}
			result.TweetViID = tweetID
		}
	}

	if a.client.CanPublishEN() {
		if text := prepareTweetText(input.TweetTextEN); text != "" {
			tweetID, err := a.client.PublishTweetEN(ctx, text)
			if err != nil {
				return contentapp.TwitterPublishResult{}, err
			}
			result.TweetEnID = tweetID
		}
	}

	return result, nil
}

var markdownLinkPattern = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
var urlPattern = regexp.MustCompile(`https?://[^\s)\]>]+`)
var invitePattern = regexp.MustCompile(`(?i)(t\.me|bit\.ly|discord\.gg|tinyurl\.com|goo\.gl)/[^\s)\]>]+`)
var handlePattern = regexp.MustCompile(`@\w+`)

func prepareTweetText(text *string) string {
	if text == nil {
		return ""
	}

	cleaned := strings.TrimSpace(*text)
	if cleaned == "" {
		return ""
	}

	cleaned = markdownLinkPattern.ReplaceAllString(cleaned, "$1")
	cleaned = urlPattern.ReplaceAllString(cleaned, "")
	cleaned = invitePattern.ReplaceAllString(cleaned, "")
	cleaned = handlePattern.ReplaceAllString(cleaned, "")
	cleaned = strings.ReplaceAll(cleaned, "*", "")
	cleaned = strings.ReplaceAll(cleaned, "_", "")
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	if len(cleaned) < 20 {
		return ""
	}
	return truncateForTweet(cleaned, 275)
}

func truncateForTweet(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	truncated := text[:maxLen]
	lastSentence := strings.LastIndex(truncated, ".")
	lastNewline := strings.LastIndex(truncated, "\n")
	cutPoint := lastSentence
	if lastNewline > cutPoint {
		cutPoint = lastNewline
	}
	if cutPoint > maxLen/2 {
		return strings.TrimSpace(truncated[:cutPoint+1]) + "…"
	}
	return strings.TrimSpace(truncated) + "…"
}

func derefTwitterResult(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
