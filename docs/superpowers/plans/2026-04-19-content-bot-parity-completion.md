# Content Bot Parity Completion Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the remaining parity work between the Go content bot and `public/content-bot-share`, then keep operator documentation accurate in `README.md`.

**Architecture:** Keep the existing Go Clean Architecture shape: domain contracts in `internal/content/domain`, use cases in `internal/content/application`, adapters in `internal/content/infrastructure`, and runtime wiring in `pkg/app/bootstrap`. Preserve DB-first runtime settings and avoid reintroducing operational config into `.env`.

**Tech Stack:** Go, Gin, GORM, Supabase Postgres, Telegram Bot API, Gemini, DeepSeek, X API.

---

## Orchestrator Brief

```yaml
task: complete remaining content-bot-share parity in Go and document usage
task_type: implementation
status: READY
current_state: core Twitter crawl -> Gemini structured rewrite -> Telegram topic publish is live; structured rewrite and Telegram admin bot parity are implemented; Telegram scope is bot-only.
plan:
  - agent: go-architect
    reason: structured rewrite and Telegram admin bot affect application contracts, settings, worker topology, and external adapter boundaries.
    entry_points:
      - internal/content/application/ports.go
      - pkg/app/bootstrap/app.go
      - internal/system/infrastructure/clients/gemini/client.go
      - internal/system/infrastructure/clients/deepseek/client.go
    required_reads:
      - public/content-bot-share/src/processor/deepseek.ts
      - public/content-bot-share/src/bot/telegram-bot.ts
    next_if_ok: go-test-writer
  - agent: go-test-writer
    reason: structured AI output, duplicate detection, and bot commands need behavior-first tests before implementation.
    entry_points:
      - internal/content/application/service_test.go
      - internal/content/infrastructure/jobs/*_test.go
      - internal/system/infrastructure/clients/gemini/client_test.go
    required_reads:
      - internal/content/domain/content_item.go
      - internal/content/infrastructure/persistence/repositories/content_repository.go
    next_if_ok: go-backend-developer
  - agent: go-backend-developer
    reason: implement contracts, jobs, clients, CLI/API hooks, and bootstrap wiring.
    entry_points:
      - internal/content/
      - internal/source/
      - internal/system/
      - cmd/cli/main.go
      - pkg/app/bootstrap/app.go
    required_reads:
      - docs/superpowers/plans/2026-04-19-content-bot-parity-completion.md
    next_if_ok: go-db-optimizer
  - agent: go-db-optimizer
    reason: validate migrations, indexes, settings jsonb shape, checkpoint keys, and queue queries.
    entry_points:
      - db/migrations/
      - internal/*/infrastructure/persistence/
    required_reads:
      - db/migrations/0001_init_content_bot_schema.sql
      - db/migrations/0014_add_settings_json_value.sql
    next_if_ok: go-code-reviewer
  - agent: go-code-reviewer
    reason: final correctness, regression, security, and runtime-safety review before claiming parity.
    entry_points:
      - internal/
      - pkg/
      - cmd/
      - db/migrations/
    required_reads:
      - public/content-bot-share/src/
    next_if_ok: go-technical-writer
  - agent: go-technical-writer
    reason: sync README, changelog, and progress after implementation waves.
    entry_points:
      - README.md
      - docs/plan/progress.md
      - changelogs/CHANGELOG.md
    required_reads:
      - docs/superpowers/plans/2026-04-19-content-bot-parity-completion.md
    next_if_ok: end
required_reads:
  - .agents/agents/go-orchestrator.md
  - docs/plan/progress.md
  - changelogs/CHANGELOG.md
risks:
  - Structured rewrite changes the AI adapter contract and can break fallback if both providers do not return compatible JSON.
  - Admin bot and current Bot API ingest share getUpdates; they must share one offset consumer to avoid stealing updates from each other.
verification:
  - go test ./...
  - go build ./cmd/api ./cmd/cli ./cmd/migrate ./cmd/worker
  - go run ./cmd/cli check-connections
  - targeted CLI probes for crawl-twitter-once, process-next, publish-next, and admin command handling where safe
docs:
  - README.md
  - docs/plan/progress.md
  - changelogs/CHANGELOG.md
```

## Current Parity State

Implemented:
- Twitter source crawl with username checkpoint initialization and `since_id` storage.
- Twitter publish with Vietnamese/English account support and `tweet_vi_id` / `tweet_en_id` persistence.
- Telegram Bot API topic-aware ingest and publish for configured group/topic targets.
- Gemini rewrite provider, DeepSeek fallback adapter, live end-to-end demo through Telegram.
- Structured AI rewrite response matching the sample JSON contract, including `shouldPublish=false` skips.
- Pre-rewrite and post-rewrite similarity duplicate detection.
- Telegram admin bot command surface inside the existing Bot API ingest offset stream.
- Operator-facing admin views for status, queue, recent content, sources, logs, retry, skip, pause/resume, and crawlnow.
- DB-first runtime settings, including `telegram_runtime` as `settings.json_value`.
- DB-first scheduler interval settings for crawl, rewrite, Telegram publish, and Twitter publish loops.
- Queue row locking, stale queue skipping, source tags/topics, source health revalidation.

