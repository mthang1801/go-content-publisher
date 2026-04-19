package application

import "context"

type CrawlPort interface {
	Name() string
	Enabled() bool
	Crawl(ctx context.Context) error
}

type RewritePort interface {
	Name() string
	Enabled() bool
	Rewrite(ctx context.Context) error
}

type PublishPort interface {
	Name() string
	Enabled() bool
	Publish(ctx context.Context) error
}

type RewriteTextPort interface {
	RewriteText(ctx context.Context, originalText string) (RewriteResult, error)
}

type RewriteResult struct {
	RewrittenText   string
	RewrittenTextEn string
	TweetTextVI     string
	TweetTextEN     string
	FactCheckNote   string
	ShouldPublish   bool
	Reason          string
}

func PlainRewriteResult(text string) RewriteResult {
	return RewriteResult{
		RewrittenText: text,
		ShouldPublish: true,
	}
}

type PublishTextPort interface {
	PublishText(ctx context.Context, text string) (string, error)
}

type TwitterPublishInput struct {
	TweetTextVI *string
	TweetTextEN *string
}

type TwitterPublishResult struct {
	TweetViID string
	TweetEnID string
}

type PublishTweetsPort interface {
	PublishTweets(ctx context.Context, input TwitterPublishInput) (TwitterPublishResult, error)
}
