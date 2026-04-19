package domain

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, source Source) (Source, error)
	ListAll(ctx context.Context) ([]Source, error)
	ListActive(ctx context.Context) ([]Source, error)
	ListActiveDueForValidation(ctx context.Context, limit int) ([]Source, error)
	ListInactiveDueForRecheck(ctx context.Context, now time.Time, limit int) ([]Source, error)
	TouchCrawl(ctx context.Context, id string, at time.Time) error
	MarkChecked(ctx context.Context, id string, at time.Time) error
	MarkInactive(ctx context.Context, id string, reason string, at time.Time) error
	MarkActive(ctx context.Context, id string, at time.Time) error
	UpdateMetadataByHandle(ctx context.Context, sourceType Type, handle string, tags []string, topics []string) error
	DeleteByHandle(ctx context.Context, sourceType Type, handle string) error
}
