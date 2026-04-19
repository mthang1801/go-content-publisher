package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-content-bot/internal/source/domain"
	"go-content-bot/internal/source/infrastructure/persistence/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type SourceRepository struct {
	db *gorm.DB
}

func NewSourceRepository(db *gorm.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) Create(ctx context.Context, source domain.Source) (domain.Source, error) {
	model := models.SourceModel{
		ID:            source.ID,
		Type:          string(source.Type),
		Handle:        source.Handle,
		Name:          source.Name,
		Tags:          pq.StringArray(source.Tags),
		Topics:        pq.StringArray(source.Topics),
		IsActive:      true,
		LastCrawledAt: source.LastCrawledAt,
		LastCheckAt:   source.LastCheckAt,
		LastError:     source.LastError,
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isDuplicateErr(err) {
			return domain.Source{}, domain.ErrSourceAlreadyExists
		}
		return domain.Source{}, fmt.Errorf("create source: %w", err)
	}

	return mapSource(model), nil
}

func (r *SourceRepository) ListAll(ctx context.Context) ([]domain.Source, error) {
	var rows []models.SourceModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list all sources: %w", err)
	}
	return mapSources(rows), nil
}

func (r *SourceRepository) ListActive(ctx context.Context) ([]domain.Source, error) {
	var rows []models.SourceModel
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list active sources: %w", err)
	}

	sources := make([]domain.Source, 0, len(rows))
	for _, row := range rows {
		sources = append(sources, mapSource(row))
	}
	return sources, nil
}

func (r *SourceRepository) ListActiveDueForValidation(ctx context.Context, limit int) ([]domain.Source, error) {
	var rows []models.SourceModel
	query := r.db.WithContext(ctx).
		Where("is_active = ? AND last_check_at IS NULL", true).
		Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list active sources due for validation: %w", err)
	}
	return mapSources(rows), nil
}

func (r *SourceRepository) ListInactiveDueForRecheck(ctx context.Context, now time.Time, limit int) ([]domain.Source, error) {
	cutoff := now.Add(-24 * time.Hour)

	var rows []models.SourceModel
	query := r.db.WithContext(ctx).
		Where("is_active = ? AND COALESCE(last_check_at, last_crawled_at, created_at) <= ?", false, cutoff).
		Order("COALESCE(last_check_at, last_crawled_at, created_at) ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list inactive sources due for recheck: %w", err)
	}
	return mapSources(rows), nil
}

func (r *SourceRepository) TouchCrawl(ctx context.Context, id string, at time.Time) error {
	if err := r.db.WithContext(ctx).
		Model(&models.SourceModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"last_crawled_at": at}).Error; err != nil {
		return fmt.Errorf("touch source crawl %s: %w", id, err)
	}
	return nil
}

func (r *SourceRepository) MarkChecked(ctx context.Context, id string, at time.Time) error {
	if err := r.db.WithContext(ctx).
		Model(&models.SourceModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"last_check_at": at, "last_error": nil}).Error; err != nil {
		return fmt.Errorf("mark source checked %s: %w", id, err)
	}
	return nil
}

func (r *SourceRepository) MarkInactive(ctx context.Context, id string, reason string, at time.Time) error {
	reason = strings.TrimSpace(reason)
	if err := r.db.WithContext(ctx).
		Model(&models.SourceModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_active":     false,
			"last_check_at": at,
			"last_error":    nullableString(reason),
		}).Error; err != nil {
		return fmt.Errorf("mark source inactive %s: %w", id, err)
	}
	return nil
}

func (r *SourceRepository) MarkActive(ctx context.Context, id string, at time.Time) error {
	if err := r.db.WithContext(ctx).
		Model(&models.SourceModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_active":     true,
			"last_check_at": at,
			"last_error":    nil,
		}).Error; err != nil {
		return fmt.Errorf("mark source active %s: %w", id, err)
	}
	return nil
}

func (r *SourceRepository) UpdateMetadataByHandle(ctx context.Context, sourceType domain.Type, handle string, tags []string, topics []string) error {
	if err := r.db.WithContext(ctx).
		Model(&models.SourceModel{}).
		Where("type = ? AND handle = ?", string(sourceType), handle).
		Updates(map[string]any{
			"tags":   pq.StringArray(tags),
			"topics": pq.StringArray(topics),
		}).Error; err != nil {
		return fmt.Errorf("update source metadata by handle: %w", err)
	}
	return nil
}

func (r *SourceRepository) DeleteByHandle(ctx context.Context, sourceType domain.Type, handle string) error {
	result := r.db.WithContext(ctx).Where("type = ? AND handle = ?", string(sourceType), handle).Delete(&models.SourceModel{})
	if result.Error != nil {
		return fmt.Errorf("delete source by handle: %w", result.Error)
	}
	return nil
}

func mapSource(model models.SourceModel) domain.Source {
	return domain.Source{
		ID:            model.ID,
		Type:          domain.Type(model.Type),
		Handle:        model.Handle,
		Name:          model.Name,
		Tags:          append([]string(nil), model.Tags...),
		Topics:        append([]string(nil), model.Topics...),
		IsActive:      model.IsActive,
		LastCrawledAt: model.LastCrawledAt,
		LastCheckAt:   model.LastCheckAt,
		LastError:     model.LastError,
		CreatedAt:     model.CreatedAt,
	}
}

func mapSources(rows []models.SourceModel) []domain.Source {
	sources := make([]domain.Source, 0, len(rows))
	for _, row := range rows {
		sources = append(sources, mapSource(row))
	}
	return sources
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func isDuplicateErr(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
