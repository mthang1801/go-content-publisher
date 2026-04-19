package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-content-bot/internal/content/domain"
	"go-content-bot/internal/content/infrastructure/persistence/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (r *ContentRepository) CreatePending(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error) {
	model := toModel(item)
	model.Status = string(domain.StatusPending)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isDuplicateErr(err) {
			return domain.ContentItem{}, domain.ErrContentAlreadyExists
		}
		return domain.ContentItem{}, fmt.Errorf("create pending content item: %w", err)
	}
	return toDomain(model), nil
}

func (r *ContentRepository) SkipStalePending(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.ContentItemModel{}).
		Where("status = ? AND crawled_at <= ?", string(domain.StatusPending), staleBefore).
		Updates(map[string]any{
			"status":      string(domain.StatusSkipped),
			"fail_reason": strings.TrimSpace(reason),
		})
	if result.Error != nil {
		return 0, fmt.Errorf("skip stale pending content items: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *ContentRepository) SkipStaleRewritten(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.ContentItemModel{}).
		Where("status = ? AND crawled_at <= ?", string(domain.StatusRewritten), staleBefore).
		Updates(map[string]any{
			"status":      string(domain.StatusSkipped),
			"fail_reason": strings.TrimSpace(reason),
		})
	if result.Error != nil {
		return 0, fmt.Errorf("skip stale rewritten content items: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *ContentRepository) FindNextPending(ctx context.Context) (*domain.ContentItem, error) {
	var model models.ContentItemModel
	err := r.db.WithContext(ctx).
		Where("status = ?", string(domain.StatusPending)).
		Order("crawled_at DESC").
		Limit(1).
		Take(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find next pending content item: %w", err)
	}
	item := toDomain(model)
	return &item, nil
}

func (r *ContentRepository) ClaimNextPending(ctx context.Context) (*domain.ContentItem, error) {
	var claimed *domain.ContentItem

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model models.ContentItemModel
		result := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", string(domain.StatusPending)).
			Order("crawled_at DESC").
			Limit(1).
			Find(&model)
		if result.Error != nil {
			return fmt.Errorf("claim next pending content item: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return nil
		}

		item := toDomain(model)
		if err := item.MarkProcessing(); err != nil {
			return fmt.Errorf("mark claimed content item processing: %w", err)
		}

		claimedModel := toModel(item)
		if err := tx.Save(&claimedModel).Error; err != nil {
			return fmt.Errorf("save claimed content item: %w", err)
		}

		claimedItem := toDomain(claimedModel)
		claimed = &claimedItem
		return nil
	})
	if err != nil {
		return nil, err
	}

	return claimed, nil
}

func (r *ContentRepository) ClaimNextReadyForPublish(ctx context.Context) (*domain.ContentItem, error) {
	var claimed *domain.ContentItem

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model models.ContentItemModel
		result := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", string(domain.StatusRewritten)).
			Order("crawled_at DESC").
			Limit(1).
			Find(&model)
		if result.Error != nil {
			return fmt.Errorf("claim next ready content item for publish: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return nil
		}

		item := toDomain(model)
		if err := item.MarkPublishing(); err != nil {
			return fmt.Errorf("mark claimed content item publishing: %w", err)
		}

		claimedModel := toModel(item)
		if err := tx.Save(&claimedModel).Error; err != nil {
			return fmt.Errorf("save claimed publish content item: %w", err)
		}

		claimedItem := toDomain(claimedModel)
		claimed = &claimedItem
		return nil
	})
	if err != nil {
		return nil, err
	}

	return claimed, nil
}

func (r *ContentRepository) FindNextPublishedReadyForTwitter(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*domain.ContentItem, error) {
	var model models.ContentItemModel
	query := r.db.WithContext(ctx).
		Table("content_items").
		Joins("LEFT JOIN sources ON sources.id = content_items.source_id").
		Where("status = ? AND rewritten_text IS NOT NULL AND (tweet_vi_id IS NULL OR tweet_en_id IS NULL)", string(domain.StatusPublished))
	if publishAfter != nil {
		query = query.Where("published_at IS NOT NULL AND published_at >= ?", *publishAfter)
	}
	if len(sourceTypes) > 0 {
		query = query.Where("sources.type IN ?", sourceTypes)
	}
	if len(sourceTags) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM unnest(COALESCE(sources.tags, ARRAY[]::text[])) AS source_tag WHERE source_tag IN ?)", sourceTags)
	}
	if len(sourceTopics) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM unnest(COALESCE(sources.topics, ARRAY[]::text[])) AS source_topic WHERE source_topic IN ?)", sourceTopics)
	}
	if len(topicKeywords) > 0 {
		orClauses := make([]string, 0, len(topicKeywords))
		args := make([]any, 0, len(topicKeywords)*6)
		for _, keyword := range topicKeywords {
			trimmed := strings.TrimSpace(keyword)
			if trimmed == "" {
				continue
			}
			pattern := "%" + trimmed + "%"
			orClauses = append(orClauses, `(content_items.original_text ILIKE ? OR content_items.rewritten_text ILIKE ? OR content_items.rewritten_text_en ILIKE ? OR content_items.tweet_text_vi ILIKE ? OR content_items.tweet_text_en ILIKE ? OR content_items.fact_check_note ILIKE ?)`)
			for i := 0; i < 6; i++ {
				args = append(args, pattern)
			}
		}
		if len(orClauses) > 0 {
			query = query.Where("("+strings.Join(orClauses, " OR ")+")", args...)
		}
	}
	err := query.Order("published_at ASC NULLS LAST, crawled_at ASC").
		Limit(1).
		Take(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find next published content item ready for twitter: %w", err)
	}
	item := toDomain(model)
	return &item, nil
}

