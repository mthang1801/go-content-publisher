package bootstrap

import (
	"context"
	"testing"
	"time"

	"go-content-bot/pkg/config"
)

func TestApplyTelegramRuntimeJSON(t *testing.T) {
	raw := []byte(`{
		"bot_token": " test-token ",
		"publish_targets": [{"chat_id":" -1002451344189 ","thread_id":5}],
		"ingest_targets": [{"chat_id":"-1002451344189","thread_id":5}],
		"admin_user_ids": [1,2,3]
	}`)

	target := &config.TelegramConfig{
		BotToken:      "env-token",
		TargetChannel: "-100legacy",
	}

	if err := applyTelegramRuntimeJSON(raw, target); err != nil {
		t.Fatalf("apply telegram runtime json: %v", err)
	}

	if target.BotToken != "test-token" {
		t.Fatalf("expected bot token override, got %q", target.BotToken)
	}
	if target.TargetChannel != "" {
		t.Fatalf("expected target channel cleared, got %q", target.TargetChannel)
	}
	if len(target.PublishTargets) != 1 || target.PublishTargets[0].ChatID != "-1002451344189" {
		t.Fatalf("unexpected publish targets: %#v", target.PublishTargets)
	}
	if target.PublishTargets[0].ThreadID == nil || *target.PublishTargets[0].ThreadID != 5 {
		t.Fatalf("unexpected publish thread id: %#v", target.PublishTargets[0].ThreadID)
	}
	if len(target.AdminUserIDs) != 3 {
		t.Fatalf("unexpected admin ids: %#v", target.AdminUserIDs)
	}
}

func TestApplyTelegramRuntimeJSONRejectsBlankChatID(t *testing.T) {
	target := &config.TelegramConfig{}
	err := applyTelegramRuntimeJSON([]byte(`{"publish_targets":[{"chat_id":"   "}]} `), target)
	if err == nil {
		t.Fatal("expected error for blank chat_id")
	}
}

type bootstrapSettingRepoStub struct {
	values map[string]string
}

func (s bootstrapSettingRepoStub) Get(ctx context.Context, key string) (string, error) {
	return s.values[key], nil
}

func TestApplySecondsSettingOverride(t *testing.T) {
	target := 10 * time.Second

	err := applySecondsSettingOverride(context.Background(), bootstrapSettingRepoStub{values: map[string]string{
		"crawl_interval_seconds": "300",
	}}, "crawl_interval_seconds", &target)
	if err != nil {
		t.Fatalf("apply seconds setting override: %v", err)
	}
	if target != 300*time.Second {
		t.Fatalf("expected 300s, got %s", target)
	}
}

func TestApplySecondsSettingOverrideRejectsNonPositiveValue(t *testing.T) {
	target := 10 * time.Second

	err := applySecondsSettingOverride(context.Background(), bootstrapSettingRepoStub{values: map[string]string{
		"crawl_interval_seconds": "0",
	}}, "crawl_interval_seconds", &target)
	if err == nil {
		t.Fatal("expected error for non-positive scheduler interval")
	}
}
