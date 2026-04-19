package sourcehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-content-bot/internal/source/application"
	"go-content-bot/internal/source/domain"

	"github.com/gin-gonic/gin"
)

type sourceRepoStub struct {
	createFn          func(ctx context.Context, source domain.Source) (domain.Source, error)
	listAllFn         func(ctx context.Context) ([]domain.Source, error)
	listFn            func(ctx context.Context) ([]domain.Source, error)
	listActiveDueFn   func(ctx context.Context, limit int) ([]domain.Source, error)
	listInactiveDueFn func(ctx context.Context, now time.Time, limit int) ([]domain.Source, error)
	touchCrawlFn      func(ctx context.Context, id string, at time.Time) error
	markCheckedFn     func(ctx context.Context, id string, at time.Time) error
	markInactiveFn    func(ctx context.Context, id string, reason string, at time.Time) error
	markActiveFn      func(ctx context.Context, id string, at time.Time) error
	updateMetadataFn  func(ctx context.Context, sourceType domain.Type, handle string, tags []string, topics []string) error
	deleteFn          func(ctx context.Context, sourceType domain.Type, handle string) error
}

func (s sourceRepoStub) Create(ctx context.Context, source domain.Source) (domain.Source, error) {
	return s.createFn(ctx, source)
}

func (s sourceRepoStub) ListActive(ctx context.Context) ([]domain.Source, error) {
	return s.listFn(ctx)
}

func (s sourceRepoStub) ListAll(ctx context.Context) ([]domain.Source, error) {
	if s.listAllFn == nil {
		return nil, nil
	}
	return s.listAllFn(ctx)
}

func (s sourceRepoStub) ListActiveDueForValidation(ctx context.Context, limit int) ([]domain.Source, error) {
	if s.listActiveDueFn == nil {
		return nil, nil
	}
	return s.listActiveDueFn(ctx, limit)
}

func (s sourceRepoStub) ListInactiveDueForRecheck(ctx context.Context, now time.Time, limit int) ([]domain.Source, error) {
	if s.listInactiveDueFn == nil {
		return nil, nil
	}
	return s.listInactiveDueFn(ctx, now, limit)
}

func (s sourceRepoStub) TouchCrawl(ctx context.Context, id string, at time.Time) error {
	if s.touchCrawlFn == nil {
		return nil
	}
	return s.touchCrawlFn(ctx, id, at)
}

func (s sourceRepoStub) MarkChecked(ctx context.Context, id string, at time.Time) error {
	if s.markCheckedFn == nil {
		return nil
	}
	return s.markCheckedFn(ctx, id, at)
}

func (s sourceRepoStub) MarkInactive(ctx context.Context, id string, reason string, at time.Time) error {
	if s.markInactiveFn == nil {
		return nil
	}
	return s.markInactiveFn(ctx, id, reason, at)
}

func (s sourceRepoStub) MarkActive(ctx context.Context, id string, at time.Time) error {
	if s.markActiveFn == nil {
		return nil
	}
	return s.markActiveFn(ctx, id, at)
}

func (s sourceRepoStub) UpdateMetadataByHandle(ctx context.Context, sourceType domain.Type, handle string, tags []string, topics []string) error {
	if s.updateMetadataFn == nil {
		return nil
	}
	return s.updateMetadataFn(ctx, sourceType, handle, tags, topics)
}

func (s sourceRepoStub) DeleteByHandle(ctx context.Context, sourceType domain.Type, handle string) error {
	return s.deleteFn(ctx, sourceType, handle)
}

func TestCreateReturnsAPIDTOShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := sourceRepoStub{
		createFn: func(ctx context.Context, source domain.Source) (domain.Source, error) {
			return domain.Source{
				ID:        "src-1",
				Type:      domain.TypeTelegram,
				Handle:    source.Handle,
				Name:      source.Name,
				Tags:      source.Tags,
				Topics:    source.Topics,
				IsActive:  true,
				CreatedAt: time.Date(2026, 4, 18, 21, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	handler := NewHandler(application.NewService(repo))
	router := gin.New()
	router.POST("/api/sources", handler.Create)

	body := bytes.NewBufferString(`{"type":"telegram","handle":"@demo","name":"Demo","tags":["macro"],"topics":["markets"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sources", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if _, ok := payload["ID"]; ok {
		t.Fatalf("expected no exported Go field names in response, got %s", rec.Body.String())
	}
	if got := payload["id"]; got != "src-1" {
		t.Fatalf("expected id src-1, got %#v", got)
	}
	if got := payload["handle"]; got != "@demo" {
		t.Fatalf("expected handle @demo, got %#v", got)
	}
	if got := payload["tags"]; len(got.([]any)) != 1 {
		t.Fatalf("expected 1 tag, got %#v", got)
	}
	if _, ok := payload["created_at"]; !ok {
		t.Fatalf("expected created_at field, got %s", rec.Body.String())
	}
}

func TestListReturnsItemsWithAPIDTOShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := sourceRepoStub{
		listFn: func(ctx context.Context) ([]domain.Source, error) {
			return []domain.Source{
				{
					ID:        "src-2",
					Type:      domain.TypeTwitter,
					Handle:    "@writer",
					Name:      "Writer",
					Tags:      []string{"markets"},
					Topics:    []string{"macro"},
					IsActive:  true,
					CreatedAt: time.Date(2026, 4, 18, 21, 5, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	handler := NewHandler(application.NewService(repo))
	router := gin.New()
	router.GET("/api/sources", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/api/sources", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}
	if _, ok := payload.Items[0]["Handle"]; ok {
		t.Fatalf("expected no exported Go field names in list response, got %s", rec.Body.String())
	}
	if got := payload.Items[0]["type"]; got != "twitter" {
		t.Fatalf("expected type twitter, got %#v", got)
	}
	if got := payload.Items[0]["tags"]; len(got.([]any)) != 1 {
		t.Fatalf("expected tags in list, got %#v", got)
	}
}

func TestReportReturnsInactiveItemsAndMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	lastCheck := time.Date(2026, 4, 19, 5, 29, 48, 0, time.UTC)
	lastError := "Bad Request: chat not found"
	repo := sourceRepoStub{
		listAllFn: func(ctx context.Context) ([]domain.Source, error) {
			return []domain.Source{
				{
					ID:          "src-3",
					Type:        domain.TypeTelegram,
					Handle:      "@missing",
					Name:        "Missing",
					Tags:        []string{"geopolitics"},
					Topics:      []string{"military"},
					IsActive:    false,
					LastCheckAt: &lastCheck,
					LastError:   &lastError,
					CreatedAt:   time.Date(2026, 4, 18, 21, 5, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	handler := NewHandler(application.NewService(repo))
	router := gin.New()
	router.GET("/api/sources/report", handler.Report)

	req := httptest.NewRequest(http.MethodGet, "/api/sources/report?type=telegram", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}
	if got := payload.Items[0]["is_active"]; got != false {
		t.Fatalf("expected inactive source, got %#v", got)
	}
	if got := payload.Items[0]["last_error"]; got != lastError {
		t.Fatalf("expected last_error %q, got %#v", lastError, got)
	}
}

func TestUpdateMetadataReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := sourceRepoStub{
		updateMetadataFn: func(ctx context.Context, sourceType domain.Type, handle string, tags []string, topics []string) error {
			if sourceType != domain.TypeTwitter {
				t.Fatalf("expected twitter type, got %s", sourceType)
			}
			if handle != "@writer" {
				t.Fatalf("expected @writer handle, got %s", handle)
			}
			if len(tags) != 1 || tags[0] != "markets" {
				t.Fatalf("expected markets tag, got %#v", tags)
			}
			if len(topics) != 1 || topics[0] != "macro" {
				t.Fatalf("expected macro topic, got %#v", topics)
			}
			return nil
		},
	}

	handler := NewHandler(application.NewService(repo))
	router := gin.New()
	router.PATCH("/api/sources/:type/:handle", handler.UpdateMetadata)

	body := bytes.NewBufferString(`{"tags":["markets"],"topics":["macro"]}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/sources/twitter/@writer", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}
