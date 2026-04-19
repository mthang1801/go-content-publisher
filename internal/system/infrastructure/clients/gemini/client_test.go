package gemini

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckListsModels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Fatalf("expected x-goog-api-key header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"models/gemini-2.5-flash"},{"name":"models/gemini-2.5-pro"}]}`))
	}))
	defer server.Close()

	client := New("test-key", "gemini-2.5-flash")
	client.baseURL = server.URL + "/v1beta"

	result, err := client.Check(context.Background())
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %s", result.Status)
	}
	if result.ModelCount != 2 {
		t.Fatalf("expected 2 models, got %d", result.ModelCount)
	}
}

func TestRewriteTextUsesGenerateContent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-2.5-flash:generateContent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(string(body), "Original source text") {
			t.Fatalf("expected original text in request body, got %s", string(body))
		}
		if !strings.Contains(string(body), "responseMimeType") {
			t.Fatalf("expected response mime type in request body, got %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"rewrittenText\":\"Ban rewrite tu Gemini\",\"rewrittenTextEn\":\"English rewrite\",\"tweetVI\":\"Tweet VI\",\"tweetEN\":\"Tweet EN\",\"factCheckNote\":\"OK\",\"shouldPublish\":true}"}]}}]}`))
	}))
	defer server.Close()

	client := New("test-key", "gemini-2.5-flash")
	client.baseURL = server.URL + "/v1beta"

	result, err := client.RewriteText(context.Background(), "Original source text")
	if err != nil {
		t.Fatalf("rewrite text: %v", err)
	}
	if result.RewrittenText != "Ban rewrite tu Gemini" {
		t.Fatalf("expected rewritten text, got %s", result.RewrittenText)
	}
	if result.TweetTextVI != "Tweet VI" {
		t.Fatalf("expected tweet vi, got %s", result.TweetTextVI)
	}
}
