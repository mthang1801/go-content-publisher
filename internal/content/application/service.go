package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"go-content-bot/internal/content/domain"
)

type Service struct {
	repo domain.Repository
}

const (
	publishDedupRecentLimit   = 100
	publishDedupRecentWindow  = 24 * time.Hour
	rewriteDedupRecentLimit   = 200
	rewriteDedupRecentWindow  = 48 * time.Hour
	rewriteOriginalThreshold  = 0.70
	rewriteRewrittenThreshold = 0.75
	adminRetryFailedLimit     = 100
	adminSkipSearchLimit      = 500
)

var vietnameseNormalizer = strings.NewReplacer(
	"à", "a", "á", "a", "ạ", "a", "ả", "a", "ã", "a",
	"â", "a", "ầ", "a", "ấ", "a", "ậ", "a", "ẩ", "a", "ẫ", "a",
	"ă", "a", "ằ", "a", "ắ", "a", "ặ", "a", "ẳ", "a", "ẵ", "a",
	"è", "e", "é", "e", "ẹ", "e", "ẻ", "e", "ẽ", "e",
	"ê", "e", "ề", "e", "ế", "e", "ệ", "e", "ể", "e", "ễ", "e",
	"ì", "i", "í", "i", "ị", "i", "ỉ", "i", "ĩ", "i",
	"ò", "o", "ó", "o", "ọ", "o", "ỏ", "o", "õ", "o",
	"ô", "o", "ồ", "o", "ố", "o", "ộ", "o", "ổ", "o", "ỗ", "o",
	"ơ", "o", "ờ", "o", "ớ", "o", "ợ", "o", "ở", "o", "ỡ", "o",
	"ù", "u", "ú", "u", "ụ", "u", "ủ", "u", "ũ", "u",
	"ư", "u", "ừ", "u", "ứ", "u", "ự", "u", "ử", "u", "ữ", "u",
	"ỳ", "y", "ý", "y", "ỵ", "y", "ỷ", "y", "ỹ", "y",
	"đ", "d",
)

type EnqueueManualInput struct {
	Text   string
	Author string
}

type EnqueueSourceInput struct {
	SourceID     string
	ExternalID   string
	OriginalText string
	AuthorName   string
	SourceURL    *string
	CrawledAt    time.Time
}

type RewriteOptions struct {
	DuplicateWindow             time.Duration
	DuplicateOriginalThreshold  float64
	DuplicateRewrittenThreshold float64
	DuplicateRecentLimit        int
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SkipStalePendingBefore(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return s.repo.SkipStalePending(ctx, staleBefore, reason)
}

func (s *Service) SkipStaleRewrittenBefore(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return s.repo.SkipStaleRewritten(ctx, staleBefore, reason)
}

func (s *Service) EnqueueManual(ctx context.Context, input EnqueueManualInput) (domain.ContentItem, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return domain.ContentItem{}, errors.New("manual content text is required")
	}

	author := strings.TrimSpace(input.Author)
	if author == "" {
		author = "Manual"
	}

	item := domain.ContentItem{
		OriginalText: text,
		AuthorName:   author,
		CrawledAt:    time.Now(),
		Status:       domain.StatusPending,
	}

	if err := item.Validate(); err != nil {
		return domain.ContentItem{}, fmt.Errorf("validate manual content item: %w", err)
	}

	return s.repo.CreatePending(ctx, item)
}

