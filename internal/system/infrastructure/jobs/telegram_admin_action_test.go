package jobs

import (
	"context"
	"strings"
	"testing"
	"time"

	contentapp "go-content-bot/internal/content/application"
	contentdomain "go-content-bot/internal/content/domain"
	sourcedomain "go-content-bot/internal/source/domain"
	telegrambot "go-content-bot/internal/system/infrastructure/clients/telegrambot"
	systemrepo "go-content-bot/internal/system/infrastructure/persistence/repositories"
)

type telegramAdminContentStub struct {
	queue        []contentdomain.ContentItem
	recent       []contentdomain.ContentItem
	retryCount   int
	skippedID    string
	skipPrefix   string
	skipReason   string
	manualInputs []contentapp.EnqueueManualInput
}

func (s *telegramAdminContentStub) ListQueue(ctx context.Context, limit int) ([]contentdomain.ContentItem, error) {
	return s.queue, nil
}

func (s *telegramAdminContentStub) ListRecent(ctx context.Context, limit int) ([]contentdomain.ContentItem, error) {
	return s.recent, nil
}

func (s *telegramAdminContentStub) RetryFailed(ctx context.Context, limit int) (int, error) {
	return s.retryCount, nil
}

func (s *telegramAdminContentStub) SkipByIDPrefix(ctx context.Context, prefix, reason string) (contentdomain.ContentItem, error) {
	s.skipPrefix = prefix
	s.skipReason = reason
	return contentdomain.ContentItem{ID: s.skippedID, Status: contentdomain.StatusSkipped, OriginalText: "skipped"}, nil
}

func (s *telegramAdminContentStub) EnqueueManual(ctx context.Context, input contentapp.EnqueueManualInput) (contentdomain.ContentItem, error) {
	s.manualInputs = append(s.manualInputs, input)
	return contentdomain.ContentItem{
		ID:           "manual-queued-id",
		Status:       contentdomain.StatusPending,
		OriginalText: input.Text,
		AuthorName:   input.Author,
	}, nil
}

type telegramAdminSourceStub struct {
	sources      []sourcedomain.Source
	created      *sourcedomain.Source
	activated    string
	inactivated  string
	inactiveNote string
	createErr    error
}

func (s *telegramAdminSourceStub) ListAll(ctx context.Context) ([]sourcedomain.Source, error) {
	return s.sources, nil
}

func (s *telegramAdminSourceStub) Create(ctx context.Context, source sourcedomain.Source) (sourcedomain.Source, error) {
	if s.createErr != nil {
		return sourcedomain.Source{}, s.createErr
	}
	source.ID = "source-created"
	source.IsActive = true
	s.created = &source
	return source, nil
}

func (s *telegramAdminSourceStub) MarkActive(ctx context.Context, id string, at time.Time) error {
	s.activated = id
	return nil
}

func (s *telegramAdminSourceStub) MarkInactive(ctx context.Context, id string, reason string, at time.Time) error {
	s.inactivated = id
	s.inactiveNote = reason
	return nil
}

type telegramAdminSettingsStub struct {
	values map[string]string
}

func (s telegramAdminSettingsStub) Get(ctx context.Context, key string) (string, error) {
	return s.values[key], nil
}

func (s telegramAdminSettingsStub) Set(ctx context.Context, key, value string) error {
	s.values[key] = value
	return nil
}

type telegramAdminSenderStub struct {
	messages []telegramAdminSentMessage
}

type telegramAdminSentMessage struct {
	chatID   string
	threadID *int64
	text     string
}

func (s *telegramAdminSenderStub) SendMessage(ctx context.Context, chatID string, threadID *int64, text string) (string, error) {
	s.messages = append(s.messages, telegramAdminSentMessage{chatID: chatID, threadID: threadID, text: text})
	return "msg-1", nil
}

type telegramAdminLogsStub struct {
	rows []systemrepo.LogRecord
}

func (s telegramAdminLogsStub) ListRecent(ctx context.Context, limit int) ([]systemrepo.LogRecord, error) {
	return s.rows, nil
}

func TestTelegramAdminHandlerTogglesAutoPublish(t *testing.T) {
	t.Parallel()

	threadID := int64(5)
	sender := &telegramAdminSenderStub{}
	settings := telegramAdminSettingsStub{values: map[string]string{"auto_publish": "true"}}
	handler := NewTelegramAdminHandler(
		[]int64{42},
		&telegramAdminContentStub{},
		&telegramAdminSourceStub{},
		settings,
		sender,
		nil,
		nil,
	)

	handled, err := handler.Handle(context.Background(), telegrambot.Message{
		Text:            "/pause",
		MessageThreadID: &threadID,
		Chat:            telegrambot.Chat{ID: -1002451344189},
		From:            &telegrambot.User{ID: 42},
	})
	if err != nil {
		t.Fatalf("handle pause: %v", err)
	}
	if !handled {
		t.Fatal("expected command handled")
	}
	if settings.values["auto_publish"] != "false" {
		t.Fatalf("expected auto_publish=false, got %q", settings.values["auto_publish"])
	}
	if len(sender.messages) != 1 || sender.messages[0].chatID != "-1002451344189" || sender.messages[0].threadID == nil || *sender.messages[0].threadID != 5 {
		t.Fatalf("unexpected reply target: %#v", sender.messages)
	}

	if _, err := handler.Handle(context.Background(), telegrambot.Message{
		Text: "/resume",
		Chat: telegrambot.Chat{ID: -1002451344189},
		From: &telegrambot.User{ID: 42},
	}); err != nil {
		t.Fatalf("handle resume: %v", err)
	}
	if settings.values["auto_publish"] != "true" {
		t.Fatalf("expected auto_publish=true, got %q", settings.values["auto_publish"])
	}
}

