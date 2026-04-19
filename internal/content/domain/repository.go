package domain

import (
	"context"
	"time"
)

type Repository interface {
	CreatePending(ctx context.Context, item ContentItem) (ContentItem, error)
	SkipStalePending(ctx context.Context, staleBefore time.Time, reason string) (int64, error)
	SkipStaleRewritten(ctx context.Context, staleBefore time.Time, reason string) (int64, error)
	ClaimNextPending(ctx context.Context) (*ContentItem, error)
	ClaimNextReadyForPublish(ctx context.Context) (*ContentItem, error)
	FindNextPublishedReadyForTwitter(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*ContentItem, error)
	FindNextPending(ctx context.Context) (*ContentItem, error)
	FindByID(ctx context.Context, id string) (*ContentItem, error)
	Save(ctx context.Context, item ContentItem) error
	ListByStatuses(ctx context.Context, statuses []Status, limit int) ([]ContentItem, error)
	ListRecent(ctx context.Context, limit int) ([]ContentItem, error)
}