func (s *Service) EnqueueFromSource(ctx context.Context, input EnqueueSourceInput) (domain.ContentItem, error) {
	sourceID := strings.TrimSpace(input.SourceID)
	if sourceID == "" {
		return domain.ContentItem{}, errors.New("source id is required")
	}

	text := strings.TrimSpace(input.OriginalText)
	if text == "" {
		return domain.ContentItem{}, errors.New("source content text is required")
	}

	author := strings.TrimSpace(input.AuthorName)
	if author == "" {
		author = "Telegram"
	}

	externalID := strings.TrimSpace(input.ExternalID)
	if externalID == "" {
		return domain.ContentItem{}, errors.New("external id is required")
	}

	crawledAt := input.CrawledAt
	if crawledAt.IsZero() {
		crawledAt = time.Now()
	}

	item := domain.ContentItem{
		SourceID:     &sourceID,
		ExternalID:   externalID,
		OriginalText: text,
		AuthorName:   author,
		SourceURL:    input.SourceURL,
		CrawledAt:    crawledAt,
		Status:       domain.StatusPending,
	}

	if err := item.Validate(); err != nil {
		return domain.ContentItem{}, fmt.Errorf("validate source content item: %w", err)
	}

	return s.repo.CreatePending(ctx, item)
}

func (s *Service) SetManualRewrite(ctx context.Context, id, text string) (domain.ContentItem, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.ContentItem{}, err
	}
	if item == nil {
		return domain.ContentItem{}, errors.New("content item not found")
	}
	if err := item.SetManualRewrite(strings.TrimSpace(text)); err != nil {
		return domain.ContentItem{}, err
	}
	if err := s.repo.Save(ctx, *item); err != nil {
		return domain.ContentItem{}, err
	}
	return *item, nil
}

func (s *Service) ListQueue(ctx context.Context, limit int) ([]domain.ContentItem, error) {
	return s.repo.ListByStatuses(ctx, []domain.Status{
		domain.StatusPending,
		domain.StatusProcessing,
		domain.StatusRewritten,
		domain.StatusPublishing,
		domain.StatusFailed,
	}, limit)
}

func (s *Service) ListRecent(ctx context.Context, limit int) ([]domain.ContentItem, error) {
	return s.repo.ListRecent(ctx, limit)
}

func (s *Service) RetryFailed(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = adminRetryFailedLimit
	}
	items, err := s.repo.ListByStatuses(ctx, []domain.Status{domain.StatusFailed}, limit)
	if err != nil {
		return 0, err
	}

	retried := 0
	for _, item := range items {
		if err := item.RetryFailed(); err != nil {
			return retried, err
		}
		if err := s.repo.Save(ctx, item); err != nil {
			return retried, err
		}
		retried++
	}
	return retried, nil
}