func TestTelegramAdminHandlerRejectsUnauthorizedUser(t *testing.T) {
	t.Parallel()

	sender := &telegramAdminSenderStub{}
	settings := telegramAdminSettingsStub{values: map[string]string{"auto_publish": "true"}}
	handler := NewTelegramAdminHandler(
		[]int64{42},
		&telegramAdminContentStub{},
		&telegramAdminSourceStub{},
		settings,
		sender,
		nil,
		nil,
	)

	handled, err := handler.Handle(context.Background(), telegrambot.Message{
		Text: "/pause",
		Chat: telegrambot.Chat{ID: -1002451344189},
		From: &telegrambot.User{ID: 99},
	})
	if err != nil {
		t.Fatalf("handle unauthorized: %v", err)
	}
	if !handled {
		t.Fatal("expected command handled")
	}
	if settings.values["auto_publish"] != "true" {
		t.Fatalf("expected setting unchanged, got %q", settings.values["auto_publish"])
	}
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "Not authorized") {
		t.Fatalf("expected unauthorized reply, got %#v", sender.messages)
	}
}

func TestTelegramAdminHandlerAddsAndRemovesSources(t *testing.T) {
	t.Parallel()

	sender := &telegramAdminSenderStub{}
	sources := &telegramAdminSourceStub{
		sources: []sourcedomain.Source{
			{ID: "source-twitter", Type: sourcedomain.TypeTwitter, Handle: "@zerohedge", Name: "ZeroHedge", IsActive: true},
		},
	}
	handler := NewTelegramAdminHandler(
		nil,
		&telegramAdminContentStub{},
		sources,
		telegramAdminSettingsStub{values: map[string]string{}},
		sender,
		nil,
		nil,
	)

	if _, err := handler.Handle(context.Background(), telegrambot.Message{
		Text: "/addsource twitter @business Bloomberg",
		Chat: telegrambot.Chat{ID: -1002451344189},
		From: &telegrambot.User{ID: 1, FirstName: "Admin"},
	}); err != nil {
		t.Fatalf("handle addsource: %v", err)
	}
	if sources.created == nil || sources.created.Type != sourcedomain.TypeTwitter || sources.created.Handle != "@business" || sources.created.Name != "Bloomberg" {
		t.Fatalf("unexpected created source: %#v", sources.created)
	}

	if _, err := handler.Handle(context.Background(), telegrambot.Message{
		Text: "/removesource twitter zerohedge",
		Chat: telegrambot.Chat{ID: -1002451344189},
		From: &telegrambot.User{ID: 1, FirstName: "Admin"},
	}); err != nil {
		t.Fatalf("handle removesource: %v", err)
	}
	if sources.inactivated != "source-twitter" {
		t.Fatalf("expected source-twitter inactivated, got %q", sources.inactivated)
	}
}

func TestTelegramAdminHandlerReactivatesExistingInactiveSource(t *testing.T) {
	t.Parallel()

	sender := &telegramAdminSenderStub{}
	sources := &telegramAdminSourceStub{
		createErr: sourcedomain.ErrSourceAlreadyExists,
		sources: []sourcedomain.Source{
			{ID: "source-twitter", Type: sourcedomain.TypeTwitter, Handle: "@zerohedge", Name: "ZeroHedge", IsActive: false},
		},
	}
	handler := NewTelegramAdminHandler(
		nil,
		&telegramAdminContentStub{},
		sources,
		telegramAdminSettingsStub{values: map[string]string{}},
		sender,
		nil,
		nil,
	)

	if _, err := handler.Handle(context.Background(), telegrambot.Message{
		Text: "/addsource twitter zerohedge ZeroHedge",
		Chat: telegrambot.Chat{ID: -1002451344189},
		From: &telegrambot.User{ID: 1, FirstName: "Admin"},
	}); err != nil {
		t.Fatalf("handle addsource existing inactive: %v", err)
	}
	if sources.activated != "source-twitter" {
		t.Fatalf("expected source-twitter activated, got %q", sources.activated)
	}
}

