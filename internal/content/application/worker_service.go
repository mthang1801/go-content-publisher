package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go-content-bot/pkg/config"
)

type WorkerService struct {
	logger           *slog.Logger
	scheduler        config.SchedulerConfig
	features         config.FeatureConfig
	crawler          CrawlPort
	rewriter         RewritePort
	publisher        PublishPort
	twitterPublisher PublishPort
}

func NewWorkerService(
	logger *slog.Logger,
	scheduler config.SchedulerConfig,
	features config.FeatureConfig,
	crawler CrawlPort,
	rewriter RewritePort,
	publisher PublishPort,
	twitterPublisher PublishPort,
) *WorkerService {
	return &WorkerService{
		logger:           logger,
		scheduler:        scheduler,
		features:         features,
		crawler:          crawler,
		rewriter:         rewriter,
		publisher:        publisher,
		twitterPublisher: twitterPublisher,
	}
}

func (s *WorkerService) Run(ctx context.Context) error {
	s.logger.Info("worker starting",
		"auto_publish", s.features.AutoPublish,
		"telegram_crawler_enabled", s.crawler.Enabled(),
		"rewriter_enabled", s.rewriter.Enabled(),
		"publisher_enabled", s.publisher.Enabled(),
		"twitter_publisher_enabled", s.twitterPublisher.Enabled(),
	)

	if err := runLoop(ctx, s.logger, s.scheduler.CrawlInterval, s.crawler.Name(), s.crawler.Crawl); err != nil {
		return err
	}
	if err := runLoop(ctx, s.logger, s.scheduler.ProcessInterval, s.rewriter.Name(), s.rewriter.Rewrite); err != nil {
		return err
	}
	if err := runLoop(ctx, s.logger, s.scheduler.PublishInterval, s.publisher.Name(), s.publisher.Publish); err != nil {
		return err
	}
	if err := runLoop(ctx, s.logger, s.scheduler.TwitterPublishInterval, s.twitterPublisher.Name(), s.twitterPublisher.Publish); err != nil {
		return err
	}

	<-ctx.Done()
	s.logger.Info("worker stopping")
	return nil
}

func runLoop(ctx context.Context, logger *slog.Logger, interval time.Duration, name string, fn func(context.Context) error) error {
	if interval <= 0 {
		return errors.New("worker interval must be positive")
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		logger.Info("worker loop configured", "loop", name, "interval", interval.String())
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := fn(ctx); err != nil && !errors.Is(err, context.Canceled) {
					logger.Warn("worker loop failed", "loop", name, "error", err.Error())
				}
			}
		}
	}()
	return nil
}

type NoopAction struct {
	name    string
	enabled bool
	reason  string
	logger  *slog.Logger
}

func NewNoopAction(name string, enabled bool, reason string, logger *slog.Logger) *NoopAction {
	return &NoopAction{name: name, enabled: enabled, reason: reason, logger: logger}
}

func (n *NoopAction) Name() string {
	return n.name
}

func (n *NoopAction) Enabled() bool {
	return n.enabled
}

func (n *NoopAction) Crawl(ctx context.Context) error {
	n.logger.Info("noop crawl adapter invoked", "adapter", n.name, "enabled", n.enabled, "reason", n.reason)
	return nil
}

func (n *NoopAction) Rewrite(ctx context.Context) error {
	n.logger.Info("noop rewrite adapter invoked", "adapter", n.name, "enabled", n.enabled, "reason", n.reason)
	return nil
}

func (n *NoopAction) Publish(ctx context.Context) error {
	n.logger.Info("noop publish adapter invoked", "adapter", n.name, "enabled", n.enabled, "reason", n.reason)
	return nil
}