func (s *Service) SkipByIDPrefix(ctx context.Context, prefix, reason string) (domain.ContentItem, error) {
	prefix = strings.TrimSpace(prefix)
	if len(prefix) < 4 {
		return domain.ContentItem{}, errors.New("content id prefix must be at least 4 characters")
	}

	items, err := s.repo.ListRecent(ctx, adminSkipSearchLimit)
	if err != nil {
		return domain.ContentItem{}, err
	}

	var matched []domain.ContentItem
	for _, item := range items {
		if strings.HasPrefix(item.ID, prefix) {
			matched = append(matched, item)
			if len(matched) > 1 {
				return domain.ContentItem{}, errors.New("content id prefix is ambiguous")
			}
		}
	}
	if len(matched) == 0 {
		return domain.ContentItem{}, errors.New("content item not found")
	}

	item := matched[0]
	if strings.TrimSpace(reason) == "" {
		reason = "skipped by Telegram admin command"
	}
	if err := item.MarkSkipped(reason); err != nil {
		return domain.ContentItem{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return domain.ContentItem{}, err
	}
	return item, nil
}

func (s *Service) ProcessNextPending(ctx context.Context, rewriter RewriteTextPort) (*domain.ContentItem, error) {
	return s.ProcessNextPendingWithOptions(ctx, rewriter, RewriteOptions{})
}

func (s *Service) ProcessNextPendingWithOptions(ctx context.Context, rewriter RewriteTextPort, options RewriteOptions) (*domain.ContentItem, error) {
	item, err := s.repo.ClaimNextPending(ctx)
	if err != nil || item == nil {
		return item, err
	}

	resolvedOptions := resolveRewriteOptions(options)
	if reason, err := s.findPreRewriteDuplicate(ctx, *item, resolvedOptions); err != nil {
		return nil, err
	} else if reason != "" {
		if err := item.MarkSkipped(reason); err != nil {
			return nil, err
		}
		if err := s.repo.Save(ctx, *item); err != nil {
			return nil, err
		}
		return item, nil
	}

	result, err := rewriter.RewriteText(ctx, item.OriginalText)
	if err != nil {
		if markErr := item.MarkFailed(err.Error()); markErr != nil {
			return nil, markErr
		}
		if saveErr := s.repo.Save(ctx, *item); saveErr != nil {
			return nil, saveErr
		}
		return nil, err
	}

	if !result.ShouldPublish {
		applyRewriteFields(item, result)
		reason := strings.TrimSpace(result.Reason)
		if reason == "" {
			reason = "rewrite result marked should_publish=false"
		}
		if err := item.MarkSkipped(reason); err != nil {
			return nil, err
		}
		if err := s.repo.Save(ctx, *item); err != nil {
			return nil, err
		}
		return item, nil
	}

	if err := item.MarkRewritten(strings.TrimSpace(result.RewrittenText)); err != nil {
		return nil, err
	}
	applyRewriteFields(item, result)

	if reason, err := s.findPostRewriteDuplicate(ctx, *item, resolvedOptions); err != nil {
		return nil, err
	} else if reason != "" {
		if err := item.MarkSkipped(reason); err != nil {
			return nil, err
		}
		if err := s.repo.Save(ctx, *item); err != nil {
			return nil, err
		}
		return item, nil
	}
	if err := s.repo.Save(ctx, *item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) findPreRewriteDuplicate(ctx context.Context, item domain.ContentItem, options RewriteOptions) (string, error) {
	recent, err := s.repo.ListByStatuses(ctx, []domain.Status{
		domain.StatusProcessing,
		domain.StatusRewritten,
		domain.StatusPublished,
		domain.StatusSkipped,
	}, options.DuplicateRecentLimit)
	if err != nil {
		return "", err
	}
	cutoff := time.Now().Add(-options.DuplicateWindow)
	for _, existing := range recent {
		if existing.ID == item.ID || existing.CrawledAt.Before(cutoff) {
			continue
		}
		if item.IsManual() {
			if matchesNormalizedDuplicate(item, existing) {
				return "duplicate manual content already processed recently", nil
			}
			continue
		}
		if similarity(item.OriginalText, existing.OriginalText) >= options.DuplicateOriginalThreshold {
			return fmt.Sprintf("duplicate original content within %.0f%% similarity", options.DuplicateOriginalThreshold*100), nil
		}
		if existing.RewrittenText != nil && similarity(item.OriginalText, *existing.RewrittenText) >= options.DuplicateOriginalThreshold {
			return fmt.Sprintf("duplicate rewritten content within %.0f%% similarity", options.DuplicateOriginalThreshold*100), nil
		}
	}
	return "", nil
}

func (s *Service) findPostRewriteDuplicate(ctx context.Context, item domain.ContentItem, options RewriteOptions) (string, error) {
	if item.RewrittenText == nil {
		return "", nil
	}
	recent, err := s.repo.ListByStatuses(ctx, []domain.Status{
		domain.StatusRewritten,
		domain.StatusPublished,
	}, options.DuplicateRecentLimit)
	if err != nil {
		return "", err
	}
	cutoff := time.Now().Add(-options.DuplicateWindow)
	for _, existing := range recent {
		if existing.ID == item.ID || existing.CrawledAt.Before(cutoff) || existing.RewrittenText == nil {
			continue
		}
		if item.IsManual() {
			if matchesNormalizedDuplicate(item, existing) {
				return "duplicate manual rewritten content already processed recently", nil
			}
			continue
		}
		if similarity(*item.RewrittenText, *existing.RewrittenText) >= options.DuplicateRewrittenThreshold {
			return fmt.Sprintf("duplicate rewritten content within %.0f%% similarity", options.DuplicateRewrittenThreshold*100), nil
		}
	}
	return "", nil
}

func resolveRewriteOptions(options RewriteOptions) RewriteOptions {
	if options.DuplicateWindow <= 0 {
		options.DuplicateWindow = rewriteDedupRecentWindow
	}
	if options.DuplicateOriginalThreshold <= 0 {
		options.DuplicateOriginalThreshold = rewriteOriginalThreshold
	}
	if options.DuplicateRewrittenThreshold <= 0 {
		options.DuplicateRewrittenThreshold = rewriteRewrittenThreshold
	}
	if options.DuplicateRecentLimit <= 0 {
		options.DuplicateRecentLimit = rewriteDedupRecentLimit
	}
	return options
}

func applyRewriteFields(item *domain.ContentItem, result RewriteResult) {
	if item == nil {
		return
	}
	if text := strings.TrimSpace(result.RewrittenText); text != "" {
		item.RewrittenText = stringPtr(text)
	}
	if text := strings.TrimSpace(result.RewrittenTextEn); text != "" {
		item.RewrittenTextEn = stringPtr(text)
	}
	if text := strings.TrimSpace(result.TweetTextVI); text != "" {
		item.TweetTextVI = stringPtr(text)
	}
	if text := strings.TrimSpace(result.TweetTextEN); text != "" {
		item.TweetTextEN = stringPtr(text)
	}
	if text := strings.TrimSpace(result.FactCheckNote); text != "" {
		item.FactCheckNote = stringPtr(text)
	}
}

func (s *Service) PublishNextReady(ctx context.Context, publisher PublishTextPort) (*domain.ContentItem, error) {
	item, err := s.repo.ClaimNextReadyForPublish(ctx)
	if err != nil || item == nil {
		return item, err
	}
	text := item.OriginalText
	if item.RewrittenText != nil && strings.TrimSpace(*item.RewrittenText) != "" {
		text = *item.RewrittenText
	}

	recentPublished, err := s.repo.ListByStatuses(ctx, []domain.Status{domain.StatusPublished}, publishDedupRecentLimit)
	if err != nil {
		return nil, err
	}
	if matchesRecentlyPublishedDuplicate(*item, recentPublished, time.Now()) {
		if err := item.MarkSkipped("duplicate of recently published content"); err != nil {
			return nil, err
		}
		if err := s.repo.Save(ctx, *item); err != nil {
			return nil, err
		}
		return item, nil
	}

	messageID, err := publisher.PublishText(ctx, text)
	if err != nil {
		if markErr := item.MarkFailed(err.Error()); markErr != nil {
			return nil, markErr
		}
		if saveErr := s.repo.Save(ctx, *item); saveErr != nil {
			return nil, saveErr
		}
		return nil, err
	}

	if err := item.MarkPublished(messageID, time.Now()); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, *item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) PublishNextToTwitter(ctx context.Context, publisher PublishTweetsPort, publishAfter *time.Time) (*domain.ContentItem, error) {
	return s.PublishNextToTwitterWithFilters(ctx, publisher, publishAfter, nil, nil, nil, nil)
}

func (s *Service) PublishNextToTwitterWithFilters(ctx context.Context, publisher PublishTweetsPort, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*domain.ContentItem, error) {
	item, err := s.repo.FindNextPublishedReadyForTwitter(ctx, publishAfter, sourceTypes, sourceTags, sourceTopics, topicKeywords)
	if err != nil || item == nil {
		return item, err
	}

	result, err := publisher.PublishTweets(ctx, TwitterPublishInput{
		TweetTextVI: chooseTwitterTextVI(item),
		TweetTextEN: chooseTwitterTextEN(item),
	})
	if err != nil {
		return nil, err
	}

	if err := item.SetTwitterPublishResults(result.TweetViID, result.TweetEnID); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, *item); err != nil {
		return nil, err
	}
	return item, nil
}

func chooseTwitterTextVI(item *domain.ContentItem) *string {
	if item == nil {
		return nil
	}
	if item.TweetTextVI != nil && strings.TrimSpace(*item.TweetTextVI) != "" {
		return item.TweetTextVI
	}
	if item.RewrittenText != nil && strings.TrimSpace(*item.RewrittenText) != "" {
		return item.RewrittenText
	}
	return nil
}

func chooseTwitterTextEN(item *domain.ContentItem) *string {
	if item == nil {
		return nil
	}
	if item.TweetTextEN != nil && strings.TrimSpace(*item.TweetTextEN) != "" {
		return item.TweetTextEN
	}
	if item.RewrittenTextEn != nil && strings.TrimSpace(*item.RewrittenTextEn) != "" {
		return item.RewrittenTextEn
	}
	return nil
}

func matchesRecentlyPublishedDuplicate(candidate domain.ContentItem, published []domain.ContentItem, now time.Time) bool {
	normalizedCandidates := normalizedContentVariants(candidate)
	if len(normalizedCandidates) == 0 {
		return false
	}

	cutoff := now.Add(-publishDedupRecentWindow)
	for _, item := range published {
		if item.PublishedAt == nil || item.PublishedAt.Before(cutoff) {
			continue
		}

		for publishedText := range normalizedContentVariants(item) {
			if _, ok := normalizedCandidates[publishedText]; ok {
				return true
			}
		}
	}
	return false
}

func normalizedContentVariants(item domain.ContentItem) map[string]struct{} {
	normalized := make(map[string]struct{}, 2)
	for _, value := range []string{item.OriginalText, derefString(item.RewrittenText)} {
		if text := normalizePublishText(value); text != "" {
			normalized[text] = struct{}{}
		}
	}
	return normalized
}

func matchesNormalizedDuplicate(candidate domain.ContentItem, existing domain.ContentItem) bool {
	candidateVariants := normalizedContentVariants(candidate)
	if len(candidateVariants) == 0 {
		return false
	}
	for existingText := range normalizedContentVariants(existing) {
		if _, ok := candidateVariants[existingText]; ok {
			return true
		}
	}
	return false
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPtr(value string) *string {
	return &value
}

func normalizePublishText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	value = vietnameseNormalizer.Replace(value)
	value = expandCommonAbbreviations(value)

	var builder strings.Builder
	lastWasSpace := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			builder.WriteRune(r)
			lastWasSpace = false
		case unicode.IsSpace(r):
			if !lastWasSpace {
				builder.WriteByte(' ')
				lastWasSpace = true
			}
		default:
			if !lastWasSpace {
				builder.WriteByte(' ')
				lastWasSpace = true
			}
		}
	}

	return strings.Join(strings.Fields(builder.String()), " ")
}

