package models

import "time"

import "github.com/lib/pq"

type SourceModel struct {
	ID            string         `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Type          string         `gorm:"column:type"`
	Handle        string         `gorm:"column:handle"`
	Name          string         `gorm:"column:name"`
	Tags          pq.StringArray `gorm:"column:tags;type:text[]"`
	Topics        pq.StringArray `gorm:"column:topics;type:text[]"`
	IsActive      bool           `gorm:"column:is_active"`
	LastCrawledAt *time.Time     `gorm:"column:last_crawled_at"`
	LastCheckAt   *time.Time     `gorm:"column:last_check_at"`
	LastError     *string        `gorm:"column:last_error"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
}

func (SourceModel) TableName() string {
	return "sources"
}
