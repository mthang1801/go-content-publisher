package domain

import (
	"errors"
	"testing"
	"time"
)

func TestContentItemTransitions(t *testing.T) {
	item := ContentItem{
		OriginalText: "hello",
		Status:       StatusPending,
	}

	if err := item.MarkProcessing(); err != nil {
		t.Fatalf("mark processing: %v", err)
	}
	if err := item.MarkRewritten("rewritten"); err != nil {
		t.Fatalf("mark rewritten: %v", err)
	}
	if err := item.MarkPublishing(); err != nil {
		t.Fatalf("mark publishing: %v", err)
	}
	if err := item.MarkPublished("123", time.Now()); err != nil {
		t.Fatalf("mark published: %v", err)
	}
}

func TestContentItemRejectsInvalidTransition(t *testing.T) {
	item := ContentItem{OriginalText: "hello", Status: StatusPending}
	if err := item.MarkPublished("123", time.Now()); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
}

func TestContentItemAllowsManualRewriteFromFailed(t *testing.T) {
	item := ContentItem{OriginalText: "hello", Status: StatusFailed}
	if err := item.SetManualRewrite("rewritten manually"); err != nil {
		t.Fatalf("set manual rewrite: %v", err)
	}
	if item.Status != StatusRewritten {
		t.Fatalf("expected rewritten status, got %s", item.Status)
	}
}
