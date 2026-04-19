package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Database  DatabaseConfig
	Scheduler SchedulerConfig
	Feature   FeatureConfig
	Rewrite   RewriteConfig
	Telegram  TelegramConfig
	Twitter   TwitterConfig
	DeepSeek  DeepSeekConfig
	Gemini    GeminiConfig
	Docker    DockerConfig
}

type AppConfig struct {
	Name        string
	Env         string
	LogLevel    string
	AutoMigrate bool
	RunAPI      bool
	RunWorker   bool
}

type HTTPConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL             string
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type SchedulerConfig struct {
	CrawlInterval          time.Duration
	ProcessInterval        time.Duration
	PublishInterval        time.Duration
	TwitterPublishInterval time.Duration
}

type FeatureConfig struct {
	AutoPublish            bool
	EnableTelegramCrawler  bool
	EnableTwitterCrawler   bool
	EnableRewriteProcessor bool
	EnableTwitterPublishVI bool
	EnableTwitterPublishEN bool
}

type RewriteConfig struct {
	Provider string
}

type TelegramConfig struct {
	BotToken       string
	TargetChannel  string
	AdminUserIDs   []int64
	PublishTargets []TelegramTarget
	IngestTargets  []TelegramTarget
}

type TelegramTarget struct {
	ChatID   string `json:"chat_id"`
	ThreadID *int64 `json:"thread_id,omitempty"`
}

type TwitterConfig struct {
	BearerToken    string
	VIAPIKey       string
	VIAPISecret    string
	VIAccessToken  string
	VIAccessSecret string
	ENAPIKey       string
	ENAPISecret    string
	ENAccessToken  string
	ENAccessSecret string
	PublishAfter   *time.Time
}

type DeepSeekConfig struct {
	APIKey string
	Model  string
}

type GeminiConfig struct {
	APIKey string
	Model  string
}

type DockerConfig struct {
	ImageName      string
	ImageTag       string
	ComposeProject string
}

