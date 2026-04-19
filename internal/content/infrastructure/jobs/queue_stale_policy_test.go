package jobs

import (
	"context"
	"testing"
	"time"

	contentapp "go-content-bot/internal/content/application"
	contentdomain "go-content-bot/internal/content/domain"
)

type queuePolicyRepoStub struct {
	skipPendingFn   func(ctx context.Context, staleBefore time.Time, reason string) (int64, error)
	skipRewrittenFn func(ctx context.Context, staleBefore time.Time, reason string) (int64, error)
}

func (s queuePolicyRepoStub) CreatePending(ctx context.Context, item contentdomain.ContentItem) (contentdomain.ContentItem, error) {
	return contentdomain.ContentItem{}, nil
}

func (s queuePolicyRepoStub) SkipStalePending(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return s.skipPendingFn(ctx, staleBefore, reason)
}

func (s queuePolicyRepoStub) SkipStaleRewritten(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return s.skipRewrittenFn(ctx, staleBefore, reason)
}

func (s queuePolicyRepoStub) ClaimNextPending(ctx context.Context) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s queuePolicyRepoStub) ClaimNextReadyForPublish(ctx context.Context) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s queuePolicyRepoStub) FindNextPublishedReadyForTwitter(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s queuePolicyRepoStub) FindNextPending(ctx context.Context) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s queuePolicyRepoStub) FindByID(ctx context.Context, id string) (*contentdomain.ContentItem, error) {
	return nil, nil
}

func (s queuePolicyRepoStub) Save(ctx context.Context, item contentdomain.ContentItem) error {
	return nil
}

func (s queuePolicyRepoStub) ListByStatuses(ctx context.Context, statuses []contentdomain.Status, limit int) ([]contentdomain.ContentItem, error) {
	return nil, nil
}

func (s queuePolicyRepoStub) ListRecent(ctx context.Context, limit int) ([]contentdomain.ContentItem, error) {
	return nil, nil
}

type queueSettingsStub struct {
	values map[string]string
}

func (s queueSettingsStub) Get(ctx context.Context, key string) (string, error) {
	return s.values[key], nil
}

func TestSkipStalePendingUsesConfiguredSeconds(t *testing.T) {
	t.Parallel()

	var seenReason string
	service := contentapp.NewService(queuePolicyRepoStub{
		skipPendingFn: func(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
			seenReason = reason
			return 2, nil
		},
		skipRewrittenFn: func(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
			return 0, nil
		},
	})

	skipped, err := SkipStalePendingForRun(context.Background(), service, queueSettingsStub{values: map[string]string{
		pendingStaleAfterSecondsSettingKey: "300",
	}})
	if err != nil {
		t.Fatalf("skip stale pending: %v", err)
	}
	if skipped != 2 {
		t.Fatalf("expected 2 skipped items, got %d", skipped)
	}
	if seenReason != stalePendingReason {
		t.Fatalf("expected reason %q, got %q", stalePendingReason, seenReason)
	}
}

func TestSkipStaleRewrittenDisabledWhenSettingEmpty(t *testing.T) {
	t.Parallel()

	called := false
	service := contentapp.NewService(queuePolicyRepoStub{
		skipPendingFn: func(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
			return 0, nil
		},
		skipRewrittenFn: func(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
			called = true
			return 0, nil
		},
	})

	skipped, err := SkipStaleRewrittenForRun(context.Background(), service, queueSettingsStub{values: map[string]string{}})
	if err != nil {
		t.Fatalf("skip stale rewritten: %v", err)
	}
	if skipped != 0 {
		t.Fatalf("expected 0 skipped items, got %d", skipped)
	}
	if called {
		t.Fatal("expected rewritten stale skip to stay disabled when setting is empty")
	}
}