func expandCommonAbbreviations(value string) string {
	replacer := strings.NewReplacer(
		"tp.hcm", "thanh pho ho chi minh",
		"tp hcm", "thanh pho ho chi minh",
		"tphcm", "thanh pho ho chi minh",
		"tp.ho chi minh", "thanh pho ho chi minh",
		"tp ho chi minh", "thanh pho ho chi minh",
		"tpho chi minh", "thanh pho ho chi minh",
	)
	return replacer.Replace(value)
}

func similarity(left string, right string) float64 {
	leftGrams := bigrams(normalizePublishText(left))
	rightGrams := bigrams(normalizePublishText(right))
	if len(leftGrams) == 0 || len(rightGrams) == 0 {
		return 0
	}

	intersection := 0
	for gram := range leftGrams {
		if _, ok := rightGrams[gram]; ok {
			intersection++
		}
	}
	return float64(2*intersection) / float64(len(leftGrams)+len(rightGrams))
}

func bigrams(value string) map[string]struct{} {
	runes := []rune(value)
	if len(runes) < 2 {
		return nil
	}
	grams := make(map[string]struct{}, len(runes)-1)
	for i := 0; i < len(runes)-1; i++ {
		grams[string(runes[i:i+2])] = struct{}{}
	}
	return grams
}