func (r *ContentRepository) FindByID(ctx context.Context, id string) (*domain.ContentItem, error) {
	var model models.ContentItemModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find content item by id: %w", err)
	}
	item := toDomain(model)
	return &item, nil
}

func (r *ContentRepository) Save(ctx context.Context, item domain.ContentItem) error {
	model := toModel(item)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return fmt.Errorf("save content item: %w", err)
	}
	return nil
}

func (r *ContentRepository) ListByStatuses(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
	values := make([]string, 0, len(statuses))
	for _, status := range statuses {
		values = append(values, string(status))
	}

	var rows []models.ContentItemModel
	query := r.db.WithContext(ctx).Where("status IN ?", values).Order("crawled_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list content items by status: %w", err)
	}

	items := make([]domain.ContentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomain(row))
	}
	return items, nil
}

func (r *ContentRepository) ListRecent(ctx context.Context, limit int) ([]domain.ContentItem, error) {
	var rows []models.ContentItemModel
	query := r.db.WithContext(ctx).Order("crawled_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list recent content items: %w", err)
	}

	items := make([]domain.ContentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomain(row))
	}
	return items, nil
}

func toModel(item domain.ContentItem) models.ContentItemModel {
	return models.ContentItemModel{
		ID:              item.ID,
		SourceID:        item.SourceID,
		ExternalID:      item.ExternalID,
		OriginalText:    item.OriginalText,
		AuthorName:      item.AuthorName,
		SourceURL:       item.SourceURL,
		CrawledAt:       item.CrawledAt,
		Status:          string(item.Status),
		RewrittenText:   item.RewrittenText,
		RewrittenTextEn: item.RewrittenTextEn,
		TweetTextVI:     item.TweetTextVI,
		TweetTextEN:     item.TweetTextEN,
		FactCheckNote:   item.FactCheckNote,
		FailReason:      item.FailReason,
		TweetViID:       item.TweetViID,
		TweetEnID:       item.TweetEnID,
		PublishedAt:     item.PublishedAt,
		PublishedMsgID:  item.PublishedMsgID,
	}
}

func toDomain(model models.ContentItemModel) domain.ContentItem {
	return domain.ContentItem{
		ID:              model.ID,
		SourceID:        model.SourceID,
		ExternalID:      model.ExternalID,
		OriginalText:    model.OriginalText,
		AuthorName:      model.AuthorName,
		SourceURL:       model.SourceURL,
		CrawledAt:       model.CrawledAt,
		Status:          domain.Status(model.Status),
		RewrittenText:   model.RewrittenText,
		RewrittenTextEn: model.RewrittenTextEn,
		TweetTextVI:     model.TweetTextVI,
		TweetTextEN:     model.TweetTextEN,
		FactCheckNote:   model.FactCheckNote,
		FailReason:      model.FailReason,
		TweetViID:       model.TweetViID,
		TweetEnID:       model.TweetEnID,
		PublishedAt:     model.PublishedAt,
		PublishedMsgID:  model.PublishedMsgID,
	}
}

func isDuplicateErr(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
