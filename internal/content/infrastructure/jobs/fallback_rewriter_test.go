package jobs

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go-content-bot/internal/content/application"
)

type namedRewriteStub struct {
	rewriteFn func(ctx context.Context, originalText string) (application.RewriteResult, error)
}

func (s namedRewriteStub) RewriteText(ctx context.Context, originalText string) (application.RewriteResult, error) {
	return s.rewriteFn(ctx, originalText)
}

func TestFallbackRewriterUsesPrimaryWhenItSucceeds(t *testing.T) {
	t.Parallel()

	secondaryCalled := false
	rewriter := NewFallbackRewriter(
		"gemini",
		namedRewriteStub{
			rewriteFn: func(ctx context.Context, originalText string) (application.RewriteResult, error) {
				return application.PlainRewriteResult("primary result"), nil
			},
		},
		"deepseek",
		namedRewriteStub{
			rewriteFn: func(ctx context.Context, originalText string) (application.RewriteResult, error) {
				secondaryCalled = true
				return application.PlainRewriteResult("secondary result"), nil
			},
		},
	)

	result, err := rewriter.RewriteText(context.Background(), "original")
	if err != nil {
		t.Fatalf("rewrite text: %v", err)
	}
	if result.RewrittenText != "primary result" {
		t.Fatalf("expected primary result, got %s", result.RewrittenText)
	}
	if secondaryCalled {
		t.Fatal("expected secondary not to be called")
	}
}

func TestFallbackRewriterFallsBackToSecondaryOnPrimaryError(t *testing.T) {
	t.Parallel()

	rewriter := NewFallbackRewriter(
		"gemini",
		namedRewriteStub{
			rewriteFn: func(ctx context.Context, originalText string) (application.RewriteResult, error) {
				return application.RewriteResult{}, errors.New("429 quota exceeded")
			},
		},
		"deepseek",
		namedRewriteStub{
			rewriteFn: func(ctx context.Context, originalText string) (application.RewriteResult, error) {
				return application.PlainRewriteResult("secondary result"), nil
			},
		},
	)

	result, err := rewriter.RewriteText(context.Background(), "original")
	if err != nil {
		t.Fatalf("rewrite text with fallback: %v", err)
	}
	if result.RewrittenText != "secondary result" {
		t.Fatalf("expected secondary result, got %s", result.RewrittenText)
	}
}

func TestFallbackRewriterReturnsCombinedErrorWhenBothFail(t *testing.T) {
	t.Parallel()

	rewriter := NewFallbackRewriter(
		"gemini",
		namedRewriteStub{
			rewriteFn: func(ctx context.Context, originalText string) (application.RewriteResult, error) {
				return application.RewriteResult{}, errors.New("gemini unavailable")
			},
		},
		"deepseek",
		namedRewriteStub{
			rewriteFn: func(ctx context.Context, originalText string) (application.RewriteResult, error) {
				return application.RewriteResult{}, errors.New("deepseek 402")
			},
		},
	)

	_, err := rewriter.RewriteText(context.Background(), "original")
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !strings.Contains(err.Error(), "gemini") || !strings.Contains(err.Error(), "deepseek") {
		t.Fatalf("expected error to mention both providers, got %v", err)
	}
}
