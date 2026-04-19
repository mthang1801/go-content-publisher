package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnvFile(t *testing.T) {
	t.Setenv("APP_NAME", "")
	temp := t.TempDir()
	path := filepath.Join(temp, ".env")
	content := `
APP_NAME=content-bot-go
APP_ENV=test
HTTP_HOST=127.0.0.1
HTTP_PORT=9090
DATABASE_URL=postgres://user:pass@example.com:5432/content_bot?sslmode=require
REWRITE_PROVIDER=gemini
ADMIN_USER_IDS=1,2,3
TELEGRAM_PUBLISH_TARGETS=[{"chat_id":"-1002451344189","thread_id":5}]
TELEGRAM_INGEST_TARGETS=[{"chat_id":"-1002451344189","thread_id":5}]
GEMINI_API_KEY=test-gemini-key
GEMINI_MODEL=gemini-2.5-flash
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.HTTP.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.HTTP.Port)
	}
	if len(cfg.Telegram.AdminUserIDs) != 3 {
		t.Fatalf("expected 3 admin IDs, got %d", len(cfg.Telegram.AdminUserIDs))
	}
	if len(cfg.Telegram.PublishTargets) != 1 || cfg.Telegram.PublishTargets[0].ChatID != "-1002451344189" {
		t.Fatalf("expected 1 publish target, got %#v", cfg.Telegram.PublishTargets)
	}
	if cfg.Telegram.PublishTargets[0].ThreadID == nil || *cfg.Telegram.PublishTargets[0].ThreadID != 5 {
		t.Fatalf("expected publish thread id 5, got %#v", cfg.Telegram.PublishTargets[0].ThreadID)
	}
	if len(cfg.Telegram.IngestTargets) != 1 || cfg.Telegram.IngestTargets[0].ChatID != "-1002451344189" {
		t.Fatalf("expected 1 ingest target, got %#v", cfg.Telegram.IngestTargets)
	}
	if cfg.Database.DSN() != "postgres://user:pass@example.com:5432/content_bot?sslmode=require" {
		t.Fatalf("unexpected dsn: %s", cfg.Database.DSN())
	}
	if cfg.Rewrite.Provider != "gemini" {
		t.Fatalf("expected rewrite provider gemini, got %s", cfg.Rewrite.Provider)
	}
	if cfg.Gemini.APIKey != "test-gemini-key" {
		t.Fatalf("expected gemini api key, got %s", cfg.Gemini.APIKey)
	}
	if cfg.Gemini.Model != "gemini-2.5-flash" {
		t.Fatalf("expected gemini model, got %s", cfg.Gemini.Model)
	}
}

func TestLoadFallsBackToLegacyTelegramTargetChannel(t *testing.T) {
	t.Setenv("APP_NAME", "content-bot-go")
	t.Setenv("DATABASE_URL", "postgres://user:pass@example.com:5432/content_bot?sslmode=require")
	t.Setenv("TELEGRAM_TARGET_CHANNEL", "-100legacy")
	t.Setenv("TELEGRAM_PUBLISH_TARGETS", "")
	t.Setenv("TELEGRAM_INGEST_TARGETS", "")

	cfg, err := Load("/path/that/does/not/exist")
	if err != nil {
		t.Fatalf("load config with legacy telegram target: %v", err)
	}

	if len(cfg.Telegram.PublishTargets) != 1 || cfg.Telegram.PublishTargets[0].ChatID != "-100legacy" {
		t.Fatalf("expected legacy publish target fallback, got %#v", cfg.Telegram.PublishTargets)
	}
	if len(cfg.Telegram.IngestTargets) != 1 || cfg.Telegram.IngestTargets[0].ChatID != "-100legacy" {
		t.Fatalf("expected legacy ingest target fallback, got %#v", cfg.Telegram.IngestTargets)
	}
}

func TestLoadFromINIFile(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_LOG_LEVEL", "")
	t.Setenv("APP_AUTO_MIGRATE", "")
	t.Setenv("APP_RUN_API", "")
	t.Setenv("APP_RUN_WORKER", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REWRITE_PROVIDER", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", "")
	t.Setenv("ADMIN_USER_IDS", "")
	t.Setenv("TELEGRAM_PUBLISH_TARGETS", "")
	t.Setenv("TELEGRAM_INGEST_TARGETS", "")

	temp := t.TempDir()
	path := filepath.Join(temp, "config.ini")
	content := `
[app]
name=content-bot-go
env=production
log_level=warn
auto_migrate=true
run_api=true
run_worker=false

[http]
port=9191

[database]
url=postgres://user:pass@example.com:5432/content_bot?sslmode=require

[rewrite]
provider=gemini

[gemini]
api_key=test-gemini-key
model=gemini-2.5-flash

[telegram]
admin_user_ids=11,22
publish_targets=[{"chat_id":"-1001"}]
ingest_targets=[{"chat_id":"-1002"}]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write ini file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config from ini: %v", err)
	}

	if !cfg.App.AutoMigrate {
		t.Fatal("expected app auto migrate true")
	}
	if cfg.App.RunWorker {
		t.Fatal("expected app run worker false")
	}
	if cfg.HTTP.Port != "9191" {
		t.Fatalf("expected port 9191, got %s", cfg.HTTP.Port)
	}
	if len(cfg.Telegram.AdminUserIDs) != 2 {
		t.Fatalf("expected 2 admin IDs, got %d", len(cfg.Telegram.AdminUserIDs))
	}
	if len(cfg.Telegram.PublishTargets) != 1 || cfg.Telegram.PublishTargets[0].ChatID != "-1001" {
		t.Fatalf("unexpected publish targets: %#v", cfg.Telegram.PublishTargets)
	}
}

func TestINIDoesNotOverrideExistingEnv(t *testing.T) {
	t.Setenv("APP_NAME", "content-bot-go")
	t.Setenv("DATABASE_URL", "postgres://user:pass@example.com:5432/content_bot?sslmode=require")
	t.Setenv("HTTP_PORT", "7070")

	temp := t.TempDir()
	path := filepath.Join(temp, "config.ini")
	content := `
[http]
port=9191
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write ini file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config from ini with env override: %v", err)
	}
	if cfg.HTTP.Port != "7070" {
		t.Fatalf("expected env port 7070 to win, got %s", cfg.HTTP.Port)
	}
}

func TestValidateRequiresDatabaseSettings(t *testing.T) {
	t.Setenv("APP_NAME", "content-bot-go")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_HOST", "")
	t.Setenv("DATABASE_NAME", "")
	t.Setenv("DATABASE_USER", "")
	t.Setenv("DATABASE_PASSWORD", "")

	if _, err := Load("/path/that/does/not/exist"); err == nil {
		t.Fatal("expected validation error when database config is incomplete")
	}
}

func TestValidateRejectsUnknownRewriteProvider(t *testing.T) {
	t.Setenv("APP_NAME", "content-bot-go")
	t.Setenv("DATABASE_URL", "postgres://user:pass@example.com:5432/content_bot?sslmode=require")
	t.Setenv("REWRITE_PROVIDER", "unknown")

	if _, err := Load("/path/that/does/not/exist"); err == nil {
		t.Fatal("expected validation error for unknown rewrite provider")
	}
}
