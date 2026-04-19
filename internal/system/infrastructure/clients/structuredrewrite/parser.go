package structuredrewrite

import (
	"encoding/json"
	"fmt"
	"strings"

	contentapp "go-content-bot/internal/content/application"
)

type responsePayload struct {
	RewrittenText   string `json:"rewrittenText"`
	RewrittenTextEn string `json:"rewrittenTextEn"`
	TweetVI         string `json:"tweetVI"`
	TweetEN         string `json:"tweetEN"`
	FactCheckNote   string `json:"factCheckNote"`
	ShouldPublish   *bool  `json:"shouldPublish"`
	Reason          string `json:"reason"`
}

func Parse(text string) (contentapp.RewriteResult, error) {
	raw := strings.TrimSpace(text)
	if raw == "" {
		return contentapp.RewriteResult{}, fmt.Errorf("empty rewrite response")
	}

	var payload responsePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		extracted, extractErr := extractJSONObject(raw)
		if extractErr != nil {
			return contentapp.RewriteResult{}, fmt.Errorf("parse structured rewrite response: %w", err)
		}
		if err := json.Unmarshal([]byte(extracted), &payload); err != nil {
			return contentapp.RewriteResult{}, fmt.Errorf("parse extracted structured rewrite response: %w", err)
		}
	}

	shouldPublish := true
	if payload.ShouldPublish != nil {
		shouldPublish = *payload.ShouldPublish
	}

	result := contentapp.RewriteResult{
		RewrittenText:   strings.TrimSpace(payload.RewrittenText),
		RewrittenTextEn: strings.TrimSpace(payload.RewrittenTextEn),
		TweetTextVI:     strings.TrimSpace(payload.TweetVI),
		TweetTextEN:     strings.TrimSpace(payload.TweetEN),
		FactCheckNote:   strings.TrimSpace(payload.FactCheckNote),
		ShouldPublish:   shouldPublish,
		Reason:          strings.TrimSpace(payload.Reason),
	}

	if result.ShouldPublish && result.RewrittenText == "" {
		return contentapp.RewriteResult{}, fmt.Errorf("structured rewrite response missing rewrittenText")
	}

	return result, nil
}

func extractJSONObject(text string) (string, error) {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end < 0 || end <= start {
		return "", fmt.Errorf("json object not found")
	}
	return text[start : end+1], nil
}
