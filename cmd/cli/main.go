package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	contentdomain "go-content-bot/internal/content/domain"
	sourcedomain "go-content-bot/internal/source/domain"
	"go-content-bot/pkg/app/bootstrap"
	"go-content-bot/pkg/buildinfo"
	"go-content-bot/pkg/clihelp"
	"go-content-bot/pkg/config"
)

func main() {
	command := "check-connections"
	args := []string{}
	if len(os.Args) > 1 {
		command = os.Args[1]
		args = os.Args[2:]
	}

	if isHelpCommand(command) {
		fmt.Print(renderHelp())
		return
	}

	if command == "check-connections" {
		results, err := bootstrap.CheckConnectionsOnly(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		encoded, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
		return
	}
	if command == "show-runtime" {
		details, err := config.LoadDetailsFromPaths()
		if err != nil {
			log.Fatal(err)
		}

		encoded, err := json.MarshalIndent(details.Config.RuntimeSummary(details.LoadedPaths), "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
		return
	}
	if command == "version" {
		encoded, err := json.MarshalIndent(buildinfo.Current("cli.exe"), "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
		return
	}

	app, err := bootstrap.New()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("cli shutdown error: %v", err)
		}
	}()

	switch command {
	case "process-next":
		result, err := app.RunProcessNext(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		if result == nil {
			fmt.Println(`{"status":"idle","message":"no pending items"}`)
			return
		}

		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	case "publish-next":
		result, err := app.RunPublishNext(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		if result == nil {
			fmt.Println(`{"status":"idle","message":"no rewritten items"}`)
			return
		}

		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	case "ingest-telegram-once":
		if err := app.RunTelegramIngestOnce(context.Background()); err != nil {
			log.Fatal(err)
		}
		fmt.Println(`{"status":"ok","message":"telegram ingest completed"}`)
	case "crawl-twitter-once":
		if err := app.RunTwitterCrawlOnce(context.Background()); err != nil {
			log.Fatal(err)
		}
		fmt.Println(`{"status":"ok","message":"twitter crawl completed"}`)
	case "revalidate-sources-once":
		if err := app.RunSourceRevalidationOnce(context.Background()); err != nil {
			log.Fatal(err)
		}
		fmt.Println(`{"status":"ok","message":"source revalidation completed"}`)
	case "publish-twitter-next":
		if err := app.RunTwitterPublishNext(context.Background()); err != nil {
			log.Fatal(err)
		}
		fmt.Println(`{"status":"ok","message":"twitter publish completed"}`)
	case "probe-telegram-targets":
		message := ""
		if len(args) > 0 {
			message = strings.TrimSpace(strings.Join(args, " "))
		}
		results, err := app.RunProbeTelegramTargets(context.Background(), message)
		if err != nil {
			log.Fatal(err)
		}
		encoded, err := json.MarshalIndent(map[string]any{"items": results}, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	case "report-twitter-sources":
		activeOnly, inactiveOnly := parseTwitterSourceFlags(args)
		if activeOnly && inactiveOnly {
			log.Fatal("use only one of --active-only or --inactive-only")
		}
		sources, err := app.SourceService.ListAll(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		filtered := make([]map[string]any, 0, len(sources))
		for _, source := range sources {
			if source.Type != sourcedomain.TypeTwitter {
				continue
			}
			if activeOnly && !source.IsActive {
				continue
			}
			if inactiveOnly && source.IsActive {
				continue
			}
			filtered = append(filtered, map[string]any{
				"id":              source.ID,
				"type":            source.Type,
				"handle":          source.Handle,
				"name":            source.Name,
				"tags":            source.Tags,
				"topics":          source.Topics,
				"is_active":       source.IsActive,
				"last_crawled_at": source.LastCrawledAt,
				"last_check_at":   source.LastCheckAt,
				"last_error":      source.LastError,
				"created_at":      source.CreatedAt,
			})
		}
		encoded, err := json.MarshalIndent(map[string]any{"items": filtered}, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	case "report-content-ops":
		limit := parseContentOpsLimit(args, 50)
		items, err := app.ContentService.ListRecent(context.Background(), limit)
		if err != nil {
			log.Fatal(err)
		}
		encoded, err := json.MarshalIndent(buildContentOpsReport(items), "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	case "settings-get":
		if len(args) != 1 {
			log.Fatal("usage: settings-get <key>")
		}
		record, err := app.RunGetSetting(context.Background(), strings.TrimSpace(args[0]))
		if err != nil {
			log.Fatal(err)
		}
		encoded, err := json.MarshalIndent(map[string]any{
			"key":         args[0],
			"value":       record.Value,
			"json_value":  decodeSettingJSONValue(record.JSONValue),
			"description": record.Description,
		}, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	case "settings-set":
		if len(args) < 2 {
			log.Fatal("usage: settings-set <key> <value>")
		}
		key := strings.TrimSpace(args[0])
		value := strings.TrimSpace(strings.Join(args[1:], " "))
		if err := app.RunSetSetting(context.Background(), key, value); err != nil {
			log.Fatal(err)
		}
		encoded, err := json.MarshalIndent(map[string]any{"key": key, "value": value}, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	case "settings-set-json":
		if len(args) < 2 {
			log.Fatal("usage: settings-set-json <key> <json>")
		}
		key := strings.TrimSpace(args[0])
		raw := strings.TrimSpace(strings.Join(args[1:], " "))
		if !json.Valid([]byte(raw)) {
			log.Fatal("settings-set-json requires valid JSON")
		}
		if err := app.RunSetSettingJSON(context.Background(), key, []byte(raw)); err != nil {
			log.Fatal(err)
		}
		encoded, err := json.MarshalIndent(map[string]any{"key": key, "json_value_set": true}, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	default:
		log.Fatalf("unknown command: %s", command)
	}
}

func decodeSettingJSONValue(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return string(raw)
	}
	return decoded
}

func parseTwitterSourceFlags(args []string) (activeOnly bool, inactiveOnly bool) {
	for _, arg := range args {
		switch arg {
		case "--active-only":
			activeOnly = true
		case "--inactive-only":
			inactiveOnly = true
		}
	}
	return activeOnly, inactiveOnly
}

type contentOpsReport struct {
	GeneratedAt            string                 `json:"generated_at"`
	Limit                  int                    `json:"limit"`
	Counts                 map[string]int         `json:"counts"`
	RewrittenReady         []contentOpsReportItem `json:"rewritten_ready"`
	ManualDuplicateSkipped []contentOpsReportItem `json:"manual_duplicate_skipped"`
	PublishedRecent        []contentOpsReportItem `json:"published_recent"`
}

type contentOpsReportItem struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	AuthorName   string  `json:"author_name"`
	OriginalText string  `json:"original_text"`
	FailReason   *string `json:"fail_reason,omitempty"`
	CrawledAt    string  `json:"crawled_at"`
	PublishedAt  *string `json:"published_at,omitempty"`
}

func parseContentOpsLimit(args []string, fallback int) int {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--limit=") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(arg, "--limit="))
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return fallback
		}
		return value
	}
	return fallback
}

func isHelpCommand(command string) bool {
	switch command {
	case "help", "-h", "--help":
		return true
	default:
		return false
	}
}

func renderHelp() string {
	return clihelp.Render(clihelp.Document{
		Binary:         "cli.exe",
		Description:    "operator tool",
		DefaultCommand: "check-connections",
		Commands: []clihelp.Command{
			{Name: "check-connections", Description: "check database and provider connectivity"},
			{Name: "version", Description: "print build and version info"},
			{Name: "show-runtime", Description: "print effective runtime config"},
			{Name: "process-next", Description: "rewrite the next pending content item"},
			{Name: "publish-next", Description: "publish the next rewritten content item"},
			{Name: "ingest-telegram-once", Description: "run one Telegram ingest pass"},
			{Name: "crawl-twitter-once", Description: "run one Twitter crawl pass"},
			{Name: "revalidate-sources-once", Description: "recheck configured sources once"},
			{Name: "publish-twitter-next", Description: "publish the next Twitter item"},
			{Name: "probe-telegram-targets", Description: "send a probe message to publish targets"},
			{Name: "report-twitter-sources", Description: "print current Twitter source report"},
			{Name: "report-content-ops", Description: "print queue and publish report"},
			{Name: "settings-get", Description: "read a runtime setting"},
			{Name: "settings-set", Description: "write a string runtime setting"},
			{Name: "settings-set-json", Description: "write a JSON runtime setting"},
			{Name: "help", Description: "show this help"},
		},
	})
}

func buildContentOpsReport(items []contentdomain.ContentItem) contentOpsReport {
	report := contentOpsReport{
		GeneratedAt:            time.Now().Format(time.RFC3339),
		Limit:                  len(items),
		Counts:                 map[string]int{"rewritten_ready": 0, "manual_duplicate_skipped": 0, "published_recent": 0},
		RewrittenReady:         make([]contentOpsReportItem, 0),
		ManualDuplicateSkipped: make([]contentOpsReportItem, 0),
		PublishedRecent:        make([]contentOpsReportItem, 0),
	}
	for _, item := range items {
		entry := toContentOpsReportItem(item)
		switch {
		case item.Status == contentdomain.StatusRewritten:
			report.Counts["rewritten_ready"]++
			report.RewrittenReady = append(report.RewrittenReady, entry)
		case item.Status == contentdomain.StatusSkipped && strings.TrimSpace(derefString(item.FailReason)) == "duplicate manual content already processed recently":
			report.Counts["manual_duplicate_skipped"]++
			report.ManualDuplicateSkipped = append(report.ManualDuplicateSkipped, entry)
		case item.Status == contentdomain.StatusPublished:
			report.Counts["published_recent"]++
			report.PublishedRecent = append(report.PublishedRecent, entry)
		}
	}
	return report
}

func toContentOpsReportItem(item contentdomain.ContentItem) contentOpsReportItem {
	entry := contentOpsReportItem{
		ID:           item.ID,
		Status:       string(item.Status),
		AuthorName:   item.AuthorName,
		OriginalText: item.OriginalText,
		FailReason:   item.FailReason,
		CrawledAt:    item.CrawledAt.Format(time.RFC3339),
	}
	if item.PublishedAt != nil {
		value := item.PublishedAt.Format(time.RFC3339)
		entry.PublishedAt = &value
	}
	return entry
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
