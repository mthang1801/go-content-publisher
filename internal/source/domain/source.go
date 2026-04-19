package domain

import (
	"errors"
	"strings"
	"time"
)

type Type string

const (
	TypeTelegram Type = "telegram"
	TypeTwitter  Type = "twitter"
)

var (
	ErrInvalidSource       = errors.New("invalid source")
	ErrSourceAlreadyExists = errors.New("source already exists")
)

type Source struct {
	ID            string
	Type          Type
	Handle        string
	Name          string
	Tags          []string
	Topics        []string
	IsActive      bool
	LastCrawledAt *time.Time
	LastCheckAt   *time.Time
	LastError     *string
	CreatedAt     time.Time
}

func (s Source) Validate() error {
	if s.Type != TypeTelegram && s.Type != TypeTwitter {
		return ErrInvalidSource
	}
	if strings.TrimSpace(s.Handle) == "" || strings.TrimSpace(s.Name) == "" {
		return ErrInvalidSource
	}
	s.Tags = normalizeLabels(s.Tags)
	s.Topics = normalizeLabels(s.Topics)
	return nil
}

func normalizeLabels(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}
