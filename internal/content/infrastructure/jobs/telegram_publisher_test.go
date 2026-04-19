package jobs

import (
	"context"
	"testing"

	"go-content-bot/pkg/config"
)

type telegramSendCall struct {
	chatID   string
	threadID *int64
	text     string
}

type telegramPublisherClientStub struct {
	calls []telegramSendCall
}

func (s *telegramPublisherClientStub) SendMessage(ctx context.Context, chatID string, threadID *int64, text string) (string, error) {
	s.calls = append(s.calls, telegramSendCall{
		chatID:   chatID,
		threadID: threadID,
		text:     text,
	})
	return chatID + ":ok", nil
}

func TestTelegramPublisherPublishesToAllTargets(t *testing.T) {
	t.Parallel()

	threadID := int64(5)
	client := &telegramPublisherClientStub{}
	publisher := NewTelegramPublisher(client, []config.TelegramTarget{
		{ChatID: "-1002451344189", ThreadID: &threadID},
		{ChatID: "-1002451344190"},
	})

	messageID, err := publisher.PublishText(context.Background(), "hello topic")
	if err != nil {
		t.Fatalf("publish text: %v", err)
	}
	if len(client.calls) != 2 {
		t.Fatalf("expected 2 send calls, got %d", len(client.calls))
	}
	if client.calls[0].threadID == nil || *client.calls[0].threadID != 5 {
		t.Fatalf("expected first call thread id 5, got %#v", client.calls[0].threadID)
	}
	if client.calls[1].threadID != nil {
		t.Fatalf("expected second call without thread id, got %#v", client.calls[1].threadID)
	}
	expected := `[{"chat_id":"-1002451344189","thread_id":5,"message_id":"-1002451344189:ok"},{"chat_id":"-1002451344190","message_id":"-1002451344190:ok"}]`
	if messageID != expected {
		t.Fatalf("expected aggregated publish result %s, got %s", expected, messageID)
	}
}
