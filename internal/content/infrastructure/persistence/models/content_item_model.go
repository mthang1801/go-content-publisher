package models

import "time"

type ContentItemModel struct {
	ID              string     `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	SourceID        *string    `gorm:"column:source_id;type:uuid"`
	ExternalID      string     `gorm:"column:external_id"`
	OriginalText    string     `gorm:"column:original_text"`
	AuthorName      string     `gorm:"column:author_name"`
	SourceURL       *string    `gorm:"column:source_url"`
	CrawledAt       time.Time  `gorm:"column:crawled_at"`
	Status          string     `gorm:"column:status"`
	RewrittenText   *string    `gorm:"column:rewritten_text"`
	RewrittenTextEn *string    `gorm:"column:rewritten_text_en"`
	TweetTextVI     *string    `gorm:"column:tweet_text_vi"`
	TweetTextEN     *string    `gorm:"column:tweet_text_en"`
	FactCheckNote   *string    `gorm:"column:fact_check_note"`
	FailReason      *string    `gorm:"column:fail_reason"`
	TweetViID       *string    `gorm:"column:tweet_vi_id"`
	TweetEnID       *string    `gorm:"column:tweet_en_id"`
	PublishedAt     *time.Time `gorm:"column:published_at"`
	PublishedMsgID  *string    `gorm:"column:published_msg_id"`
}

func (ContentItemModel) TableName() string {
	return "content_items"
}
