package jobs

import (
	"context"
	"fmt"
	"strings"

	"go-content-bot/internal/content/application"
)

type FallbackRewriter struct {
	primaryName   string
	primary       application.RewriteTextPort
	secondaryName string
	secondary     application.RewriteTextPort
}

func NewFallbackRewriter(
	primaryName string,
	primary application.RewriteTextPort,
	secondaryName string,
	secondary application.RewriteTextPort,
) *FallbackRewriter {
	return &FallbackRewriter{
		primaryName:   strings.TrimSpace(primaryName),
		primary:       primary,
		secondaryName: strings.TrimSpace(secondaryName),
		secondary:     secondary,
	}
}

func (r *FallbackRewriter) RewriteText(ctx context.Context, originalText string) (application.RewriteResult, error) {
	result, primaryErr := r.primary.RewriteText(ctx, originalText)
	if primaryErr == nil {
		return result, nil
	}
	if r.secondary == nil {
		return application.RewriteResult{}, primaryErr
	}

	result, secondaryErr := r.secondary.RewriteText(ctx, originalText)
	if secondaryErr == nil {
		return result, nil
	}

	return application.RewriteResult{}, fmt.Errorf(
		"%s rewrite failed: %v; %s fallback failed: %v",
		defaultRewriteName(r.primaryName, "primary"),
		primaryErr,
		defaultRewriteName(r.secondaryName, "secondary"),
		secondaryErr,
	)
}

func defaultRewriteName(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
