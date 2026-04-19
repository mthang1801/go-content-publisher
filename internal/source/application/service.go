package application

import (
	"context"
	"fmt"
	"time"

	"go-content-bot/internal/source/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, source domain.Source) (domain.Source, error) {
	if err := source.Validate(); err != nil {
		return domain.Source{}, fmt.Errorf("validate source: %w", err)
	}
	return s.repo.Create(ctx, source)
}

func (s *Service) ListAll(ctx context.Context) ([]domain.Source, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) ListActive(ctx context.Context) ([]domain.Source, error) {
	return s.repo.ListActive(ctx)
}

func (s *Service) ListActiveDueForValidation(ctx context.Context, limit int) ([]domain.Source, error) {
	return s.repo.ListActiveDueForValidation(ctx, limit)
}

func (s *Service) ListInactiveDueForRecheck(ctx context.Context, now time.Time, limit int) ([]domain.Source, error) {
	return s.repo.ListInactiveDueForRecheck(ctx, now, limit)
}

func (s *Service) TouchCrawl(ctx context.Context, id string, at time.Time) error {
	return s.repo.TouchCrawl(ctx, id, at)
}

func (s *Service) MarkChecked(ctx context.Context, id string, at time.Time) error {
	return s.repo.MarkChecked(ctx, id, at)
}

func (s *Service) MarkInactive(ctx context.Context, id string, reason string, at time.Time) error {
	return s.repo.MarkInactive(ctx, id, reason, at)
}

func (s *Service) MarkActive(ctx context.Context, id string, at time.Time) error {
	return s.repo.MarkActive(ctx, id, at)
}

func (s *Service) UpdateMetadataByHandle(ctx context.Context, sourceType domain.Type, handle string, tags []string, topics []string) error {
	source := domain.Source{
		Type:   sourceType,
		Handle: handle,
		Name:   "placeholder",
		Tags:   tags,
		Topics: topics,
	}
	if err := source.Validate(); err != nil {
		return fmt.Errorf("validate source metadata: %w", err)
	}
	return s.repo.UpdateMetadataByHandle(ctx, sourceType, handle, source.Tags, source.Topics)
}

func (s *Service) DeleteByHandle(ctx context.Context, sourceType domain.Type, handle string) error {
	return s.repo.DeleteByHandle(ctx, sourceType, handle)
}