Not yet implemented:
- No additional Telegram runtime mode is planned; the project standardizes on Bot API chats/topics where the bot is present.

## Wave 1: Structured Rewrite Parity

**Files:**
- Modify: `internal/content/application/ports.go`
- Modify: `internal/content/application/service.go`
- Modify: `internal/content/domain/content_item.go`
- Modify: `internal/system/infrastructure/clients/gemini/client.go`
- Modify: `internal/system/infrastructure/clients/deepseek/client.go`
- Modify: `internal/content/infrastructure/jobs/fallback_rewriter.go`
- Modify: `internal/content/infrastructure/persistence/repositories/content_repository.go`
- Test: `internal/content/application/service_test.go`
- Test: `internal/system/infrastructure/clients/gemini/client_test.go`
- Test: `internal/content/infrastructure/jobs/fallback_rewriter_test.go`

- [x] Define `RewriteResult` with `RewrittenText`, `RewrittenTextEn`, `TweetVI`, `TweetEN`, `FactCheckNote`, `ShouldPublish`, and `Reason`.
- [x] Replace `RewriteTextPort` string-only contract with structured rewrite while keeping a compatibility shim if needed for small tests.
- [x] Update Gemini and DeepSeek prompts to request JSON with the same fields as `public/content-bot-share/src/processor/deepseek.ts`.
- [x] Parse JSON responses robustly, including fallback extraction from fenced or surrounding text.
- [x] Persist all structured fields into existing `content_items` columns.
- [x] If `shouldPublish=false`, mark the item `failed` or `skipped` with `fail_reason`; use `skipped` for expected editorial rejection and `failed` for provider/runtime errors.
- [x] Add pre-rewrite duplicate detection against recent processed items before spending AI tokens.
- [x] Add post-rewrite duplicate detection against recent rewritten/published content.
- [x] Make duplicate windows and thresholds settings-driven:
  - `rewrite_duplicate_window_hours`
  - `rewrite_duplicate_original_threshold`
  - `rewrite_duplicate_rewritten_threshold`
- [x] Verify `process-next` still claims atomically and newest-first after the contract change.
- [x] Run `go test ./internal/content/... ./internal/system/infrastructure/clients/...`.

Acceptance:
- A Gemini rewrite produces all available DB fields, not just `rewritten_text`.
- `tweet_text_vi` is used by Twitter publisher when present.
- `shouldPublish=false` prevents Telegram/Twitter publish and records a clear reason.
- Duplicate items can be skipped before AI call.

## Wave 2: Telegram Admin Bot Parity

**Files:**
- Create: `internal/system/infrastructure/jobs/telegram_admin_action.go`
- Create: `internal/system/infrastructure/jobs/telegram_admin_action_test.go`
- Modify: `internal/system/infrastructure/clients/telegrambot/client.go`
- Modify: `internal/content/application/service.go`
- Modify: `internal/content/domain/content_item.go`
- Create: `internal/system/infrastructure/persistence/models/log_model.go`
- Create: `internal/system/infrastructure/persistence/repositories/log_repository.go`
- Modify: `internal/content/infrastructure/jobs/telegram_ingest_action.go`
- Modify: `pkg/app/bootstrap/app.go`

- [x] Avoid a second independent `getUpdates` consumer. Admin command handling is integrated before `TelegramIngestAction` content enqueue.
- [x] Add command parsing for:
  - `/start`
  - `/status`
  - `/queue`
  - `/recent`
  - `/sources`
  - `/addsource <type> <handle> <name>`
  - `/removesource <handle>`
  - `/retry`
  - `/pause`
  - `/resume`
  - `/logs`
  - `/skip <id-prefix>`
  - `/crawlnow`
- [x] Enforce admin access using `telegram_runtime.admin_user_ids`; if empty, preserve sample behavior and allow all.
- [x] Store `/pause` and `/resume` through `settings.auto_publish`.
- [x] Make Telegram publishing read `settings.auto_publish` at runtime so `/pause` and `/resume` do not require worker restart.
- [x] Implement source removal as deactivate, not hard delete.
- [x] Add retry and skip application methods for failed items and ID-prefix lookup using existing repository seams.
- [x] Use plain-text admin replies so Markdown parse fallback is not needed for this wave.
- [x] Add tests for command authorization, command parsing, settings writes, source/content updates, and ingest offset integration.

Acceptance:
- An admin can operate the bot from Telegram without CLI for normal runtime tasks.
- Bot commands do not steal update offsets from content ingest.
- `auto_publish` can be toggled and observed through `/status`.

## Wave 3: README And Operator Docs

**Files:**
- Modify: `README.md`
- Modify: `docs/plan/progress.md`
- Modify: `changelogs/CHANGELOG.md`

- [x] Replace the stale template README with content-bot-specific setup and operations.
- [x] Document architecture, binaries, settings, JSONB Telegram runtime, source management, worker modes, CLI, API, Docker, and demo flow.
- [x] Document Telegram as a Bot API-only operational surface for this project.
- [x] Update progress after each completed wave.
- [ ] Update changelog once this repository has a tracked changelog file.
- [x] Update changelog once this repository has a tracked changelog file.

Acceptance:
- A new operator can configure Supabase, AI provider, Telegram, Twitter, settings, and run a demo without reading internal code.