func TestTelegramAdminHandlerQueueRetrySkipLogsAndCrawlNow(t *testing.T) {
	t.Parallel()

	crawled := false
	sender := &telegramAdminSenderStub{}
	content := &telegramAdminContentStub{
		queue: []contentdomain.ContentItem{
			{ID: "12345678-aaaa", Status: contentdomain.StatusPending, OriginalText: "pending item"},
		},
		retryCount: 2,
		skippedID:  "12345678-aaaa",
	}
	handler := NewTelegramAdminHandler(
		nil,
		content,
		&telegramAdminSourceStub{},
		telegramAdminSettingsStub{values: map[string]string{}},
		sender,
		telegramAdminLogsStub{rows: []systemrepo.LogRecord{{Level: "info", Module: "worker", Message: "started", CreatedAt: time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)}}},
		func(ctx context.Context) error {
			crawled = true
			return nil
		},
	)

	for _, command := range []string{"/queue", "/retry", "/skip 1234", "/logs", "/crawlnow"} {
		if handled, err := handler.Handle(context.Background(), telegrambot.Message{
			Text: command,
			Chat: telegrambot.Chat{ID: -1002451344189},
			From: &telegrambot.User{ID: 1, FirstName: "Admin"},
		}); err != nil {
			t.Fatalf("handle %s: %v", command, err)
		} else if !handled {
			t.Fatalf("expected %s handled", command)
		}
	}

	if content.skipPrefix != "1234" {
		t.Fatalf("expected skip prefix 1234, got %q", content.skipPrefix)
	}
	if !crawled {
		t.Fatal("expected crawlnow callback invoked")
	}
	if len(sender.messages) != 5 {
		t.Fatalf("expected 5 replies, got %d", len(sender.messages))
	}
}

func TestTelegramAdminHandlerAddsManualContentFromCommandAndPlainText(t *testing.T) {
	t.Parallel()

	sender := &telegramAdminSenderStub{}
	content := &telegramAdminContentStub{}
	handler := NewTelegramAdminHandler(
		[]int64{42},
		content,
		&telegramAdminSourceStub{},
		telegramAdminSettingsStub{values: map[string]string{}},
		sender,
		nil,
		nil,
	)

	for _, message := range []telegrambot.Message{
		{
			Text: "/add Market breadth improved after softer bond yields",
			Chat: telegrambot.Chat{ID: -1002451344189},
			From: &telegrambot.User{ID: 42, FirstName: "Minh"},
		},
		{
			Text: "ECB comments lifted risk appetite across European equities",
			Chat: telegrambot.Chat{ID: -1002451344189},
			From: &telegrambot.User{ID: 42, FirstName: "Minh"},
		},
	} {
		handled, err := handler.Handle(context.Background(), message)
		if err != nil {
			t.Fatalf("handle manual message: %v", err)
		}
		if !handled {
			t.Fatal("expected manual content handled")
		}
	}

	if len(content.manualInputs) != 2 {
		t.Fatalf("expected 2 manual inputs, got %d", len(content.manualInputs))
	}
	if content.manualInputs[0].Text != "Market breadth improved after softer bond yields" {
		t.Fatalf("unexpected /add text: %#v", content.manualInputs[0])
	}
	if content.manualInputs[1].Text != "ECB comments lifted risk appetite across European equities" {
		t.Fatalf("unexpected plain text: %#v", content.manualInputs[1])
	}
	if content.manualInputs[0].Author != "Minh" || content.manualInputs[1].Author != "Minh" {
		t.Fatalf("unexpected authors: %#v", content.manualInputs)
	}
	if len(sender.messages) != 2 {
		t.Fatalf("expected 2 replies, got %d", len(sender.messages))
	}
	if !strings.Contains(sender.messages[0].text, "Added manual content") || !strings.Contains(sender.messages[1].text, "Added manual content") {
		t.Fatalf("unexpected replies: %#v", sender.messages)
	}
}

func TestTelegramAdminHandlerIgnoresUnauthorizedPlainText(t *testing.T) {
	t.Parallel()

	sender := &telegramAdminSenderStub{}
	content := &telegramAdminContentStub{}
	handler := NewTelegramAdminHandler(
		[]int64{42},
		content,
		&telegramAdminSourceStub{},
		telegramAdminSettingsStub{values: map[string]string{}},
		sender,
		nil,
		nil,
	)

	handled, err := handler.Handle(context.Background(), telegrambot.Message{
		Text: "manual note",
		Chat: telegrambot.Chat{ID: -1002451344189},
		From: &telegrambot.User{ID: 99, FirstName: "Other"},
	})
	if err != nil {
		t.Fatalf("handle unauthorized plain text: %v", err)
	}
	if handled {
		t.Fatal("expected unauthorized plain text not handled")
	}
	if len(content.manualInputs) != 0 {
		t.Fatalf("expected no manual enqueue, got %#v", content.manualInputs)
	}
	if len(sender.messages) != 0 {
		t.Fatalf("expected no reply, got %#v", sender.messages)
	}
}