func Load(paths ...string) (Config, error) {
	for _, path := range choosePaths(paths) {
		if _, err := os.Stat(path); err == nil {
			switch strings.ToLower(filepath.Ext(path)) {
			case ".ini":
				if err := loadINI(path); err != nil {
					return Config{}, fmt.Errorf("load ini file %s: %w", path, err)
				}
			default:
				if err := godotenv.Overload(path); err != nil {
					return Config{}, fmt.Errorf("load env file %s: %w", path, err)
				}
			}
		}
	}

	cfg := Config{
		App: AppConfig{
			Name:        getOrDefault("APP_NAME", "content-bot-go"),
			Env:         getOrDefault("APP_ENV", "development"),
			LogLevel:    getOrDefault("APP_LOG_LEVEL", "info"),
			AutoMigrate: mustBool("APP_AUTO_MIGRATE", false),
			RunAPI:      mustBool("APP_RUN_API", true),
			RunWorker:   mustBool("APP_RUN_WORKER", true),
		},
		HTTP: HTTPConfig{
			Host:         getOrDefault("HTTP_HOST", "0.0.0.0"),
			Port:         getOrDefault("HTTP_PORT", "8080"),
			ReadTimeout:  mustDuration("HTTP_READ_TIMEOUT", "15s"),
			WriteTimeout: mustDuration("HTTP_WRITE_TIMEOUT", "15s"),
		},
		Database: DatabaseConfig{
			URL:             os.Getenv("DATABASE_URL"),
			Host:            os.Getenv("DATABASE_HOST"),
			Port:            getOrDefault("DATABASE_PORT", "5432"),
			Name:            os.Getenv("DATABASE_NAME"),
			User:            os.Getenv("DATABASE_USER"),
			Password:        os.Getenv("DATABASE_PASSWORD"),
			SSLMode:         getOrDefault("DATABASE_SSL_MODE", "require"),
			MaxOpenConns:    mustInt("DATABASE_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    mustInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: mustDuration("DATABASE_CONN_MAX_LIFETIME", "30m"),
		},
		Scheduler: SchedulerConfig{
			CrawlInterval:          mustSeconds("CRAWL_INTERVAL_SECONDS", 300),
			ProcessInterval:        mustSeconds("PROCESS_INTERVAL_SECONDS", 30),
			PublishInterval:        mustSeconds("PUBLISH_INTERVAL_SECONDS", 10),
			TwitterPublishInterval: mustSeconds("TWITTER_PUBLISH_INTERVAL_SECONDS", 600),
		},
		Feature: FeatureConfig{
			AutoPublish:            mustBool("AUTO_PUBLISH", false),
			EnableTelegramCrawler:  mustBool("ENABLE_TELEGRAM_CRAWLER", false),
			EnableTwitterCrawler:   mustBool("ENABLE_TWITTER_CRAWLER", false),
			EnableRewriteProcessor: mustRewriteEnabled(),
			EnableTwitterPublishVI: mustBool("ENABLE_TWITTER_PUBLISH_VI", false),
			EnableTwitterPublishEN: mustBool("ENABLE_TWITTER_PUBLISH_EN", false),
		},
		Rewrite: RewriteConfig{
			Provider: getOrDefault("REWRITE_PROVIDER", "deepseek"),
		},
		Telegram: TelegramConfig{
			BotToken:       os.Getenv("TELEGRAM_BOT_TOKEN"),
			TargetChannel:  os.Getenv("TELEGRAM_TARGET_CHANNEL"),
			AdminUserIDs:   mustInt64Slice("ADMIN_USER_IDS"),
			PublishTargets: mustTelegramTargets("TELEGRAM_PUBLISH_TARGETS"),
			IngestTargets:  mustTelegramTargets("TELEGRAM_INGEST_TARGETS"),
		},
		Twitter: TwitterConfig{
			BearerToken:    os.Getenv("TWITTER_BEARER_TOKEN"),
			VIAPIKey:       os.Getenv("TWITTER_VI_API_KEY"),
			VIAPISecret:    os.Getenv("TWITTER_VI_API_SECRET"),
			VIAccessToken:  os.Getenv("TWITTER_VI_ACCESS_TOKEN"),
			VIAccessSecret: os.Getenv("TWITTER_VI_ACCESS_SECRET"),
			ENAPIKey:       os.Getenv("TWITTER_EN_API_KEY"),
			ENAPISecret:    os.Getenv("TWITTER_EN_API_SECRET"),
			ENAccessToken:  os.Getenv("TWITTER_EN_ACCESS_TOKEN"),
			ENAccessSecret: os.Getenv("TWITTER_EN_ACCESS_SECRET"),
			PublishAfter:   optionalTime("TWITTER_PUBLISH_AFTER"),
		},
		DeepSeek: DeepSeekConfig{
			APIKey: os.Getenv("DEEPSEEK_API_KEY"),
			Model:  getOrDefault("DEEPSEEK_MODEL", "deepseek-chat"),
		},
		Gemini: GeminiConfig{
			APIKey: os.Getenv("GEMINI_API_KEY"),
			Model:  getOrDefault("GEMINI_MODEL", "gemini-2.5-flash"),
		},
		Docker: DockerConfig{
			ImageName:      getOrDefault("DOCKER_IMAGE_NAME", "content-bot-go"),
			ImageTag:       getOrDefault("DOCKER_IMAGE_TAG", "dev"),
			ComposeProject: getOrDefault("COMPOSE_PROJECT_NAME", "content-bot-go"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	cfg.Telegram.applyLegacyFallbacks()

	return cfg, nil
}

func optionalTime(key string) *time.Time {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(fmt.Sprintf("invalid time for %s: %v", key, err))
	}
	return &parsed
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return errors.New("APP_NAME is required")
	}
	switch strings.ToLower(strings.TrimSpace(c.Rewrite.Provider)) {
	case "deepseek", "gemini":
	default:
		return fmt.Errorf("REWRITE_PROVIDER must be one of: deepseek, gemini")
	}
	if c.Database.URL == "" {
		missing := []string{}
		if c.Database.Host == "" {
			missing = append(missing, "DATABASE_HOST")
		}
		if c.Database.Name == "" {
			missing = append(missing, "DATABASE_NAME")
		}
		if c.Database.User == "" {
			missing = append(missing, "DATABASE_USER")
		}
		if c.Database.Password == "" {
			missing = append(missing, "DATABASE_PASSWORD")
		}
		if len(missing) > 0 {
			return fmt.Errorf("database configuration incomplete, missing: %s", strings.Join(missing, ", "))
		}
	}
	return nil
}

func (c DatabaseConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

func (c *TelegramConfig) applyLegacyFallbacks() {
	legacyChannel := strings.TrimSpace(c.TargetChannel)
	if legacyChannel == "" {
		return
	}

	if len(c.PublishTargets) == 0 {
		c.PublishTargets = []TelegramTarget{{ChatID: legacyChannel}}
	}
	if len(c.IngestTargets) == 0 {
		c.IngestTargets = []TelegramTarget{{ChatID: legacyChannel}}
	}
}

func mustRewriteEnabled() bool {
	raw := strings.TrimSpace(os.Getenv("ENABLE_REWRITE_PROCESSOR"))
	if raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			panic(fmt.Sprintf("invalid bool for ENABLE_REWRITE_PROCESSOR: %v", err))
		}
		return value
	}
	return mustBool("ENABLE_DEEPSEEK_PROCESSOR", false)
}

func choosePaths(paths []string) []string {
	if len(paths) > 0 {
		return paths
	}
	return []string{"config/config.ini", "config.ini", "config/.env", ".env"}
}

func getOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func mustInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid int for %s: %v", key, err))
	}
	return value
}

func mustBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid bool for %s: %v", key, err))
	}
	return value
}

func mustTelegramTargets(key string) []TelegramTarget {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	var targets []TelegramTarget
	if err := json.Unmarshal([]byte(raw), &targets); err != nil {
		panic(fmt.Sprintf("invalid telegram targets JSON for %s: %v", key, err))
	}

	for i := range targets {
		targets[i].ChatID = strings.TrimSpace(targets[i].ChatID)
		if targets[i].ChatID == "" {
			panic(fmt.Sprintf("invalid telegram target in %s: chat_id is required", key))
		}
	}

	return targets
}

func mustDuration(key, fallback string) time.Duration {
	raw := getOrDefault(key, fallback)
	value, err := time.ParseDuration(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid duration for %s: %v", key, err))
	}
	return value
}

func mustSeconds(key string, fallback int) time.Duration {
	return time.Duration(mustInt(key, fallback)) * time.Second
}

func mustInt64Slice(key string) []int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid int64 in %s: %v", key, err))
		}
		values = append(values, value)
	}
	return values
}

func loadINI(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		key, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid ini line %q", line)
		}
		envKey, ok := iniKeyToEnvKey(section, key)
		if !ok {
			continue
		}
		if strings.TrimSpace(os.Getenv(envKey)) != "" {
			continue
		}
		if err := os.Setenv(envKey, normalizeINIValue(rawValue)); err != nil {
			return fmt.Errorf("set %s from ini: %w", envKey, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func normalizeINIValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
	}
	return strings.TrimSpace(value)
}

func iniKeyToEnvKey(section, key string) (string, bool) {
	section = strings.ToLower(strings.TrimSpace(section))
	key = strings.ToLower(strings.TrimSpace(key))

	switch section {
	case "app":
		switch key {
		case "name":
			return "APP_NAME", true
		case "env":
			return "APP_ENV", true
		case "log_level":
			return "APP_LOG_LEVEL", true
		case "auto_migrate":
			return "APP_AUTO_MIGRATE", true
		case "run_api":
			return "APP_RUN_API", true
		case "run_worker":
			return "APP_RUN_WORKER", true
		}
	case "http":
		return "HTTP_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_")), true
	case "database":
		return "DATABASE_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_")), true
	case "scheduler":
		return strings.ToUpper(strings.ReplaceAll(key, "-", "_")), true
	case "feature":
		return strings.ToUpper(strings.ReplaceAll(key, "-", "_")), true
	case "rewrite":
		if key == "provider" {
			return "REWRITE_PROVIDER", true
		}
	case "telegram":
		switch key {
		case "bot_token":
			return "TELEGRAM_BOT_TOKEN", true
		case "target_channel":
			return "TELEGRAM_TARGET_CHANNEL", true
		case "admin_user_ids":
			return "ADMIN_USER_IDS", true
		case "publish_targets":
			return "TELEGRAM_PUBLISH_TARGETS", true
		case "ingest_targets":
			return "TELEGRAM_INGEST_TARGETS", true
		}
	case "twitter":
		return "TWITTER_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_")), true
	case "deepseek":
		return "DEEPSEEK_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_")), true
	case "gemini":
		return "GEMINI_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_")), true
	case "docker":
		switch key {
		case "image_name":
			return "DOCKER_IMAGE_NAME", true
		case "image_tag":
			return "DOCKER_IMAGE_TAG", true
		case "compose_project":
			return "COMPOSE_PROJECT_NAME", true
		}
	}
	return "", false
}
