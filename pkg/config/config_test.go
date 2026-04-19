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
	t.Setenv("APP_LOG_MAX_SIZE_MB", "")
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
log_max_size_mb=25
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
	if cfg.App.LogMaxSizeMB != 25 {
		t.Fatalf("expected log max size 25MB, got %d", cfg.App.LogMaxSizeMB)
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

func TestLoadFromINIFilePreservesEmptyHTTPPort(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_PORT", "")

	temp := t.TempDir()
	path := filepath.Join(temp, "config.ini")
	content := `
[app]
name=content-bot-go

[http]
port=

[database]
url=postgres://user:pass@example.com:5432/content_bot?sslmode=require
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write ini file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config from ini: %v", err)
	}

	if cfg.HTTP.Port != "" {
		t.Fatalf("expected empty HTTP port to be preserved, got %q", cfg.HTTP.Port)
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

func TestLoadDetailsFromPathsTracksLoadedFiles(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("DATABASE_URL", "")

	temp := t.TempDir()
	path := filepath.Join(temp, "config.ini")
	content := `
[app]
name=content-bot-go

[database]
url=postgres://user:pass@example.com:5432/content_bot?sslmode=require
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write ini file: %v", err)
	}

	details, err := LoadDetailsFromPaths(path)
	if err != nil {
		t.Fatalf("load details from ini: %v", err)
	}

	if len(details.LoadedPaths) != 1 {
		t.Fatalf("expected 1 loaded path, got %d", len(details.LoadedPaths))
	}
	if details.LoadedPaths[0] != path {
		absolutePath, _ := filepath.Abs(path)
		if details.LoadedPaths[0] != absolutePath {
			t.Fatalf("expected loaded path %q or %q, got %q", path, absolutePath, details.LoadedPaths[0])
		}
	}
}

func TestRuntimeSummaryInfersSupabaseSessionPooler(t *testing.T) {
	cfg := Config{
		App: AppConfig{
			Name:   "content-bot-go",
			Env:    "production",
			RunAPI: true,
		},
		HTTP: HTTPConfig{
			Host: "0.0.0.0",
			Port: "auto",
		},
		Database: DatabaseConfig{
			URL:     "postgres://postgres.project-ref:secret@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require",
			SSLMode: "require",
		},
	}

	summary := cfg.RuntimeSummary([]string{"/tmp/config.ini"})
	if summary.Database.Mode != "supabase_session_pooler" {
		t.Fatalf("expected supabase session pooler mode, got %q", summary.Database.Mode)
	}
	if summary.Database.Target != "aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres" {
		t.Fatalf("unexpected database target: %q", summary.Database.Target)
	}
	if summary.HTTP.PortMode != "auto" {
		t.Fatalf("expected auto port mode, got %q", summary.HTTP.PortMode)
	}
	if len(summary.ConfigFiles) != 1 || summary.ConfigFiles[0] != "/tmp/config.ini" {
		t.Fatalf("unexpected config files: %#v", summary.ConfigFiles)
	}
}

func TestRuntimeSummaryInfersLocalSplitFieldDatabase(t *testing.T) {
	cfg := Config{
		App: AppConfig{Name: "content-bot-go"},
		HTTP: HTTPConfig{
			Host: "127.0.0.1",
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:    "127.0.0.1",
			Port:    "5432",
			Name:    "content_bot",
			SSLMode: "disable",
		},
	}

	summary := cfg.RuntimeSummary(nil)
	if summary.Database.Mode != "local" {
		t.Fatalf("expected local database mode, got %q", summary.Database.Mode)
	}
	if summary.Database.ConfigSource != "DATABASE_HOST/DATABASE_NAME" {
		t.Fatalf("unexpected config source: %q", summary.Database.ConfigSource)
	}
	if summary.HTTP.PortMode != "fixed" {
		t.Fatalf("expected fixed port mode, got %q", summary.HTTP.PortMode)
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
