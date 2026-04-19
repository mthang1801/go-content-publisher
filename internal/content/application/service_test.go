package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-content-bot/internal/content/domain"
)

type contentRepoServiceStub struct {
	createPendingFn                    func(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error)
	claimNextPendingFn                 func(ctx context.Context) (*domain.ContentItem, error)
	claimNextReadyForPublishFn         func(ctx context.Context) (*domain.ContentItem, error)
	findNextPublishedReadyForTwitterFn func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*domain.ContentItem, error)
	findNextPendingFn                  func(ctx context.Context) (*domain.ContentItem, error)
	findByIDFn                         func(ctx context.Context, id string) (*domain.ContentItem, error)
	saveFn                             func(ctx context.Context, item domain.ContentItem) error
	listByStatusesFn                   func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error)
	listRecentFn                       func(ctx context.Context, limit int) ([]domain.ContentItem, error)
}

func (s contentRepoServiceStub) CreatePending(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error) {
	return s.createPendingFn(ctx, item)
}

func (s contentRepoServiceStub) SkipStalePending(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return 0, nil
}

func (s contentRepoServiceStub) SkipStaleRewritten(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return 0, nil
}

func (s contentRepoServiceStub) ClaimNextPending(ctx context.Context) (*domain.ContentItem, error) {
	if s.claimNextPendingFn == nil {
		return nil, nil
	}
	return s.claimNextPendingFn(ctx)
}

func (s contentRepoServiceStub) ClaimNextReadyForPublish(ctx context.Context) (*domain.ContentItem, error) {
	if s.claimNextReadyForPublishFn == nil {
		return nil, nil
	}
	return s.claimNextReadyForPublishFn(ctx)
}

func (s contentRepoServiceStub) FindNextPublishedReadyForTwitter(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*domain.ContentItem, error) {
	if s.findNextPublishedReadyForTwitterFn == nil {
		return nil, nil
	}
	return s.findNextPublishedReadyForTwitterFn(ctx, publishAfter, sourceTypes, sourceTags, sourceTopics, topicKeywords)
}

func (s contentRepoServiceStub) FindNextPending(ctx context.Context) (*domain.ContentItem, error) {
	if s.findNextPendingFn == nil {
		return nil, nil
	}
	return s.findNextPendingFn(ctx)
}

func (s contentRepoServiceStub) FindByID(ctx context.Context, id string) (*domain.ContentItem, error) {
	if s.findByIDFn == nil {
		return nil, nil
	}
	return s.findByIDFn(ctx, id)
}

func (s contentRepoServiceStub) Save(ctx context.Context, item domain.ContentItem) error {
	if s.saveFn == nil {
		return nil
	}
	return s.saveFn(ctx, item)
}

func (s contentRepoServiceStub) ListByStatuses(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
	if s.listByStatusesFn == nil {
		return nil, nil
	}
	return s.listByStatusesFn(ctx, statuses, limit)
}

func (s contentRepoServiceStub) ListRecent(ctx context.Context, limit int) ([]domain.ContentItem, error) {
	if s.listRecentFn == nil {
		return nil, nil
	}
	return s.listRecentFn(ctx, limit)
}

type rewriteTextStub struct {
	rewriteFn func(ctx context.Context, originalText string) (RewriteResult, error)
}

func (s rewriteTextStub) RewriteText(ctx context.Context, originalText string) (RewriteResult, error) {
	return s.rewriteFn(ctx, originalText)
}

type publishTextStub struct {
	publishFn func(ctx context.Context, text string) (string, error)
}

func (s publishTextStub) PublishText(ctx context.Context, text string) (string, error) {
	return s.publishFn(ctx, text)
}

type publishTweetsStub struct {
	publishFn func(ctx context.Context, input TwitterPublishInput) (TwitterPublishResult, error)
}

func (s publishTweetsStub) PublishTweets(ctx context.Context, input TwitterPublishInput) (TwitterPublishResult, error) {
	return s.publishFn(ctx, input)
}

func TestEnqueueManualCreatesPendingItem(t *testing.T) {
	repo := contentRepoServiceStub{
		createPendingFn: func(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error) {
			if item.Status != domain.StatusPending {
				t.Fatalf("expected pending status, got %s", item.Status)
			}
			if item.AuthorName != "Manual" {
				t.Fatalf("expected Manual author, got %s", item.AuthorName)
			}
			item.ID = "item-1"
			return item, nil
		},
	}

	service := NewService(repo)
	item, err := service.EnqueueManual(context.Background(), EnqueueManualInput{
		Text:   "hello world",
		Author: "",
	})
	if err != nil {
		t.Fatalf("enqueue manual: %v", err)
	}
	if item.ID != "item-1" {
		t.Fatalf("expected item id item-1, got %s", item.ID)
	}
}

