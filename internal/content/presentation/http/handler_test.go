package contenthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-content-bot/internal/content/application"
	"go-content-bot/internal/content/domain"

	"github.com/gin-gonic/gin"
)

type contentRepoStub struct {
	createPendingFn  func(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error)
	findByIDFn       func(ctx context.Context, id string) (*domain.ContentItem, error)
	listByStatusesFn func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error)
	listRecentFn     func(ctx context.Context, limit int) ([]domain.ContentItem, error)
	saveFn           func(ctx context.Context, item domain.ContentItem) error
}

func (s contentRepoStub) CreatePending(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error) {
	return s.createPendingFn(ctx, item)
}

func (s contentRepoStub) SkipStalePending(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return 0, nil
}

func (s contentRepoStub) SkipStaleRewritten(ctx context.Context, staleBefore time.Time, reason string) (int64, error) {
	return 0, nil
}

func (s contentRepoStub) ClaimNextPending(ctx context.Context) (*domain.ContentItem, error) {
	return nil, nil
}

func (s contentRepoStub) ClaimNextReadyForPublish(ctx context.Context) (*domain.ContentItem, error) {
	return nil, nil
}

func (s contentRepoStub) FindNextPublishedReadyForTwitter(ctx context.Context, publishAfter *time.Time, sourceTypes []string, sourceTags []string, sourceTopics []string, topicKeywords []string) (*domain.ContentItem, error) {
	return nil, nil
}

func (s contentRepoStub) FindNextPending(ctx context.Context) (*domain.ContentItem, error) {
	return nil, nil
}

func (s contentRepoStub) FindByID(ctx context.Context, id string) (*domain.ContentItem, error) {
	return s.findByIDFn(ctx, id)
}

func (s contentRepoStub) Save(ctx context.Context, item domain.ContentItem) error {
	return s.saveFn(ctx, item)
}

func (s contentRepoStub) ListByStatuses(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
	return s.listByStatusesFn(ctx, statuses, limit)
}

func (s contentRepoStub) ListRecent(ctx context.Context, limit int) ([]domain.ContentItem, error) {
	return s.listRecentFn(ctx, limit)
}

func TestQueueReturnsAPIDTOShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := contentRepoStub{
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			return []domain.ContentItem{
				{
					ID:           "item-1",
					ExternalID:   "external-1",
					OriginalText: "hello",
					AuthorName:   "author",
					CrawledAt:    time.Date(2026, 4, 18, 21, 10, 0, 0, time.UTC),
					Status:       domain.StatusPending,
				},
			}, nil
		},
		listRecentFn: func(ctx context.Context, limit int) ([]domain.ContentItem, error) {
			return nil, nil
		},
	}

	handler := NewHandler(application.NewService(repo))
	router := gin.New()
	router.GET("/api/content/queue", handler.Queue)

	req := httptest.NewRequest(http.MethodGet, "/api/content/queue", nil)
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
	if _, ok := payload.Items[0]["OriginalText"]; ok {
		t.Fatalf("expected no exported Go field names in queue response, got %s", rec.Body.String())
	}
	if got := payload.Items[0]["original_text"]; got != "hello" {
		t.Fatalf("expected original_text hello, got %#v", got)
	}
	if got := payload.Items[0]["status"]; got != "pending" {
		t.Fatalf("expected status pending, got %#v", got)
	}
}

func TestCreateReturnsAPIDTOShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := contentRepoStub{
		createPendingFn: func(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error) {
			item.ID = "item-10"
			item.CrawledAt = time.Date(2026, 4, 18, 21, 20, 0, 0, time.UTC)
			return item, nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			return nil, nil
		},
		listRecentFn: func(ctx context.Context, limit int) ([]domain.ContentItem, error) {
			return nil, nil
		},
	}

	handler := NewHandler(application.NewService(repo))
	router := gin.New()
	router.POST("/api/content", handler.Create)

	req := httptest.NewRequest(http.MethodPost, "/api/content", strings.NewReader(`{"text":"hello from api","author":"tester"}`))
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

	if got := payload["id"]; got != "item-10" {
		t.Fatalf("expected id item-10, got %#v", got)
	}
	if got := payload["original_text"]; got != "hello from api" {
		t.Fatalf("expected original_text hello from api, got %#v", got)
	}
	if got := payload["author_name"]; got != "tester" {
		t.Fatalf("expected author_name tester, got %#v", got)
	}
	if _, ok := payload["OriginalText"]; ok {
		t.Fatalf("expected no exported Go field names in response, got %s", rec.Body.String())
	}
}

func TestManualRewriteReturnsUpdatedDTOShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := contentRepoStub{
		createPendingFn: func(ctx context.Context, item domain.ContentItem) (domain.ContentItem, error) {
			return domain.ContentItem{}, nil
		},
		findByIDFn: func(ctx context.Context, id string) (*domain.ContentItem, error) {
			item := domain.ContentItem{
				ID:           id,
				OriginalText: "hello from api",
				AuthorName:   "tester",
				CrawledAt:    time.Date(2026, 4, 18, 21, 22, 0, 0, time.UTC),
				Status:       domain.StatusFailed,
			}
			return &item, nil
		},
		saveFn: func(ctx context.Context, item domain.ContentItem) error {
			return nil
		},
		listByStatusesFn: func(ctx context.Context, statuses []domain.Status, limit int) ([]domain.ContentItem, error) {
			return nil, nil
		},
		listRecentFn: func(ctx context.Context, limit int) ([]domain.ContentItem, error) {
			return nil, nil
		},
	}

	handler := NewHandler(application.NewService(repo))
	router := gin.New()
	router.POST("/api/content/:id/manual-rewrite", handler.ManualRewrite)

	req := httptest.NewRequest(http.MethodPost, "/api/content/item-11/manual-rewrite", strings.NewReader(`{"rewritten_text":"manual rewritten text"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got := payload["status"]; got != "rewritten" {
		t.Fatalf("expected status rewritten, got %#v", got)
	}
	if got := payload["rewritten_text"]; got != "manual rewritten text" {
		t.Fatalf("expected rewritten_text manual rewritten text, got %#v", got)
	}
}
