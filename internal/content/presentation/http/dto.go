package contenthttp

import "go-content-bot/internal/content/domain"

type contentItemResponse struct {
	ID              string  `json:"id"`
	SourceID        *string `json:"source_id"`
	ExternalID      string  `json:"external_id"`
	OriginalText    string  `json:"original_text"`
	AuthorName      string  `json:"author_name"`
	SourceURL       *string `json:"source_url"`
	CrawledAt       string  `json:"crawled_at"`
	Status          string  `json:"status"`
	RewrittenText   *string `json:"rewritten_text"`
	RewrittenTextEn *string `json:"rewritten_text_en"`
	TweetTextVI     *string `json:"tweet_text_vi"`
	TweetTextEN     *string `json:"tweet_text_en"`
	FactCheckNote   *string `json:"fact_check_note"`
	FailReason      *string `json:"fail_reason"`
	TweetViID       *string `json:"tweet_vi_id"`
	TweetEnID       *string `json:"tweet_en_id"`
	PublishedAt     *string `json:"published_at"`
	PublishedMsgID  *string `json:"published_msg_id"`
}

func toContentItemResponse(item domain.ContentItem) contentItemResponse {
	var publishedAt *string
	if item.PublishedAt != nil {
		value := item.PublishedAt.Format(timeLayout)
		publishedAt = &value
	}

	return contentItemResponse{
		ID:              item.ID,
		SourceID:        item.SourceID,
		ExternalID:      item.ExternalID,
		OriginalText:    item.OriginalText,
		AuthorName:      item.AuthorName,
		SourceURL:       item.SourceURL,
		CrawledAt:       item.CrawledAt.Format(timeLayout),
		Status:          string(item.Status),
		RewrittenText:   item.RewrittenText,
		RewrittenTextEn: item.RewrittenTextEn,
		TweetTextVI:     item.TweetTextVI,
		TweetTextEN:     item.TweetTextEN,
		FactCheckNote:   item.FactCheckNote,
		FailReason:      item.FailReason,
		TweetViID:       item.TweetViID,
		TweetEnID:       item.TweetEnID,
		PublishedAt:     publishedAt,
		PublishedMsgID:  item.PublishedMsgID,
	}
}

func toContentItemResponses(items []domain.ContentItem) []contentItemResponse {
	responses := make([]contentItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toContentItemResponse(item))
	}
	return responses
}
