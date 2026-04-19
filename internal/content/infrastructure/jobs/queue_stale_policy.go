package jobs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	contentapp "go-content-bot/internal/content/application"
)

const pendingStaleAfterSecondsSettingKey = "pending_stale_after_seconds"
const rewrittenStaleAfterSecondsSettingKey = "rewritten_stale_after_seconds"

const stalePendingReason = "stale pending item exceeded pending_stale_after_seconds"
const staleRewrittenReason = "stale rewritten item exceeded rewritten_stale_after_seconds"

type queueSettingsStore interface {
	Get(ctx context.Context, key string) (string, error)
}

func resolveQueueStaleAfter(ctx context.Context, settings queueSettingsStore, primaryKey string) (time.Duration, error) {
	if settings == nil {
		return 0, nil
	}
	value, err := settings.Get(ctx, primaryKey)
	if err != nil {
		return 0, err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", primaryKey, err)
	}
	if seconds <= 0 {
		return 0, nil
	}
	return time.Duration(seconds) * time.Second, nil
}

func skipStalePending(ctx context.Context, service *contentapp.Service, settings queueSettingsStore) (int64, error) {
	staleAfter, err := resolveQueueStaleAfter(ctx, settings, pendingStaleAfterSecondsSettingKey)
	if err != nil || staleAfter <= 0 {
		return 0, err
	}
	return service.SkipStalePendingBefore(ctx, time.Now().Add(-staleAfter), stalePendingReason)
}

func skipStaleRewritten(ctx context.Context, service *contentapp.Service, settings queueSettingsStore) (int64, error) {
	staleAfter, err := resolveQueueStaleAfter(ctx, settings, rewrittenStaleAfterSecondsSettingKey)
	if err != nil || staleAfter <= 0 {
		return 0, err
	}
	return service.SkipStaleRewrittenBefore(ctx, time.Now().Add(-staleAfter), staleRewrittenReason)
}

func SkipStalePendingForRun(ctx context.Context, service *contentapp.Service, settings queueSettingsStore) (int64, error) {
	return skipStalePending(ctx, service, settings)
}

func SkipStaleRewrittenForRun(ctx context.Context, service *contentapp.Service, settings queueSettingsStore) (int64, error) {
	return skipStaleRewritten(ctx, service, settings)
}
