package sourcehttp

import (
	"time"

	"go-content-bot/internal/source/domain"
)

type sourceResponse struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	Handle        string   `json:"handle"`
	Name          string   `json:"name"`
	Tags          []string `json:"tags,omitempty"`
	Topics        []string `json:"topics,omitempty"`
	IsActive      bool     `json:"is_active"`
	LastCrawledAt *string  `json:"last_crawled_at,omitempty"`
	LastCheckAt   *string  `json:"last_check_at,omitempty"`
	LastError     *string  `json:"last_error,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

func toSourceResponse(source domain.Source) sourceResponse {
	return sourceResponse{
		ID:            source.ID,
		Type:          string(source.Type),
		Handle:        source.Handle,
		Name:          source.Name,
		Tags:          source.Tags,
		Topics:        source.Topics,
		IsActive:      source.IsActive,
		LastCrawledAt: formatSourceTime(source.LastCrawledAt),
		LastCheckAt:   formatSourceTime(source.LastCheckAt),
		LastError:     source.LastError,
		CreatedAt:     source.CreatedAt.Format(timeLayout),
	}
}

func toSourceResponses(items []domain.Source) []sourceResponse {
	responses := make([]sourceResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toSourceResponse(item))
	}
	return responses
}

func formatSourceTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(timeLayout)
	return &formatted
}