func TestEnqueueFromSourceCreatesPendingItem(t *testing.T) {
	sourceID := "source-1"
	repo := contentRepoServiceStub{
		createPendingFn: func(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error) {
			if item.Status != domain.StatusPending {
				t.Fatalf("expected pending status, got %s", item.Status)
			}
			if item.SourceID == nil || *item.SourceID != sourceID {
				t.Fatalf("expected source id %s, got %#v", sourceID, item.SourceID)
			}
			if item.ExternalID != "tg:-100:55" {
				t.Fatalf("expected external id tg:-100:55, got %s", item.ExternalID)
			}
			item.ID = "item-source-1"
			return item, nil
		},
	}

	service := NewService(repo)
	item, err := service.EnqueueFromSource(context.Background(), EnqueueSourceInput{
		SourceID:     sourceID,
		ExternalID:   "tg:-100:55",
		OriginalText: "telegram content",
		AuthorName:   "Coding",
	})
	if err != nil {
		t.Fatalf("enqueue from source: %v", err)
	}
	if item.ID != "item-source-1" {
		t.Fatalf("expected item id item-source-1, got %s", item.ID)
	}
}

func TestRetryFailedMovesFailedItemsBackToPending(t *testing.T) {
	saved := make([]domain.ContentItem, 0)
	repo := contentRepoServiceStub{
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			if len(statuses) != 1 || statuses[0] != domain.StatusFailed {
				t.Fatalf("expected failed status filter, got %#v", statuses)
			}
			return []domain.ContentItem{
				{ID: "item-1", OriginalText: "failed", AuthorName: "bot", CrawledAt: time.Now(), Status: domain.StatusFailed, FailReason: stringPtr("temporary")},
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	count, err := NewService(repo).RetryFailed(context.Background(), 10)
	if err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 retried item, got %d", count)
	}
	if len(saved) != 1 || saved[0].Status != domain.StatusPending || saved[0].FailReason != nil {
		t.Fatalf("unexpected saved item: %#v", saved)
	}
}

func TestSkipByIDPrefixMarksSingleRecentMatchSkipped(t *testing.T) {
	saved := make([]domain.ContentItem, 0)
	repo := contentRepoServiceStub{
		listRecentFn: func(ctx context.Context, limit int) ([]domain.ContentItem, error) {
			return []domain.ContentItem{
				{ID: "12345678-aaaa", OriginalText: "skip me", AuthorName: "bot", CrawledAt: time.Now(), Status: domain.StatusPending},
				{ID: "99999999-bbbb", OriginalText: "keep me", AuthorName: "bot", CrawledAt: time.Now(), Status: domain.StatusPending},
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	item, err := NewService(repo).SkipByIDPrefix(context.Background(), "1234", "admin skip")
	if err != nil {
		t.Fatalf("skip by id prefix: %v", err)
	}
	if item.ID != "12345678-aaaa" || item.Status != domain.StatusSkipped {
		t.Fatalf("unexpected skipped item: %#v", item)
	}
	if len(saved) != 1 || saved[0].FailReason == nil || *saved[0].FailReason != "admin skip" {
		t.Fatalf("unexpected saved item: %#v", saved)
	}
}

func TestSkipByIDPrefixRejectsAmbiguousPrefix(t *testing.T) {
	repo := contentRepoServiceStub{
		listRecentFn: func(ctx context.Context, limit int) ([]domain.ContentItem, error) {
			return []domain.ContentItem{
				{ID: "12345678-aaaa", OriginalText: "one", AuthorName: "bot", CrawledAt: time.Now(), Status: domain.StatusPending},
				{ID: "12349999-bbbb", OriginalText: "two", AuthorName: "bot", CrawledAt: time.Now(), Status: domain.StatusPending},
			}, nil
		},
	}

	if _, err := NewService(repo).SkipByIDPrefix(context.Background(), "1234", "admin skip"); err == nil {
		t.Fatal("expected ambiguous prefix error")
	}
}

func TestProcessNextPendingRewritesItem(t *testing.T) {
	saved := []domain.ContentItem{}
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			item := domain.ContentItem{
				ID:           "item-1",
				OriginalText: "original",
				AuthorName:   "author",
				CrawledAt:    time.Now(),
				Status:       domain.StatusProcessing,
			}
			return &item, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			if originalText != "original" {
				t.Fatalf("expected original text, got %s", originalText)
			}
			return RewriteResult{
				RewrittenText:   "rewritten text",
				RewrittenTextEn: "rewritten english",
				TweetTextVI:     "tweet vi",
				TweetTextEN:     "tweet en",
				FactCheckNote:   "fact checked",
				ShouldPublish:   true,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("process next pending: %v", err)
	}
	if item == nil || item.Status != domain.StatusRewritten {
		t.Fatalf("expected rewritten item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save after claim (rewritten only), got %d", len(saved))
	}
	if saved[0].RewrittenTextEn == nil || *saved[0].RewrittenTextEn != "rewritten english" {
		t.Fatalf("expected rewritten english field, got %#v", saved[0].RewrittenTextEn)
	}
	if saved[0].TweetTextVI == nil || *saved[0].TweetTextVI != "tweet vi" {
		t.Fatalf("expected tweet vi field, got %#v", saved[0].TweetTextVI)
	}
	if saved[0].FactCheckNote == nil || *saved[0].FactCheckNote != "fact checked" {
		t.Fatalf("expected fact check note field, got %#v", saved[0].FactCheckNote)
	}
}

func TestProcessNextPendingSkipsWhenRewriteSaysShouldNotPublish(t *testing.T) {
	saved := []domain.ContentItem{}
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:           "item-skip-ai",
				OriginalText: "Join this channel and claim free bonus",
				AuthorName:   "author",
				CrawledAt:    time.Now(),
				Status:       domain.StatusProcessing,
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			return RewriteResult{
				ShouldPublish: false,
				Reason:        "Nội dung quảng cáo/referral",
				FactCheckNote: "Không phù hợp để đăng",
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("process next pending with shouldPublish=false: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(saved))
	}
	if saved[0].FailReason == nil || *saved[0].FailReason != "Nội dung quảng cáo/referral" {
		t.Fatalf("expected skip reason, got %#v", saved[0].FailReason)
	}
	if saved[0].FactCheckNote == nil || *saved[0].FactCheckNote != "Không phù hợp để đăng" {
		t.Fatalf("expected fact check note, got %#v", saved[0].FactCheckNote)
	}
}

func TestProcessNextPendingSkipsPreRewriteDuplicate(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:           "item-pre-dup",
				OriginalText: "Fed keeps rates unchanged as markets watch inflation",
				AuthorName:   "author",
				CrawledAt:    now,
				Status:       domain.StatusProcessing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			return []domain.ContentItem{
				{
					ID:           "item-existing",
					OriginalText: "Fed keeps rates unchanged as markets watch inflation",
					AuthorName:   "author",
					CrawledAt:    now.Add(-time.Hour),
					Status:       domain.StatusPublished,
				},
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			t.Fatal("rewriter should not be called for pre-rewrite duplicate")
			return RewriteResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("process duplicate pending: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(saved))
	}
}

func TestProcessNextPendingAllowsManualNearDuplicate(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:           "item-manual-near-dup",
				OriginalText: "Hôm nay thời tiết TP.HCM ra sao? Dự báo tuần tới giúp tôi nhé",
				AuthorName:   "Manual",
				CrawledAt:    now,
				Status:       domain.StatusProcessing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			return []domain.ContentItem{
				{
					ID:           "item-existing-weather",
					OriginalText: "Xin chào, hôm nay thời tiết tại thành phố Hồ Chí Minh như thế nào nhỉ? Dự báo thời tiết cho tuần tới giúp tôi nhé",
					AuthorName:   "Manual",
					CrawledAt:    now.Add(-time.Hour),
					Status:       domain.StatusPublished,
				},
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			return RewriteResult{
				RewrittenText: "Tổng hợp nhanh thời tiết TP.HCM hôm nay và xu hướng tuần tới.",
				ShouldPublish: true,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("process manual near-duplicate pending: %v", err)
	}
	if item == nil || item.Status != domain.StatusRewritten {
		t.Fatalf("expected rewritten manual item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(saved))
	}
}

func TestProcessNextPendingSkipsExactManualDuplicatePreRewrite(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:           "item-manual-exact-dup",
				OriginalText: "Xin chào thế giới!!!",
				AuthorName:   "Manual",
				CrawledAt:    now,
				Status:       domain.StatusProcessing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			return []domain.ContentItem{
				{
					ID:           "item-existing-manual",
					OriginalText: "xin chào   thế giới",
					AuthorName:   "Manual",
					CrawledAt:    now.Add(-time.Hour),
					Status:       domain.StatusPublished,
				},
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			t.Fatal("rewriter should not be called for exact manual duplicate")
			return RewriteResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("process exact manual duplicate pending: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped manual item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(saved))
	}
	if saved[0].FailReason == nil || *saved[0].FailReason != "duplicate manual content already processed recently" {
		t.Fatalf("expected exact manual duplicate reason, got %#v", saved[0].FailReason)
	}
}

func TestProcessNextPendingSkipsPostRewriteDuplicate(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	listCalls := 0
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:           "item-post-dup",
				OriginalText: "new source text",
				AuthorName:   "author",
				CrawledAt:    now,
				Status:       domain.StatusProcessing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			listCalls++
			if listCalls == 1 {
				return nil, nil
			}
			return []domain.ContentItem{
				{
					ID:            "item-existing-rewrite",
					OriginalText:  "older source text",
					RewrittenText: strPtr("Giá dầu tăng khi căng thẳng địa chính trị leo thang."),
					AuthorName:    "author",
					CrawledAt:     now.Add(-time.Hour),
					Status:        domain.StatusPublished,
				},
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			return RewriteResult{
				RewrittenText: "Giá dầu tăng khi căng thẳng địa chính trị leo thang.",
				ShouldPublish: true,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("process post-rewrite duplicate: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(saved))
	}
	if saved[0].RewrittenText == nil {
		t.Fatal("expected rewritten text to be preserved on skipped item")
	}
}

func TestProcessNextPendingSkipsExactManualDuplicatePostRewrite(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	listCalls := 0
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:           "item-manual-post-dup",
				OriginalText: "Thời tiết Sài Gòn tuần tới",
				AuthorName:   "Manual",
				CrawledAt:    now,
				Status:       domain.StatusProcessing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			listCalls++
			if listCalls == 1 {
				return nil, nil
			}
			return []domain.ContentItem{
				{
					ID:            "item-existing-manual-rewrite",
					OriginalText:  "Khác",
					RewrittenText: strPtr("Thời tiết thành phố Hồ Chí Minh tuần tới"),
					AuthorName:    "Manual",
					CrawledAt:     now.Add(-time.Hour),
					Status:        domain.StatusPublished,
				},
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			return RewriteResult{
				RewrittenText: "Thời tiết TP.HCM tuần tới",
				ShouldPublish: true,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("process exact manual duplicate post-rewrite: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped manual item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(saved))
	}
	if saved[0].FailReason == nil || *saved[0].FailReason != "duplicate manual rewritten content already processed recently" {
		t.Fatalf("expected manual post-rewrite duplicate reason, got %#v", saved[0].FailReason)
	}
}

func TestPublishNextReadyPublishesRewrittenItem(t *testing.T) {
	saved := []domain.ContentItem{}
	repo := contentRepoServiceStub{
		claimNextReadyForPublishFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:            "item-2",
				OriginalText:  "original",
				RewrittenText: strPtr("ready for telegram"),
				AuthorName:    "author",
				CrawledAt:     time.Now(),
				Status:        domain.StatusPublishing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			return nil, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.PublishNextReady(context.Background(), publishTextStub{
		publishFn: func(ctx context.Context, text string) (string, error) {
			if text != "ready for telegram" {
				t.Fatalf("expected rewritten text, got %s", text)
			}
			return "137", nil
		},
	})
	if err != nil {
		t.Fatalf("publish next ready: %v", err)
	}
	if item == nil || item.Status != domain.StatusPublished {
		t.Fatalf("expected published item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save on publish, got %d", len(saved))
	}
}

func TestPublishNextReadySkipsDuplicateByPublishedRewrittenText(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	repo := contentRepoServiceStub{
		claimNextReadyForPublishFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:            "item-dup-1",
				OriginalText:  "raw text",
				RewrittenText: strPtr("Xin chào    thế giới!!!"),
				AuthorName:    "author",
				CrawledAt:     now,
				Status:        domain.StatusPublishing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			if len(statuses) == 1 && statuses[0] == domain.StatusPublished {
				return []domain.ContentItem{
					{
						ID:            "item-published-1",
						OriginalText:  "something else",
						RewrittenText: strPtr("xin chào thế giới"),
						AuthorName:    "author",
						CrawledAt:     now.Add(-time.Hour),
						Status:        domain.StatusPublished,
						PublishedAt:   timePtr(now.Add(-time.Hour)),
					},
				}, nil
			}
			return nil, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.PublishNextReady(context.Background(), publishTextStub{
		publishFn: func(ctx context.Context, text string) (string, error) {
			t.Fatal("publisher should not be called for duplicate content")
			return "", nil
		},
	})
	if err != nil {
		t.Fatalf("publish next ready with duplicate rewritten text: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save for skipped item, got %d", len(saved))
	}
	if saved[0].FailReason == nil || *saved[0].FailReason == "" {
		t.Fatalf("expected skip reason to be recorded, got %#v", saved[0].FailReason)
	}
}

func TestPublishNextReadySkipsDuplicateByPublishedOriginalText(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	repo := contentRepoServiceStub{
		claimNextReadyForPublishFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:            "item-dup-2",
				OriginalText:  "ignored",
				RewrittenText: strPtr("Dự báo thời tiết TP.HCM tuần tới"),
				AuthorName:    "author",
				CrawledAt:     now,
				Status:        domain.StatusPublishing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			if len(statuses) == 1 && statuses[0] == domain.StatusPublished {
				return []domain.ContentItem{
					{
						ID:           "item-published-2",
						OriginalText: "du bao THOI tiet tp hcm tuan toi",
						AuthorName:   "author",
						CrawledAt:    now.Add(-2 * time.Hour),
						Status:       domain.StatusPublished,
						PublishedAt:  timePtr(now.Add(-2 * time.Hour)),
					},
				}, nil
			}
			return nil, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.PublishNextReady(context.Background(), publishTextStub{
		publishFn: func(ctx context.Context, text string) (string, error) {
			t.Fatal("publisher should not be called for duplicate original text")
			return "", nil
		},
	})
	if err != nil {
		t.Fatalf("publish next ready with duplicate original text: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save for skipped item, got %d", len(saved))
	}
}

func TestPublishNextReadySkipsDuplicateWhenOriginalMatchesButRewriteDiffers(t *testing.T) {
	saved := []domain.ContentItem{}
	now := time.Now()
	original := "Xin chào, hôm nay thời tiết tại thành phố Hồ Chí Minh như thế nào nhỉ? Dự báo thời tiết cho tuần tới giúp tôi nhé"

	repo := contentRepoServiceStub{
		claimNextReadyForPublishFn: func(ctx context.Context) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:            "item-dup-live",
				OriginalText:  original,
				RewrittenText: strPtr("Thời tiết TP.HCM hôm nay và dự báo tuần tới."),
				AuthorName:    "author",
				CrawledAt:     now,
				Status:        domain.StatusPublishing,
			}, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			if len(statuses) == 1 && statuses[0] == domain.StatusPublished {
				return []domain.ContentItem{
					{
						ID:            "item-published-live",
						OriginalText:  original,
						RewrittenText: strPtr("Thời tiết Thành phố Hồ Chí Minh hôm nay và dự báo tuần tới."),
						AuthorName:    "author",
						CrawledAt:     now.Add(-time.Minute),
						Status:        domain.StatusPublished,
						PublishedAt:   timePtr(now.Add(-time.Minute)),
					},
				}, nil
			}
			return nil, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.PublishNextReady(context.Background(), publishTextStub{
		publishFn: func(ctx context.Context, text string) (string, error) {
			t.Fatal("publisher should not be called when original text already matches recent published item")
			return "", nil
		},
	})
	if err != nil {
		t.Fatalf("publish next ready with matching original text: %v", err)
	}
	if item == nil || item.Status != domain.StatusSkipped {
		t.Fatalf("expected skipped item, got %#v", item)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save for skipped item, got %d", len(saved))
	}
}

func TestPublishNextToTwitterPublishesAvailableTexts(t *testing.T) {
	saved := []domain.ContentItem{}
	repo := contentRepoServiceStub{
		findNextPublishedReadyForTwitterFn: func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*domain.ContentItem, error) {
			return &domain.ContentItem{
				ID:              "item-twitter-1",
				OriginalText:    "original",
				RewrittenText:   strPtr("ban viet lai tieng viet"),
				RewrittenTextEn: strPtr("rewritten english"),
				TweetTextVI:     strPtr("tweet vi ngan"),
				AuthorName:      "author",
				CrawledAt:       time.Now(),
				Status:          domain.StatusPublished,
			}, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.PublishNextToTwitter(context.Background(), publishTweetsStub{
		publishFn: func(ctx context.Context, input TwitterPublishInput) (TwitterPublishResult, error) {
			if input.TweetTextVI == nil || *input.TweetTextVI != "tweet vi ngan" {
				t.Fatalf("expected tweet vi ngan, got %#v", input.TweetTextVI)
			}
			if input.TweetTextEN == nil || *input.TweetTextEN != "rewritten english" {
				t.Fatalf("expected english fallback, got %#v", input.TweetTextEN)
			}
			return TwitterPublishResult{
				TweetViID: "tweet-vi-1",
				TweetEnID: "tweet-en-1",
			}, nil
		},
	}, nil)
	if err != nil {
		t.Fatalf("publish next to twitter: %v", err)
	}
	if item == nil {
		t.Fatal("expected item to be returned")
	}
	if item.TweetViID == nil || *item.TweetViID != "tweet-vi-1" {
		t.Fatalf("expected tweet vi id to be saved, got %#v", item.TweetViID)
	}
	if item.TweetEnID == nil || *item.TweetEnID != "tweet-en-1" {
		t.Fatalf("expected tweet en id to be saved, got %#v", item.TweetEnID)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save on twitter publish, got %d", len(saved))
	}
}

func TestPublishNextToTwitterReturnsIdleWhenNoEligibleItem(t *testing.T) {
	service := NewService(contentRepoServiceStub{
		findNextPublishedReadyForTwitterFn: func(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*domain.ContentItem, error) {
			return nil, nil
		},
	})

	item, err := service.PublishNextToTwitter(context.Background(), publishTweetsStub{
		publishFn: func(ctx context.Context, input TwitterPublishInput) (TwitterPublishResult, error) {
			t.Fatal("twitter publisher should not be called without eligible item")
			return TwitterPublishResult{}, nil
		},
	}, nil)
	if err != nil {
		t.Fatalf("publish next to twitter idle: %v", err)
	}
	if item != nil {
		t.Fatalf("expected nil item, got %#v", item)
	}
}

func TestProcessNextPendingMarksFailureWhenRewriteFails(t *testing.T) {
	saved := []domain.ContentItem{}
	repo := contentRepoServiceStub{
		claimNextPendingFn: func(ctx context.Context) (*domain.ContentItem, error) {
			item := domain.ContentItem{
				ID:           "item-3",
				OriginalText: "original",
				AuthorName:   "author",
				CrawledAt:    time.Now(),
				Status:       domain.StatusProcessing,
			}
			return &item, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	_, err := service.ProcessNextPending(context.Background(), rewriteTextStub{
		rewriteFn: func(ctx context.Context, originalText string) (RewriteResult, error) {
			return RewriteResult{}, errors.New("deepseek unavailable")
		},
	})
	if err == nil {
		t.Fatal("expected rewrite error")
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save after claim (failed only), got %d", len(saved))
	}
	if saved[0].Status != domain.StatusFailed {
		t.Fatalf("expected failed status, got %s", saved[0].Status)
	}
}

func TestSetManualRewriteUpdatesFailedItem(t *testing.T) {
	saved := []domain.ContentItem{}
	repo := contentRepoServiceStub{
		findByIDFn: func(ctx context.Context, id string) (*domain.ContentItem, error) {
			item := domain.ContentItem{
				ID:           id,
				OriginalText: "original",
				AuthorName:   "author",
				CrawledAt:    time.Now(),
				Status:       domain.StatusFailed,
			}
			return &item, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			saved = append(saved, item)
			return nil
		},
	}

	service := NewService(repo)
	item, err := service.SetManualRewrite(context.Background(), "item-4", "manual rewrite text")
	if err != nil {
		t.Fatalf("set manual rewrite: %v", err)
	}
	if item.Status != domain.StatusRewritten {
		t.Fatalf("expected rewritten status, got %s", item.Status)
	}
	if item.RewrittenText == nil || *item.RewrittenText != "manual rewrite text" {
		t.Fatalf("expected rewritten text to be updated, got %#v", item.RewrittenText)
	}
	if len(saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(saved))
	}
}

func strPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
