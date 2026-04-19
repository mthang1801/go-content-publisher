package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
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
	Name         string
	Env          string
	LogLevel     string
	LogMaxSizeMB int
	AutoMigrate  bool
	RunAPI       bool
	RunWorker    bool
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

type LoadDetails struct {
	Config      Config
	LoadedPaths []string
}

type RuntimeSummary struct {
	App         RuntimeAppSummary      `json:"app"`
	HTTP        RuntimeHTTPSummary     `json:"http"`
	Database    RuntimeDatabaseSummary `json:"database"`
	ConfigFiles []string               `json:"config_files"`
}

type RuntimeAppSummary struct {
	Name         string `json:"name"`
	Env          string `json:"env"`
	LogMaxSizeMB int    `json:"log_max_size_mb"`
	AutoMigrate  bool   `json:"auto_migrate"`
	RunAPI       bool   `json:"run_api"`
	RunWorker    bool   `json:"run_worker"`
}

type RuntimeHTTPSummary struct {
	Host       string `json:"host"`
	PortConfig string `json:"port_config"`
	PortMode   string `json:"port_mode"`
}

type RuntimeDatabaseSummary struct {
	Target       string `json:"target"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	Name         string `json:"name"`
	Mode         string `json:"mode"`
	SSLMode      string `json:"ssl_mode"`
	ConfigSource string `json:"config_source"`
}

func Load(paths ...string) (Config, error) {
	details, err := LoadDetailsFromPaths(paths...)
	if err != nil {
		return Config{}, err
	}
	return details.Config, nil
}

func LoadDetailsFromPaths(paths ...string) (LoadDetails, error) {
	loadedPaths := make([]string, 0, 4)
	for _, path := range choosePaths(paths) {
		if _, err := os.Stat(path); err == nil {
			absolutePath := absolutePathOrOriginal(path)
			if !containsPath(loadedPaths, absolutePath) {
				loadedPaths = append(loadedPaths, absolutePath)
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".ini":
				if err := loadINI(path); err != nil {
					return LoadDetails{}, fmt.Errorf("load ini file %s: %w", path, err)
				}
			default:
				if err := godotenv.Overload(path); err != nil {
					return LoadDetails{}, fmt.Errorf("load env file %s: %w", path, err)
				}
			}
		}
	}

	cfg := Config{
		App: AppConfig{
			Name:         getOrDefault("APP_NAME", "content-bot-go"),
			Env:          getOrDefault("APP_ENV", "development"),
			LogLevel:     getOrDefault("APP_LOG_LEVEL", "info"),
			LogMaxSizeMB: mustInt("APP_LOG_MAX_SIZE_MB", 10),
			AutoMigrate:  mustBool("APP_AUTO_MIGRATE", false),
			RunAPI:       mustBool("APP_RUN_API", true),
			RunWorker:    mustBool("APP_RUN_WORKER", true),
		},
		HTTP: HTTPConfig{
			Host:         getOrDefault("HTTP_HOST", "0.0.0.0"),
			Port:         getOrDefaultPreserveEmpty("HTTP_PORT", "8080"),
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
		return LoadDetails{}, err
	}

	cfg.Telegram.applyLegacyFallbacks()

	return LoadDetails{
		Config:      cfg,
		LoadedPaths: loadedPaths,
	}, nil
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

	defaults := []string{"config/config.ini", "config.ini", "config/.env", ".env"}
	candidates := append([]string(nil), defaults...)

	if executable, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executable)
		for _, path := range defaults {
			candidate := filepath.Join(executableDir, path)
			if containsPath(candidates, candidate) {
				continue
			}
			candidates = append(candidates, candidate)
		}
	}

	return candidates
}

func containsPath(paths []string, candidate string) bool {
	for _, path := range paths {
		if path == candidate {
			return true
		}
	}
	return false
}

func absolutePathOrOriginal(path string) string {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absolutePath
}

func getOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getOrDefaultPreserveEmpty(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return strings.TrimSpace(value)
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
		case "log_max_size_mb":
			return "APP_LOG_MAX_SIZE_MB", true
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

func (c Config) RuntimeSummary(loadedPaths []string) RuntimeSummary {
	host, port, name, mode, configSource := c.Database.runtimeDetails()
	return RuntimeSummary{
		App: RuntimeAppSummary{
			Name:         c.App.Name,
			Env:          c.App.Env,
			LogMaxSizeMB: c.App.LogMaxSizeMB,
			AutoMigrate:  c.App.AutoMigrate,
			RunAPI:       c.App.RunAPI,
			RunWorker:    c.App.RunWorker,
		},
		HTTP: RuntimeHTTPSummary{
			Host:       c.HTTP.Host,
			PortConfig: c.HTTP.Port,
			PortMode:   normalizeHTTPPortMode(c.HTTP.Port),
		},
		Database: RuntimeDatabaseSummary{
			Target:       databaseTarget(host, port, name),
			Host:         host,
			Port:         port,
			Name:         name,
			Mode:         mode,
			SSLMode:      c.Database.SSLMode,
			ConfigSource: configSource,
		},
		ConfigFiles: append([]string(nil), loadedPaths...),
	}
}

func normalizeHTTPPortMode(port string) string {
	port = strings.TrimSpace(port)
	if port == "" || port == "0" || strings.EqualFold(port, "auto") {
		return "auto"
	}
	return "fixed"
}

func (c DatabaseConfig) runtimeDetails() (host, port, name, mode, configSource string) {
	if strings.TrimSpace(c.URL) != "" {
		host, port, name = parseDatabaseURL(c.URL)
		mode = inferDatabaseMode(host, port)
		return host, port, name, mode, "DATABASE_URL"
	}

	host = strings.TrimSpace(c.Host)
	port = strings.TrimSpace(c.Port)
	name = strings.TrimSpace(c.Name)
	mode = inferDatabaseMode(host, port)
	return host, port, name, mode, "DATABASE_HOST/DATABASE_NAME"
}

func parseDatabaseURL(raw string) (host, port, name string) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", "", ""
	}

	host = strings.TrimSpace(parsed.Hostname())
	port = strings.TrimSpace(parsed.Port())
	name = strings.TrimPrefix(strings.TrimSpace(parsed.Path), "/")
	if port == "" {
		port = "5432"
	}
	return host, port, name
}

func inferDatabaseMode(host, port string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	port = strings.TrimSpace(port)

	switch {
	case host == "" && port == "":
		return "unknown"
	case host == "localhost" || host == "127.0.0.1" || host == "::1":
		return "local"
	case strings.HasSuffix(host, ".pooler.supabase.com") && port == "5432":
		return "supabase_session_pooler"
	case strings.HasSuffix(host, ".pooler.supabase.com") && port == "6543":
		return "supabase_transaction_pooler"
	case strings.HasSuffix(host, ".supabase.co"):
		return "supabase_direct"
	default:
		return "custom"
	}
}

func databaseTarget(host, port, name string) string {
	parts := make([]string, 0, 2)
	address := strings.TrimSpace(host)
	if address != "" && strings.TrimSpace(port) != "" {
		address = address + ":" + strings.TrimSpace(port)
	}
	if address != "" {
		parts = append(parts, address)
	}
	if strings.TrimSpace(name) != "" {
		parts = append(parts, strings.TrimSpace(name))
	}
	return strings.Join(parts, "/")
}
