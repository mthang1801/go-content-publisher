package telegrambot

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSendMessageSanitizesInvalidUTF8Text(t *testing.T) {
	t.Parallel()

	captured := url.Values{}
	client := New("bot-token")
	client.baseURL = "https://example.invalid"
	client.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			values, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parse query: %v", err)
			}
			captured = values
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":123}}`)),
			}, nil
		}),
	}

	messageID, err := client.SendMessage(context.Background(), "-100123", nil, "Xin chào\xff thế giới")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if messageID != "123" {
		t.Fatalf("expected message id 123, got %s", messageID)
	}
	if got := captured.Get("text"); got != "Xin chào thế giới" {
		t.Fatalf("expected sanitized text, got %q", got)
	}
}

func TestSanitizeTelegramTextKeepsValidUTF8Trimmed(t *testing.T) {
	t.Parallel()

	if got := sanitizeTelegramText("  Xin chào Telegram  "); got != "Xin chào Telegram" {
		t.Fatalf("expected trimmed valid utf8 text, got %q", got)
	}
	if got := sanitizeTelegramText(strings.Repeat("a", 3)); got != "aaa" {
		t.Fatalf("expected unchanged ascii text, got %q", got)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
