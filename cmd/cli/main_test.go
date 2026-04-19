package main

import (
	"testing"
	"time"

	contentdomain "go-content-bot/internal/content/domain"
)

func TestBuildContentOpsReportGroupsRelevantStatuses(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 19, 12, 40, 0, 0, time.UTC)
	publishedAt := now.Add(5 * time.Minute)
	duplicateReason := "duplicate manual content already processed recently"

	report := buildContentOpsReport([]contentdomain.ContentItem{
		{
			ID:           "rewritten-1",
			Status:       contentdomain.StatusRewritten,
			AuthorName:   "A",
			OriginalText: "rewritten item",
			CrawledAt:    now,
		},
		{
			ID:           "skipped-1",
			Status:       contentdomain.StatusSkipped,
			AuthorName:   "B",
			OriginalText: "duplicate item",
			FailReason:   &duplicateReason,
			CrawledAt:    now,
		},
		{
			ID:           "published-1",
			Status:       contentdomain.StatusPublished,
			AuthorName:   "C",
			OriginalText: "published item",
			CrawledAt:    now,
			PublishedAt:  &publishedAt,
		},
		{
			ID:           "skipped-other",
			Status:       contentdomain.StatusSkipped,
			AuthorName:   "D",
			OriginalText: "other skipped item",
			CrawledAt:    now,
		},
	})

	if report.Counts["rewritten_ready"] != 1 {
		t.Fatalf("expected 1 rewritten item, got %d", report.Counts["rewritten_ready"])
	}
	if report.Counts["manual_duplicate_skipped"] != 1 {
		t.Fatalf("expected 1 manual duplicate item, got %d", report.Counts["manual_duplicate_skipped"])
	}
	if report.Counts["published_recent"] != 1 {
		t.Fatalf("expected 1 published item, got %d", report.Counts["published_recent"])
	}
	if len(report.RewrittenReady) != 1 || report.RewrittenReady[0].ID != "rewritten-1" {
		t.Fatalf("unexpected rewritten group: %#v", report.RewrittenReady)
	}
	if len(report.ManualDuplicateSkipped) != 1 || report.ManualDuplicateSkipped[0].ID != "skipped-1" {
		t.Fatalf("unexpected duplicate group: %#v", report.ManualDuplicateSkipped)
	}
	if len(report.PublishedRecent) != 1 || report.PublishedRecent[0].ID != "published-1" {
		t.Fatalf("unexpected published group: %#v", report.PublishedRecent)
	}
}

func TestParseContentOpsLimit(t *testing.T) {
	t.Parallel()

	if got := parseContentOpsLimit([]string{"--limit=25"}, 50); got != 25 {
		t.Fatalf("expected limit 25, got %d", got)
	}
	if got := parseContentOpsLimit([]string{"--limit=0"}, 50); got != 50 {
		t.Fatalf("expected fallback limit 50, got %d", got)
	}
	if got := parseContentOpsLimit([]string{"--other=1"}, 50); got != 50 {
		t.Fatalf("expected fallback limit 50, got %d", got)
	}
}
