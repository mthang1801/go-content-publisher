package repositories

import (
	"context"
	"fmt"
	"time"

	"go-content-bot/internal/system/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type LogRecord struct {
	ID        string
	Level     string
	Module    string
	Message   string
	CreatedAt time.Time
}

type LogRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) ListRecent(ctx context.Context, limit int) ([]LogRecord, error) {
	var rows []models.LogModel
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list recent logs: %w", err)
	}

	records := make([]LogRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, LogRecord{
			ID:        row.ID,
			Level:     row.Level,
			Module:    row.Module,
			Message:   row.Message,
			CreatedAt: row.CreatedAt,
		})
	}
	return records, nil
}
