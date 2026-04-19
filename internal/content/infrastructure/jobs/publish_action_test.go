package jobs

import (
	"context"
	"log/slog"
	"testing"

	contentapp "go-content-bot/internal/content/application"
)

type publishActionPublisherStub struct {
	calls int
}

func (s *publishActionPublisherStub) PublishText(ctx context.Context, text string) (string, error) {
	s.calls++
	return "msg-1", nil
}

func TestPublishActionReadsAutoPublishSettingAtRuntime(t *testing.T) {
	t.Parallel()

	service := contentapp.NewService(queuePolicyRepoStub{})
	publisher := &publishActionPublisherStub{}
	action := NewPublishAction(
		service,
		publisher,
		queueSettingsStub{values: map[string]string{autoPublishSettingKey: "false"}},
		slog.New(slog.NewTextHandler(testWriter{t}, nil)),
		true,
	)

	if err := action.Publish(context.Background()); err != nil {
		t.Fatalf("publish action: %v", err)
	}
	if publisher.calls != 0 {
		t.Fatalf("expected publisher not called when auto_publish=false, got %d calls", publisher.calls)
	}
}
