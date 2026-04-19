package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	contentapp "go-content-bot/internal/content/application"
)

type CompositeCrawlAction struct {
	name    string
	actions []contentapp.CrawlPort
	logger  *slog.Logger
}

func NewCompositeCrawlAction(name string, logger *slog.Logger, actions ...contentapp.CrawlPort) *CompositeCrawlAction {
	return &CompositeCrawlAction{
		name:    name,
		actions: actions,
		logger:  logger,
	}
}

func (a *CompositeCrawlAction) Name() string {
	if strings.TrimSpace(a.name) == "" {
		return "crawl"
	}
	return a.name
}

func (a *CompositeCrawlAction) Enabled() bool {
	for _, action := range a.actions {
		if action != nil && action.Enabled() {
			return true
		}
	}
	return false
}

func (a *CompositeCrawlAction) Crawl(ctx context.Context) error {
	var errs []error
	for _, action := range a.actions {
		if action == nil || !action.Enabled() {
			continue
		}
		if err := action.Crawl(ctx); err != nil {
			if a.logger != nil {
				a.logger.Warn("composite crawl subaction failed", "action", action.Name(), "error", err.Error())
			}
			errs = append(errs, fmt.Errorf("%s: %w", action.Name(), err))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
