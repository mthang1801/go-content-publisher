package domain

import (
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusRewritten  Status = "rewritten"
	StatusPublishing Status = "publishing"
	StatusPublished  Status = "published"
	StatusFailed     Status = "failed"
	StatusSkipped    Status = "skipped"
)

var ErrInvalidTransition = errors.New("invalid content item status transition")

type ContentItem struct {
	ID              string
	SourceID        *string
	ExternalID      string
	OriginalText    string
	AuthorName      string
	SourceURL       *string
	CrawledAt       time.Time
	Status          Status
	RewrittenText   *string
	RewrittenTextEn *string
	TweetTextVI     *string
	TweetTextEN     *string
	FactCheckNote   *string
	FailReason      *string
	TweetViID       *string
	TweetEnID       *string
	PublishedAt     *time.Time
	PublishedMsgID  *string
}

func (c *ContentItem) MarkProcessing() error {
	if c.Status != StatusPending {
		return ErrInvalidTransition
	}
	c.Status = StatusProcessing
	c.FailReason = nil
	return nil
}

func (c *ContentItem) MarkRewritten(text string) error {
	if c.Status != StatusProcessing {
		return ErrInvalidTransition
	}
	if strings.TrimSpace(text) == "" {
		return ErrInvalidTransition
	}
	c.Status = StatusRewritten
	c.RewrittenText = strPtr(text)
	return nil
}

func (c *ContentItem) MarkPublished(messageID string, publishedAt time.Time) error {
	if c.Status != StatusRewritten && c.Status != StatusPublishing {
		return ErrInvalidTransition
	}
	c.Status = StatusPublished
	c.PublishedMsgID = strPtr(messageID)
	c.PublishedAt = &publishedAt
	return nil
}

func (c *ContentItem) MarkPublishing() error {
	if c.Status != StatusRewritten {
		return ErrInvalidTransition
	}
	c.Status = StatusPublishing
	c.FailReason = nil
	return nil
}

func (c *ContentItem) MarkSkipped(reason string) error {
	if c.Status != StatusPending && c.Status != StatusProcessing && c.Status != StatusRewritten && c.Status != StatusPublishing && c.Status != StatusFailed {
		return ErrInvalidTransition
	}
	c.Status = StatusSkipped
	c.FailReason = strPtr(strings.TrimSpace(reason))
	return nil
}

func (c *ContentItem) MarkFailed(reason string) error {
	if c.Status != StatusPending && c.Status != StatusProcessing && c.Status != StatusRewritten && c.Status != StatusPublishing {
		return ErrInvalidTransition
	}
	c.Status = StatusFailed
	c.FailReason = strPtr(strings.TrimSpace(reason))
	return nil
}

func (c *ContentItem) RetryFailed() error {
	if c.Status != StatusFailed {
		return ErrInvalidTransition
	}
	c.Status = StatusPending
	c.FailReason = nil
	return nil
}

func (c *ContentItem) SetManualRewrite(text string) error {
	if c.Status == StatusPublished {
		return ErrInvalidTransition
	}
	if strings.TrimSpace(text) == "" {
		return ErrInvalidTransition
	}

	c.Status = StatusRewritten
	c.RewrittenText = strPtr(text)
	c.FailReason = nil
	return nil
}

func (c *ContentItem) SetTwitterPublishResults(tweetViID, tweetEnID string) error {
	if c.Status != StatusPublished {
		return ErrInvalidTransition
	}
	if strings.TrimSpace(tweetViID) != "" {
		c.TweetViID = strPtr(strings.TrimSpace(tweetViID))
	}
	if strings.TrimSpace(tweetEnID) != "" {
		c.TweetEnID = strPtr(strings.TrimSpace(tweetEnID))
	}
	return nil
}

func (c ContentItem) IsManual() bool {
	return c.SourceID == nil && strings.TrimSpace(c.ExternalID) == ""
}

func (c ContentItem) Validate() error {
	if strings.TrimSpace(c.OriginalText) == "" {
		return ErrInvalidTransition
	}
	switch c.Status {
	case StatusPending, StatusProcessing, StatusRewritten, StatusPublishing, StatusPublished, StatusFailed, StatusSkipped:
		return nil
	default:
		return ErrInvalidTransition
	}
}

func strPtr(value string) *string {
	return &value
}
