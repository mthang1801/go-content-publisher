package repositories

import (
	"context"
	"fmt"

	"go-content-bot/internal/system/infrastructure/persistence/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettingRepository struct {
	db *gorm.DB
}

type SettingRecord struct {
	Key         string
	Value       string
	JSONValue   []byte
	Description *string
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) GetRecord(ctx context.Context, key string) (SettingRecord, error) {
	var model models.SettingModel
	result := r.db.WithContext(ctx).Where("key = ?", key).Limit(1).Find(&model)
	if result.Error != nil {
		return SettingRecord{}, fmt.Errorf("get setting record %s: %w", key, result.Error)
	}
	if result.RowsAffected == 0 {
		return SettingRecord{}, nil
	}
	return SettingRecord{
		Key:         model.Key,
		Value:       model.Value,
		JSONValue:   append([]byte(nil), model.JSONValue...),
		Description: model.Description,
	}, nil
}

func (r *SettingRepository) Get(ctx context.Context, key string) (string, error) {
	record, err := r.GetRecord(ctx, key)
	if err != nil {
		return "", err
	}
	if record.Key == "" {
		return "", nil
	}
	return record.Value, nil
}

func (r *SettingRepository) Set(ctx context.Context, key, value string) error {
	model := models.SettingModel{
		Key:   key,
		Value: value,
	}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).
		Create(&model).Error; err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

func (r *SettingRepository) SetJSON(ctx context.Context, key string, jsonValue []byte) error {
	model := models.SettingModel{
		Key:       key,
		JSONValue: append([]byte(nil), jsonValue...),
	}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"json_value"}),
		}).
		Create(&model).Error; err != nil {
		return fmt.Errorf("set setting json %s: %w", key, err)
	}
	return nil
}
