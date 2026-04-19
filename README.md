# Go Content Bot

Go Content Bot is a Go/Gin/GORM rewrite of the TypeScript sample in `public/content-bot-share`.

The bot crawls configured sources, stores content in Supabase Postgres, rewrites content through an AI provider, publishes to Telegram group topics, and can optionally publish to X/Twitter.

## Current Status

Implemented and verified:

- Supabase Postgres persistence through GORM.
- HTTP API, CLI, worker, and migration binaries.
- Twitter/X crawler with username-based `since_id` checkpoints.
- Twitter/X publish with `tweet_vi_id` and `tweet_en_id` persistence.
- Telegram Bot API topic-aware ingest and publish.
- Telegram runtime config stored in `settings.json_value` as JSONB.
- Gemini rewrite provider with DeepSeek fallback wiring.
- Structured rewrite JSON contract for `rewrittenText`, `rewrittenTextEn`, `tweetVI`, `tweetEN`, `factCheckNote`, `shouldPublish`, and `reason`.
- Similarity duplicate detection before and after rewrite.
- Telegram admin bot commands integrated into the existing Bot API ingest offset stream.
- Telegram bot-only runtime model for ingest, admin commands, and publish.
- DB-first runtime settings for worker controls and publish policy.
- Queue locking with `FOR UPDATE SKIP LOCKED`.
- Source tags/topics, source health metadata, and 24-hour inactive source revalidation.
- Dockerfile and Docker Compose deployment artifacts.

Still planned for full `content-bot-share` parity:

- Operator and documentation polish around the bot-only runtime path.
- Additional queue and publish policy tuning where needed for safe demos.

The detailed remaining implementation plan is in `docs/superpowers/plans/2026-04-19-content-bot-parity-completion.md`.

## Architecture

The code follows Clean Architecture:

```text
Presentation -> Application -> Domain <- Infrastructure
```

Main folders:

- `cmd/api`: Gin HTTP API.
- `cmd/worker`: scheduled crawl, rewrite, Telegram publish, and Twitter publish loops.
- `cmd/migrate`: SQL migration runner.
- `cmd/cli`: operational one-shot commands.
- `internal/content`: content queue, rewrite, publish, and crawler jobs.
- `internal/source`: source management and source revalidation.
- `internal/system`: external clients and settings persistence.
- `pkg/config`: environment-backed bootstrap defaults.
- `pkg/app/bootstrap`: runtime wiring and DB-first settings overrides.
- `db/migrations`: Supabase/Postgres SQL migrations.

## Runtime Model

The worker runs four independent loops:

- Crawl loop: source revalidation, Telegram Bot API ingest, Twitter crawl.
- Rewrite loop: pending content to rewritten content through the configured AI provider.
- Telegram publish loop: rewritten content to Telegram group/topic when `auto_publish=true`.
- Twitter publish loop: Telegram-published content to X/Twitter when Twitter publish flags and policy allow it.

Important distinction:

- `auto_publish=false` does not stop crawling or rewriting.
- `auto_publish=false` only stops automatic Telegram publishing.
- Twitter publish is controlled separately by `enable_twitter_publish_vi`, `enable_twitter_publish_en`, and Twitter publish policy settings.

## Configuration

The project is now mostly DB-first for runtime behavior.

Keep infrastructure credentials and static bootstrap defaults in `config/.env`. Keep runtime toggles and routing policy in the `settings` table.

### Environment

Create `config/.env` from `config/.env.example`, then fill:

```dotenv
APP_NAME=content-bot-go
APP_ENV=development
APP_LOG_LEVEL=info

HTTP_HOST=0.0.0.0
HTTP_PORT=8080

DATABASE_URL=postgres://postgres:[YOUR-PASSWORD]@db.[YOUR-PROJECT-REF].supabase.co:5432/postgres?sslmode=require

TWITTER_BEARER_TOKEN=
TWITTER_VI_API_KEY=
TWITTER_VI_API_SECRET=
TWITTER_VI_ACCESS_TOKEN=
TWITTER_VI_ACCESS_SECRET=
TWITTER_EN_API_KEY=
TWITTER_EN_API_SECRET=
TWITTER_EN_ACCESS_TOKEN=
TWITTER_EN_ACCESS_SECRET=

DEEPSEEK_API_KEY=
DEEPSEEK_MODEL=deepseek-chat

GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.5-flash
```

Telegram Bot token, targets, and scheduler intervals are no longer stored in `.env`; use the `telegram_runtime` and interval settings below.

### Operator Config Matrix

Use this table as the fastest rule of thumb:

| Concern | Store In | Example Keys | Restart Required |
| --- | --- | --- | --- |
| Process bootstrap and local runtime shape | `config/.env` or `config.ini` | `DATABASE_URL`, `HTTP_PORT`, `APP_LOG_LEVEL`, `auto_migrate`, `run_api`, `run_worker` | Yes |
| AI and upstream credentials | `config/.env` or `config.ini` | `GEMINI_API_KEY`, `DEEPSEEK_API_KEY`, Twitter API tokens | Yes |
| Telegram bot runtime contract | `settings.json_value` | `telegram_runtime.bot_token`, `publish_targets`, `ingest_targets`, `admin_user_ids` | Yes |
| Live Telegram auto-post switch | `settings.value` | `auto_publish` | No |
| Rewrite provider selection | `settings.value` | `rewrite_provider` | Yes |
| Worker feature flags | `settings.value` | `enable_telegram_crawler`, `enable_twitter_crawler`, `enable_rewrite_processor`, `enable_twitter_publish_vi`, `enable_twitter_publish_en` | Yes |
| Worker loop intervals | `settings.value` | `crawl_interval_seconds`, `process_interval_seconds`, `publish_interval_seconds`, `twitter_publish_interval_seconds` | Yes |
| Duplicate, stale queue, and publish policy | `settings.value` | `rewrite_duplicate_*`, `pending_stale_after_seconds`, `rewritten_stale_after_seconds`, `twitter_publish_*` | Yes unless noted otherwise in code/docs |

Rule:

- if a value is a secret or basic process bootstrap dependency, keep it in `config/.env` or `config.ini`
- if a value changes crawling, rewriting, routing, or publish behavior, keep it in `settings`
- right now only `auto_publish` is explicitly designed for live runtime changes without restart

### Telegram Runtime JSONB

Telegram runtime config lives in `settings` under key `telegram_runtime`, column `json_value`.

Read it:

```bash
go run ./cmd/cli settings-get telegram_runtime
```

Set it:

```bash
go run ./cmd/cli settings-set-json telegram_runtime '{"bot_token":"<telegram-bot-token>","publish_targets":[{"chat_id":"-1002451344189","thread_id":5}],"ingest_targets":[{"chat_id":"-1002451344189","thread_id":5}],"admin_user_ids":[]}'
```

Fields:

- `bot_token`: Telegram Bot API token.
- `publish_targets`: list of Telegram destinations for rewritten content.
- `ingest_targets`: list of Telegram chats/topics the Bot API ingest path accepts.
- `admin_user_ids`: admin bot allowlist; empty means allow all, matching the TypeScript sample behavior.

Target shape:

- Regular Telegram group or channel: only `chat_id` is needed.
- Forum supergroup topic: both `chat_id` and `thread_id` are needed.
- If a target is a normal group and you send `thread_id`, Telegram may reject it.
- If a target is a forum topic and you omit `thread_id`, the bot will post to the main chat instead of the intended topic, or may fail depending on permissions and chat shape.

Restart API/worker after changing `telegram_runtime`; Telegram clients and targets are created during bootstrap.

### Telegram Target Examples

Regular group:

```json
{
  "chat_id": "-1005268400327"
}
```

Forum topic:

```json
{
  "chat_id": "-1002451344189",
  "thread_id": 5
}
```

Typical `telegram_runtime` for a normal group:

```bash
go run ./cmd/cli settings-set-json telegram_runtime '{"bot_token":"<telegram-bot-token>","publish_targets":[{"chat_id":"-1005268400327"}],"ingest_targets":[{"chat_id":"-1005268400327"}],"admin_user_ids":[]}'
```

### Telegram Admin Bot

Admin commands are handled inside the same `getUpdates` loop as Telegram ingest. This avoids a second Bot API consumer stealing offsets.

Commands are accepted only from configured `telegram_runtime.ingest_targets`. If `telegram_runtime.admin_user_ids` is empty, every sender in the target chat/topic is allowed; otherwise the sender `from.id` must be in that allowlist.

Supported commands:

```text
/start
/help
/add <text>
/status
/queue
/recent
/sources
/addsource <twitter|telegram> <handle> [name]
/removesource [twitter|telegram] <handle>
/retry
/skip <content-id-prefix>
/pause
/resume
/logs
/crawlnow
```

Important behavior:

- `/pause` writes `settings.auto_publish=false`; it does not stop crawl or rewrite.
- `/resume` writes `settings.auto_publish=true`; the Telegram publish loop reads this setting at runtime.
- `/add <text>` pushes manual content straight into the pending queue.
- Any non-command text from an authorized admin inside `ingest_targets` is also queued as manual content.
- Manual/admin items now bypass fuzzy similarity dedupe during rewrite; only exact normalized duplicates are skipped early, while Telegram publish still blocks exact recent reposts.
- `/addsource` reactivates an existing inactive source when the same type/handle already exists.
- `/removesource` marks a source inactive instead of hard deleting it.
- `/retry` moves recent failed items back to `pending`.
- `/skip` marks one matching recent queue item as skipped; ambiguous prefixes are rejected.
- `/crawlnow` runs source revalidation and Twitter crawl. It intentionally does not recursively call Telegram Bot API ingest.
- `/logs` reads from the `logs` table when persisted logs exist.

## Glossary

Common terms used in this project:

- `source`: one upstream content origin, for example an X account or Telegram source.
- `twitter source`: a source with `type=twitter`, usually stored by username such as `@zerohedge`.
- `telegram source`: a source with `type=telegram`.
- `Bot API ingest`: Telegram reading path that only works for chats/topics where the bot actually receives updates.
- `ingest_targets`: Telegram chats/topics where the bot is allowed to read updates and admin commands.
- `publish_targets`: Telegram chats/topics where rewritten content is posted.
- `chat_id`: Telegram destination identifier, usually a negative ID for groups/channels.
- `thread_id`: Telegram forum topic ID inside a forum-enabled supergroup.
- `regular group`: a Telegram group without topics; use `chat_id` only.
- `forum topic`: a topic inside a forum-enabled supergroup; use both `chat_id` and `thread_id`.
- `pending`: newly crawled content waiting for rewrite.
- `processing`: content currently claimed by the rewrite step.
- `rewritten`: content already rewritten and ready to publish.
- `publishing`: content currently claimed by the Telegram publish step.
- `published`: content already sent to Telegram successfully.
- `failed`: content that hit a runtime or provider error.
- `skipped`: content intentionally not processed further, for example stale or duplicate items.
- `auto_publish`: runtime flag controlling only Telegram auto-post from `rewritten` items.
- `rewrite provider`: the primary AI adapter used for rewrite, usually `gemini` or `deepseek`.
- `fallback`: secondary provider called only when the primary provider fails.
- `checkpoint`: last seen upstream ID stored in `settings`, for example `tw_since_<handle>`.
- `stale pending`: a pending item older than `pending_stale_after_seconds`, automatically skipped so fresh news can move first.
- `stale rewritten`: a rewritten item older than `rewritten_stale_after_seconds`, automatically skipped so late posts do not go out.
- `tags` and `topics`: source metadata used for publish routing and filtering.
- `one-shot command`: a CLI command that runs a single step once, for example `crawl-twitter-once`.
- `worker`: the long-running process that executes crawl, rewrite, and publish loops automatically.

## Settings

Use CLI commands:

```bash
go run ./cmd/cli settings-get <key>
go run ./cmd/cli settings-set <key> <value>
go run ./cmd/cli settings-set-json <key> '<json>'
```

Core runtime settings:

| Key | Meaning |
| --- | --- |
| `auto_publish` | Enables automatic Telegram publish from `rewritten` items. |
| `enable_telegram_crawler` | Enables Telegram Bot API ingest loop. |
| `enable_twitter_crawler` | Enables Twitter/X crawl loop. |
| `enable_rewrite_processor` | Enables AI rewrite loop. |
| `enable_twitter_publish_vi` | Enables Vietnamese X/Twitter publish. |
| `enable_twitter_publish_en` | Enables English X/Twitter publish. |
| `crawl_interval_seconds` | Worker crawl loop interval in seconds. |
| `process_interval_seconds` | Worker rewrite loop interval in seconds. |
| `publish_interval_seconds` | Worker Telegram publish loop interval in seconds. |
| `twitter_publish_interval_seconds` | Worker Twitter publish loop interval in seconds. |
| `rewrite_provider` | Primary rewrite provider, usually `gemini` or `deepseek`. |
| `rewrite_duplicate_window_hours` | Lookback window for duplicate detection before/after rewrite. |
| `rewrite_duplicate_original_threshold` | Similarity threshold for skipping original-text duplicates before AI calls. |
| `rewrite_duplicate_rewritten_threshold` | Similarity threshold for skipping rewritten duplicates after AI calls. |
| `pending_stale_after_seconds` | Skips old pending items so fresh content can be processed first. |
| `rewritten_stale_after_seconds` | Skips old rewritten items so stale content is not posted late. |
| `twitter_publish_after` | Only tweet items published after this RFC3339 timestamp. |
| `twitter_publish_source_types` | CSV of eligible source types, for example `twitter`. |
| `twitter_publish_source_topics` | CSV of eligible source topics; takes priority over source tags. |
| `twitter_publish_source_tags` | CSV of eligible source tags when topics are empty. |
| `twitter_publish_topic_keywords` | Optional text keyword filter for Twitter publish. |
Example develop/demo mode:

```bash
go run ./cmd/cli settings-set enable_twitter_crawler true
go run ./cmd/cli settings-set enable_rewrite_processor true
go run ./cmd/cli settings-set auto_publish false
go run ./cmd/cli settings-set crawl_interval_seconds 300
go run ./cmd/cli settings-set process_interval_seconds 30
go run ./cmd/cli settings-set publish_interval_seconds 10
go run ./cmd/cli settings-set twitter_publish_interval_seconds 600
go run ./cmd/cli settings-set enable_twitter_publish_vi true
go run ./cmd/cli settings-set enable_twitter_publish_en false
```

This crawls and rewrites automatically, does not auto-post to Telegram, and only allows Vietnamese Twitter publishing when eligible.

### Settings Cookbook

These presets are short operator starting points. Apply them, then restart API/worker unless you only changed `auto_publish`.

Scripts:

```bash
./scripts/settings-preset-demo.sh
./scripts/settings-preset-safe-production.sh
./scripts/settings-preset-twitter-only.sh
```

#### Demo

Use when you want a fast live demo with Twitter crawl, rewrite, and Telegram publish enabled.

Shortcut:

```bash
./scripts/settings-preset-demo.sh
```

```bash
go run ./cmd/cli settings-set enable_twitter_crawler true
go run ./cmd/cli settings-set enable_telegram_crawler true
go run ./cmd/cli settings-set enable_rewrite_processor true
go run ./cmd/cli settings-set auto_publish true
go run ./cmd/cli settings-set crawl_interval_seconds 300
go run ./cmd/cli settings-set process_interval_seconds 30
go run ./cmd/cli settings-set publish_interval_seconds 10
go run ./cmd/cli settings-set twitter_publish_interval_seconds 600
go run ./cmd/cli settings-set rewrite_provider gemini
go run ./cmd/cli settings-set enable_twitter_publish_vi false
go run ./cmd/cli settings-set enable_twitter_publish_en false
```

Why:

- shows the full `crawl -> rewrite -> Telegram publish` path quickly
- avoids accidental outbound Twitter posting during demos

#### Safe Production

Use when you want stable crawling and rewrite in production, but human review before Telegram or Twitter distribution.

Shortcut:

```bash
./scripts/settings-preset-safe-production.sh
```

```bash
go run ./cmd/cli settings-set enable_twitter_crawler true
go run ./cmd/cli settings-set enable_telegram_crawler true
go run ./cmd/cli settings-set enable_rewrite_processor true
go run ./cmd/cli settings-set auto_publish false
go run ./cmd/cli settings-set crawl_interval_seconds 300
go run ./cmd/cli settings-set process_interval_seconds 60
go run ./cmd/cli settings-set publish_interval_seconds 30
go run ./cmd/cli settings-set twitter_publish_interval_seconds 900
go run ./cmd/cli settings-set rewrite_provider gemini
go run ./cmd/cli settings-set enable_twitter_publish_vi false
go run ./cmd/cli settings-set enable_twitter_publish_en false
```

Why:

- keeps ingest and rewrite running
- blocks automatic outbound posting until an operator explicitly reviews and publishes
- lowers churn a bit compared with an aggressive demo setup

#### Twitter-Only Intake

Use when you want only Twitter/X as the upstream source of new content.

Shortcut:

```bash
./scripts/settings-preset-twitter-only.sh
```

```bash
go run ./cmd/cli settings-set enable_twitter_crawler true
go run ./cmd/cli settings-set enable_telegram_crawler false
go run ./cmd/cli settings-set enable_rewrite_processor true
go run ./cmd/cli settings-set auto_publish false
go run ./cmd/cli settings-set crawl_interval_seconds 300
go run ./cmd/cli settings-set process_interval_seconds 60
go run ./cmd/cli settings-set publish_interval_seconds 30
go run ./cmd/cli settings-set twitter_publish_interval_seconds 900
go run ./cmd/cli settings-set rewrite_provider gemini
go run ./cmd/cli settings-set enable_twitter_publish_vi false
go run ./cmd/cli settings-set enable_twitter_publish_en false
```

Why:

- ingests only from Twitter/X
- disables Telegram ingest noise from groups or topics
- keeps content in the queue for review without auto-distribution

Important note:

- current outbound Twitter publish still depends on the content pipeline state after Telegram-side publishing, so this preset is truly `Twitter-only intake`, not a fully isolated outbound Twitter-only mode

## Database

Run migrations:

```bash
go run ./cmd/migrate
```

Main tables:

- `sources`: source registry for Twitter and Telegram.
- `content_items`: crawled, rewritten, published, skipped, and failed content.
- `settings`: runtime values, descriptions, and optional JSONB config.
- `logs`: optional persisted operator logs surfaced by `/logs`.

Source metadata:

- `tags` and `topics` are used for routing/publish policy.
- `last_crawled_at`, `last_check_at`, and `last_error` are used for source health and revalidation.
- inactive sources are rechecked after the configured 24-hour gate in the source revalidation job.

## Operator Quick Start

### Local Binary Flow

1. Fill bootstrap config:

```bash
cp config/.env.example config/.env
```

Fill `DATABASE_URL`, AI keys, and Twitter credentials in `config/.env`.

2. Run migrations:

```bash
go run ./cmd/migrate
```

3. Set Telegram runtime:

```bash
go run ./cmd/cli settings-set-json telegram_runtime '{"bot_token":"<telegram-bot-token>","publish_targets":[{"chat_id":"<telegram-chat-id>"}],"ingest_targets":[{"chat_id":"<telegram-chat-id>"}],"admin_user_ids":[]}'
```

4. Enable the worker flags you want:

```bash
go run ./cmd/cli settings-set enable_twitter_crawler true
go run ./cmd/cli settings-set enable_rewrite_processor true
go run ./cmd/cli settings-set auto_publish true
```

5. Start the unified runtime:

```bash
go run ./cmd/content-bot
```

6. Verify:

```bash
curl http://127.0.0.1:8080/healthz
go run ./cmd/cli settings-get telegram_runtime
go run ./cmd/cli settings-get auto_publish
```

### Docker Or Podman Flow

1. Fill `config/.env`.
2. Run migrations once:

```bash
go run ./cmd/migrate
```

3. Set `telegram_runtime` and the worker flags:

```bash
go run ./cmd/cli settings-set-json telegram_runtime '{"bot_token":"<telegram-bot-token>","publish_targets":[{"chat_id":"<telegram-chat-id>"}],"ingest_targets":[{"chat_id":"<telegram-chat-id>"}],"admin_user_ids":[]}'
go run ./cmd/cli settings-set enable_twitter_crawler true
go run ./cmd/cli settings-set enable_rewrite_processor true
go run ./cmd/cli settings-set auto_publish true
```

4. Start containers:

```bash
docker compose up -d --build
podman-compose up -d --build
```

5. Verify:

```bash
curl http://127.0.0.1:8080/healthz
podman-compose ps
podman logs go-content-publisher_worker_1 --tail=100
```

Notes:

- replace `chat_id` with a normal group ID or add `thread_id` for a forum topic
- only `auto_publish` is designed for live no-restart toggling; other settings changes should usually be followed by an API/worker restart
- if you ship a packaged executable, the same flow applies but bootstrap values can live in `config.ini` instead of `config/.env`

## Operator Demo Flow

Use this when you want a fast end-to-end demo without waiting for the long-running worker cadence.

1. Make sure the API and worker runtime are already running. If not, start the unified runtime in another terminal:

```bash
curl http://127.0.0.1:8080/healthz
go run ./cmd/content-bot
```

2. Add one Twitter source if it does not already exist:

```bash
curl -X POST http://127.0.0.1:8080/api/sources \
  -H 'Content-Type: application/json' \
  -d '{"type":"twitter","handle":"@business","name":"Bloomberg Business"}'
```

If the source already exists, skip this step.

3. Crawl once:

```bash
go run ./cmd/cli crawl-twitter-once
go run ./cmd/cli report-twitter-sources --active-only
```

4. Rewrite one pending item:

```bash
go run ./cmd/cli process-next
go run ./cmd/cli report-content-ops --limit=10
```

5. Publish one rewritten item to Telegram:

```bash
go run ./cmd/cli publish-next
go run ./cmd/cli report-content-ops --limit=10
```

6. Confirm the result:

```bash
go run ./cmd/cli settings-get auto_publish
go run ./cmd/cli probe-telegram-targets
curl "http://127.0.0.1:8080/api/content?limit=10"
```

Expected outcome:

- `crawl-twitter-once` inserts one or more `pending` items from active sources
- `process-next` moves one item toward `rewritten` or `skipped`
- `publish-next` moves one rewritten item to `published` and stores Telegram publish metadata
- `report-content-ops` shows the queue moving without having to inspect raw rows manually

Fast retry tips:

- if no new tweets appear, choose a more active source or wait for upstream activity
- if `process-next` skips content, inspect duplicate or editorial skip reasons in `report-content-ops`
- if `publish-next` finds nothing, check `auto_publish`, `telegram_runtime`, and whether the latest item is already `published` or `skipped`

## Failure Playbook

Use this when the quick demo flow does not behave as expected.

### 1. `crawl-twitter-once` does not create new items

Likely causes:

- the source has no fresh tweets since its stored checkpoint
- the source is inactive
- Twitter credentials are invalid or rate-limited

First checks:

```bash
go run ./cmd/cli check-connections
go run ./cmd/cli report-twitter-sources --active-only
go run ./cmd/cli report-twitter-sources --inactive-only
go run ./cmd/cli crawl-twitter-once
curl "http://127.0.0.1:8080/api/content/queue?limit=20"
```

Fast fixes:

- switch to a more active source
- reactivate or correct the source handle
- wait for upstream activity if `since_id` is already at the latest tweet

### 2. `process-next` returns `skipped` or does not produce rewritten content

Likely causes:

- the item was detected as a duplicate
- the provider returned `shouldPublish=false`
- the item became stale under `pending_stale_after_seconds`
- the rewrite provider or API key is failing

First checks:

```bash
go run ./cmd/cli check-connections
go run ./cmd/cli settings-get rewrite_provider
go run ./cmd/cli settings-get pending_stale_after_seconds
go run ./cmd/cli report-content-ops --limit=20
curl "http://127.0.0.1:8080/api/content/queue?limit=20"
```

Fast fixes:

- enqueue a clearly new manual or source-driven item
- inspect the skip reason in `report-content-ops`
- raise stale thresholds only if your demo cadence truly needs it

### 3. `publish-next` does not post to Telegram

Likely causes:

- there is no `rewritten` item ready to publish
- `auto_publish=false` for the worker-driven path
- `telegram_runtime` points to the wrong chat or topic
- the bot is not in the target group/topic or lacks send permission

First checks:

```bash
go run ./cmd/cli settings-get auto_publish
go run ./cmd/cli settings-get telegram_runtime
go run ./cmd/cli probe-telegram-targets
go run ./cmd/cli publish-next
curl "http://127.0.0.1:8080/api/content/recent?limit=20"
```

Fast fixes:

- confirm at least one item is already `rewritten`
- add `thread_id` if the target is a forum topic
- re-add the bot to the group or channel and retry the target probe

### 4. `probe-telegram-targets` fails

Likely causes:

- `telegram_runtime.bot_token` is wrong
- `chat_id` is wrong
- the bot is not a member of the target chat
- the target is a topic but `thread_id` is missing or invalid

First checks:

```bash
go run ./cmd/cli check-connections
go run ./cmd/cli settings-get telegram_runtime
go run ./cmd/cli probe-telegram-targets
```

Fast fixes:

- verify the bot token and chat ID
- add the bot to the target group or channel
- include the correct `thread_id` for forum topics

If the problem is specifically Telegram routing or permissions, continue with the deeper `Telegram Troubleshooting` section below.

## Running Locally

1. Prepare environment:

```bash
cp config/.env.example config/.env
```

Fill the required credentials in `config/.env`, then continue.

2. Run migrations:

```bash
go run ./cmd/migrate
```

3. Check external connectivity:

```bash
go run ./cmd/cli check-connections
```

4. Start API:

```bash
go run ./cmd/api
```

5. Start worker:

```bash
go run ./cmd/worker
```

6. Check health:

```bash
curl http://127.0.0.1:8080/healthz
```

### Build Commands

Build every binary:

```bash
go build ./cmd/api ./cmd/cli ./cmd/migrate ./cmd/worker
```

Build one binary only:

```bash
go build ./cmd/api
go build ./cmd/worker
go build ./cmd/cli
go build ./cmd/migrate
```

### Test Commands

Run all tests:

```bash
go test ./...
```

Run focused tests:

```bash
go test ./internal/content/...
go test ./internal/source/...
go test ./internal/system/...
```

### Detailed Local Run Playbook

Minimal operator sequence:

```bash
go run ./cmd/migrate
go run ./cmd/cli check-connections
go run ./cmd/api
go run ./cmd/worker
```

Manual one-shot sequence without long-running worker:

```bash
go run ./cmd/cli crawl-twitter-once
go run ./cmd/cli ingest-telegram-once
go run ./cmd/cli process-next
go run ./cmd/cli publish-next
go run ./cmd/cli publish-twitter-next
```

This is useful for debugging one stage at a time.

## CLI Commands

Operational commands:

```bash
go run ./cmd/cli check-connections
go run ./cmd/cli crawl-twitter-once
go run ./cmd/cli ingest-telegram-once
go run ./cmd/cli process-next
go run ./cmd/cli publish-next
go run ./cmd/cli publish-twitter-next
go run ./cmd/cli probe-telegram-targets
go run ./cmd/cli revalidate-sources-once
go run ./cmd/cli report-twitter-sources
go run ./cmd/cli report-twitter-sources --active-only
go run ./cmd/cli report-twitter-sources --inactive-only
go run ./cmd/cli report-content-ops --limit=20
```

`report-content-ops` is the compact operator view for current Telegram/manual demos. It groups recent content into:

- `rewritten_ready`: items ready for Telegram publish
- `manual_duplicate_skipped`: manual inputs skipped by exact duplicate protection
- `published_recent`: recently published items

Settings commands:

```bash
go run ./cmd/cli settings-get auto_publish
go run ./cmd/cli settings-set auto_publish true
go run ./cmd/cli settings-set-json telegram_runtime '{"bot_token":"...","publish_targets":[],"ingest_targets":[],"admin_user_ids":[]}'
```

Recommended operator commands:

```bash
go run ./cmd/cli settings-get telegram_runtime
go run ./cmd/cli settings-get auto_publish
go run ./cmd/cli settings-get pending_stale_after_seconds
go run ./cmd/cli settings-get rewritten_stale_after_seconds
go run ./cmd/cli settings-get rewrite_provider
go run ./cmd/cli probe-telegram-targets
```

## System Architecture & API Spec

This section is the developer-facing companion to the operator-first sections above. It documents how the binaries, workers, queues, external APIs, and HTTP endpoints fit together today.

### Architecture Overview

```mermaid
flowchart LR
    Admin[Telegram Admin User]
    Sources[Twitter Sources\nTelegram Chats/Topics]
    API[cmd/api\nGin HTTP API]
    CLI[cmd/cli\nOne-shot Operations]
    Worker[cmd/worker\nScheduler Loops]
    Migrate[cmd/migrate\nSQL Runner]
    DB[(Supabase Postgres)]
    Telegram[Telegram Bot API]
    Gemini[Gemini API]
    DeepSeek[DeepSeek API]
    Twitter[X API]

    Admin -->|commands /add /status /pause| Telegram
    Sources -->|tweets| Twitter
    Sources -->|chat updates| Telegram

    API --> DB
    CLI --> DB
    Worker --> DB
    Migrate --> DB

    Worker -->|getUpdates/sendMessage| Telegram
    CLI -->|check/probe/send| Telegram

    Worker -->|rewrite primary| Gemini
    Worker -->|rewrite fallback| DeepSeek
    CLI -->|process-next| Gemini
    CLI -->|process-next fallback| DeepSeek

    Worker -->|crawl/publish| Twitter
    CLI -->|crawl-twitter-once/publish-twitter-next| Twitter
```

#### Runtime Components

| Component | Responsibility | Reads/Writes |
| --- | --- | --- |
| `cmd/api` | Exposes HTTP endpoints for sources and content. | Reads/writes `sources`, `content_items`. |
| `cmd/worker` | Runs crawl, rewrite, Telegram publish, and Twitter publish loops. | Reads/writes `sources`, `content_items`, `settings`, `logs`. |
| `cmd/cli` | Runs operator one-shot commands for debug, checks, and queue actions. | Reads/writes DB and external APIs depending on command. |
| `cmd/migrate` | Applies SQL migrations from `db/migrations`. | Writes schema and seed data into Supabase Postgres. |
| `Telegram Bot API` | Delivers admin/manual input and accepts outbound publishes. | `getUpdates`, `sendMessage`, `getChat`, `getMe`. |
| `Gemini` / `DeepSeek` | Rewrite providers for structured content transformation. | Read input text, return structured rewrite result. |
| `X API` | Twitter/X crawl and publish surface. | User lookup, timeline fetch, tweet create. |
| `Supabase Postgres` | Source registry, content queue, runtime settings, logs. | Durable system state. |

#### Bounded Context Map

| Package | Purpose | Key Entities |
| --- | --- | --- |
| `internal/source` | Source registry and source health management. | `Source` |
| `internal/content` | Queue lifecycle, rewrite, publish, manual enqueue. | `ContentItem` |
| `internal/system` | External clients, settings, logs, connectivity checks. | `SettingRecord`, `LogRecord` |
| `pkg/app/bootstrap` | Wires config, repositories, services, jobs, HTTP server, worker. | `App` |

### Data Model ERD

```mermaid
erDiagram
    SOURCES ||--o{ CONTENT_ITEMS : "source_id"

    SOURCES {
        uuid id
        string type
        string handle
        string name
        boolean is_active
        string[] tags
        string[] topics
        timestamp last_crawled_at
        timestamp last_check_at
        string last_error
        timestamp created_at
    }

    CONTENT_ITEMS {
        uuid id
        uuid source_id
        string external_id
        string original_text
        string author_name
        string source_url
        timestamp crawled_at
        string status
        string rewritten_text
        string rewritten_text_en
        string tweet_text_vi
        string tweet_text_en
        string fact_check_note
        string fail_reason
        string tweet_vi_id
        string tweet_en_id
        timestamp published_at
        string published_msg_id
    }

    SETTINGS {
        string key
        string value
        jsonb json_value
        string description
    }

    LOGS {
        uuid id
        string level
        string module
        string message
        timestamp created_at
    }
```

#### Data Model Notes

| Table | Role |
| --- | --- |
| `sources` | Registry of Twitter and Telegram origins plus health metadata. |
| `content_items` | Main queue and publish history for crawled and manual content. |
| `settings` | DB-first runtime control plane, including booleans, intervals, and JSON contracts like `telegram_runtime`. |
| `logs` | Optional persisted operational logs surfaced by the admin bot. |

### Processing Flows

#### Flow 1: Twitter Crawl to Telegram and Twitter Publish

```mermaid
flowchart TD
    A[Active Twitter Source] --> B[worker crawl loop or crawl-twitter-once]
    B --> C[Resolve username and tw_since checkpoint]
    C --> D[Fetch new tweets from X API]
    D --> E[Insert pending content_items]
    E --> F[rewrite loop or process-next]
    F --> G{shouldPublish?}
    G -->|no| H[mark skipped]
    G -->|yes| I[mark rewritten]
    I --> J[publish loop or publish-next]
    J --> K[sendMessage to Telegram target]
    K --> L[mark published]
    L --> M[twitter publish loop or publish-twitter-next]
    M --> N[post tweet_vi and/or tweet_en]
```

#### Flow 2: Telegram Admin and Manual Input

```mermaid
flowchart TD
    A[Admin sends /add or plain text] --> B[Telegram Bot API getUpdates]
    B --> C[Telegram ingest loop]
    C --> D{matches ingest_targets?}
    D -->|no| E[ignore]
    D -->|yes| F[Telegram admin handler]
    F --> G{command or plain text}
    G -->|/add or plain text| H[EnqueueManual]
    G -->|status queue pause etc.| I[execute admin action]
    H --> J[content_items.status = pending]
    J --> K[rewrite loop or process-next]
    K --> L{shouldPublish?}
    L -->|no| M[mark skipped]
    L -->|yes| N[mark rewritten]
    N --> O[publish-next or auto publish loop]
    O --> P[Telegram sendMessage]
    P --> Q[mark published]
```

#### Flow 3: Worker Scheduling and Settings

```mermaid
flowchart LR
    Settings[(settings)]
    Crawl[Crawl Loop]
    Process[Rewrite Loop]
    Publish[Telegram Publish Loop]
    TwitterPublish[Twitter Publish Loop]

    Settings --> Crawl
    Settings --> Process
    Settings --> Publish
    Settings --> TwitterPublish

    Crawl -->|source revalidation| DB[(content_items / sources)]
    Crawl -->|telegram ingest| DB
    Crawl -->|twitter crawl| DB
    Process --> DB
    Publish --> DB
    TwitterPublish --> DB
```

### Sequence Diagrams

#### Sequence 1: Telegram Manual Input to Publish

```mermaid
sequenceDiagram
    actor Admin as Telegram Admin
    participant TG as Telegram Bot API
    participant Worker as cmd/worker
    participant AdminHandler as TelegramAdminHandler
    participant Content as ContentService
    participant DB as Supabase Postgres
    participant AI as Gemini/DeepSeek

    Admin->>TG: Send /add <text> or plain text
    Worker->>TG: getUpdates
    TG-->>Worker: update payload
    Worker->>AdminHandler: Handle(message)
    AdminHandler->>Content: EnqueueManual(text, author)
    Content->>DB: INSERT content_items(status=pending)
    Worker->>Content: ProcessNextPending()
    Content->>DB: claim next pending
    Content->>AI: RewriteText(original_text)
    AI-->>Content: structured rewrite result
    Content->>DB: UPDATE status=rewritten or skipped
    Worker->>Content: PublishNextReady()
    Content->>DB: claim next rewritten
    Worker->>TG: sendMessage(rewritten_text)
    TG-->>Worker: message_id
    Content->>DB: UPDATE status=published
```

#### Sequence 2: Twitter Crawl to Telegram and Twitter Publish

```mermaid
sequenceDiagram
    participant Worker as cmd/worker
    participant X as X API
    participant Content as ContentService
    participant DB as Supabase Postgres
    participant AI as Gemini/DeepSeek
    participant TG as Telegram Bot API

    Worker->>DB: read active twitter sources
    Worker->>DB: read tw_uid_* and tw_since_*
    Worker->>X: user lookup / timeline fetch
    X-->>Worker: new tweets
    Worker->>Content: EnqueueFromSource(...)
    Content->>DB: INSERT content_items(status=pending)
    Worker->>Content: ProcessNextPending()
    Content->>AI: RewriteText(original_text)
    AI-->>Content: rewritten payload
    Content->>DB: UPDATE status=rewritten or skipped
    Worker->>TG: sendMessage(rewritten_text)
    TG-->>Worker: message_id
    Content->>DB: UPDATE status=published
    Worker->>X: create tweet_vi / tweet_en
    X-->>Worker: tweet ids
    Content->>DB: UPDATE tweet_vi_id / tweet_en_id
```

### Content State Machine

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> processing: claim next pending
    pending --> skipped: stale pending / duplicate pre-check
    processing --> rewritten: structured rewrite accepted
    processing --> skipped: shouldPublish=false / duplicate post-check
    processing --> failed: provider/runtime error
    rewritten --> publishing: claim next ready for Telegram
    rewritten --> skipped: stale rewritten
    publishing --> published: Telegram send success
    publishing --> failed: Telegram publish error
    failed --> pending: retry failed
    failed --> skipped: admin skip
```

#### Queue State Semantics

| State | Meaning | Typical Writer |
| --- | --- | --- |
| `pending` | Waiting for rewrite. | Twitter crawl, Telegram manual/admin enqueue, API manual enqueue |
| `processing` | Claimed by rewrite worker. | `process-next`, worker rewrite loop |
| `rewritten` | Ready for Telegram publish. | rewrite service |
| `publishing` | Claimed by Telegram publish worker. | `publish-next`, worker publish loop |
| `published` | Sent successfully to Telegram. | Telegram publish job |
| `failed` | Runtime/provider/publish error. | rewrite or publish jobs |
| `skipped` | Intentionally not processed further. | duplicate check, editorial reject, stale policy, admin skip |

### HTTP API Reference

#### Endpoint Summary

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Lightweight health probe for the API process. |
| `GET` | `/api/sources` | List active sources only. |
| `GET` | `/api/sources/report` | List all sources, optionally filtered by `type`. |
| `POST` | `/api/sources` | Create a new source. |
| `PATCH` | `/api/sources/:type/:handle` | Update `tags` and `topics` for an existing source. |
| `DELETE` | `/api/sources/:type/:handle` | Delete a source by type and handle. |
| `POST` | `/api/content` | Enqueue manual content. |
| `POST` | `/api/content/:id/manual-rewrite` | Override content rewrite manually. |
| `GET` | `/api/content/queue` | List queue-visible items. |
| `GET` | `/api/content/recent` | List recent content items. |

#### Common Response Shapes

Health:

```json
{
  "status": "ok"
}
```

List wrapper:

```json
{
  "items": []
}
```

Error:

```json
{
  "error": "message"
}
```

#### `GET /healthz`

Purpose: process-level readiness probe for the HTTP API.

Success response:

```json
{
  "status": "ok"
}
```

Status codes:

| Code | Meaning |
| --- | --- |
| `200` | API process is up and responding. |

Example:

```bash
curl http://127.0.0.1:8080/healthz
```

#### `GET /api/sources`

Purpose: list only active sources.

Query params: none.

Success response item shape:

```json
{
  "id": "src-123",
  "type": "twitter",
  "handle": "@business",
  "name": "Bloomberg",
  "tags": ["macro", "news"],
  "topics": ["rates", "equities"],
  "is_active": true,
  "last_crawled_at": "2026-04-19T11:00:00Z",
  "last_check_at": "2026-04-19T11:00:00Z",
  "last_error": null,
  "created_at": "2026-04-19T10:00:00Z"
}
```

Status codes:

| Code | Meaning |
| --- | --- |
| `200` | Active source list returned. |
| `500` | Failed to list sources. |

Example:

```bash
curl http://127.0.0.1:8080/api/sources
```

#### `GET /api/sources/report`

Purpose: list all sources, including inactive ones.

Query params:

| Name | Type | Required | Meaning |
| --- | --- | --- | --- |
| `type` | string | no | Optional filter, typically `twitter` or `telegram`. |

Status codes:

| Code | Meaning |
| --- | --- |
| `200` | Source report returned. |
| `500` | Failed to build source report. |

Examples:

```bash
curl http://127.0.0.1:8080/api/sources/report
curl "http://127.0.0.1:8080/api/sources/report?type=twitter"
```

#### `POST /api/sources`

Purpose: create a new source record.

Request body:

```json
{
  "type": "twitter",
  "handle": "@business",
  "name": "Bloomberg",
  "tags": ["macro", "news"],
  "topics": ["rates", "equities"]
}
```

Request fields:

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `type` | string | yes | `twitter` or `telegram`. |
| `handle` | string | yes | Username, chat identifier, or source handle. |
| `name` | string | yes | Display name. |
| `tags` | string[] | no | Routing metadata. |
| `topics` | string[] | no | Routing metadata. |

Success response: a single source item in the same shape as `GET /api/sources`.

Status codes:

| Code | Meaning |
| --- | --- |
| `201` | Source created. |
| `400` | Invalid request or invalid source. |
| `409` | Source already exists. |
| `500` | Failed to create source. |

Example:

```bash
curl -X POST http://127.0.0.1:8080/api/sources \
  -H 'Content-Type: application/json' \
  -d '{
    "type":"twitter",
    "handle":"@business",
    "name":"Bloomberg",
    "tags":["macro","news"],
    "topics":["rates","equities"]
  }'
```

#### `PATCH /api/sources/:type/:handle`

Purpose: update routing metadata for an existing source.

Path params:

| Param | Meaning |
| --- | --- |
| `type` | Source type, usually `twitter` or `telegram`. |
| `handle` | Source handle exactly as stored or equivalent normalized form. |

Request body:

```json
{
  "tags": ["macro", "high-priority"],
  "topics": ["rates", "fx"]
}
```

Status codes:

| Code | Meaning |
| --- | --- |
| `204` | Metadata updated. |
| `400` | Invalid request or missing handle. |
| `500` | Failed to update metadata. |

Example:

```bash
curl -X PATCH http://127.0.0.1:8080/api/sources/twitter/@business \
  -H 'Content-Type: application/json' \
  -d '{"tags":["macro","high-priority"],"topics":["rates","fx"]}'
```

#### `DELETE /api/sources/:type/:handle`

Purpose: delete a source by type and handle.

Status codes:

| Code | Meaning |
| --- | --- |
| `204` | Source deleted. |
| `400` | Missing handle. |
| `500` | Failed to delete source. |

Example:

```bash
curl -X DELETE http://127.0.0.1:8080/api/sources/twitter/@business
```

#### `POST /api/content`

Purpose: enqueue manual content directly through HTTP.

Request body:

```json
{
  "text": "Brent crude rose after new supply disruption concerns in the Middle East.",
  "author": "operator"
}
```

Request fields:

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `text` | string | yes | Raw content to enqueue. |
| `author` | string | no | Manual author label. Defaults to `Manual`. |

Success response item shape:

```json
{
  "id": "item-123",
  "source_id": null,
  "external_id": "",
  "original_text": "Brent crude rose after new supply disruption concerns in the Middle East.",
  "author_name": "operator",
  "source_url": null,
  "crawled_at": "2026-04-19T11:00:00Z",
  "status": "pending",
  "rewritten_text": null,
  "rewritten_text_en": null,
  "tweet_text_vi": null,
  "tweet_text_en": null,
  "fact_check_note": null,
  "fail_reason": null,
  "tweet_vi_id": null,
  "tweet_en_id": null,
  "published_at": null,
  "published_msg_id": null
}
```

Status codes:

| Code | Meaning |
| --- | --- |
| `201` | Manual content queued. |
| `400` | Invalid request or enqueue failed. |

Example:

```bash
curl -X POST http://127.0.0.1:8080/api/content \
  -H 'Content-Type: application/json' \
  -d '{
    "text":"Brent crude rose after new supply disruption concerns in the Middle East.",
    "author":"operator"
  }'
```

#### `POST /api/content/:id/manual-rewrite`

Purpose: override a specific item with a manual rewritten body.

Path params:

| Param | Meaning |
| --- | --- |
| `id` | Content item ID. |

Request body:

```json
{
  "rewritten_text": "Giá dầu Brent tăng do lo ngại gián đoạn nguồn cung mới tại Trung Đông."
}
```

Status codes:

| Code | Meaning |
| --- | --- |
| `200` | Manual rewrite applied. |
| `400` | Invalid request or rewrite could not be applied. |

Example:

```bash
curl -X POST http://127.0.0.1:8080/api/content/item-123/manual-rewrite \
  -H 'Content-Type: application/json' \
  -d '{
    "rewritten_text":"Giá dầu Brent tăng do lo ngại gián đoạn nguồn cung mới tại Trung Đông."
  }'
```

#### `GET /api/content/queue`

Purpose: list queue-visible items in these states:

- `pending`
- `processing`
- `rewritten`
- `publishing`
- `failed`

Query params:

| Name | Type | Required | Meaning |
| --- | --- | --- | --- |
| `limit` | integer | no | Maximum items to return. Defaults to `50`. |

Status codes:

| Code | Meaning |
| --- | --- |
| `200` | Queue returned. |
| `500` | Failed to list queue. |

Example:

```bash
curl "http://127.0.0.1:8080/api/content/queue?limit=20"
```

#### `GET /api/content/recent`

Purpose: list recent content items regardless of whether they are manual or source-driven.

Query params:

| Name | Type | Required | Meaning |
| --- | --- | --- | --- |
| `limit` | integer | no | Maximum items to return. Defaults to `20`. |

Status codes:

| Code | Meaning |
| --- | --- |
| `200` | Recent content returned. |
| `500` | Failed to list recent content. |

Example:

```bash
curl "http://127.0.0.1:8080/api/content/recent?limit=20"
```

### Runtime Contracts

#### `telegram_runtime` JSON Shape

```json
{
  "bot_token": "<telegram-bot-token>",
  "publish_targets": [
    {
      "chat_id": "-1003816931426"
    }
  ],
  "ingest_targets": [
    {
      "chat_id": "-1003816931426"
    }
  ],
  "admin_user_ids": []
}
```

Rules:

| Field | Meaning |
| --- | --- |
| `bot_token` | Telegram Bot API credential used by ingest, admin commands, and publish. |
| `publish_targets` | One or more Telegram destinations for rewritten content. |
| `ingest_targets` | Chats/topics from which admin commands and plain-text manual input are accepted. |
| `admin_user_ids` | Optional allowlist; empty means allow every sender in an ingest target. |

#### Worker Loop Settings

| Setting | Used By | Effect |
| --- | --- | --- |
| `crawl_interval_seconds` | Worker crawl loop | Controls source revalidation, Telegram ingest, and Twitter crawl cadence. |
| `process_interval_seconds` | Worker rewrite loop | Controls how often pending items are rewritten. |
| `publish_interval_seconds` | Worker Telegram publish loop | Controls how often rewritten items are published to Telegram. |
| `twitter_publish_interval_seconds` | Worker Twitter publish loop | Controls Twitter post cadence after Telegram publish. |
| `auto_publish` | Telegram publish loop | Enables or disables automatic Telegram publish only. |

#### Source Runtime Fields

| Field | Meaning |
| --- | --- |
| `tags` | Free-form routing labels for publish policy and source grouping. |
| `topics` | Higher-priority routing labels used by Twitter publish policy. |
| `last_crawled_at` | Last successful crawl/ingest touch for the source. |
| `last_check_at` | Last validation/revalidation attempt. |
| `last_error` | Last validation or crawl error used for operator diagnosis. |

## Source Management

Add sources through the API or direct DB operations. A source needs:

- `type`: `twitter` or `telegram`
- `handle`: username or chat identifier
- `name`: display name

Twitter behavior:

- First crawl initializes `tw_uid_<handle>` and `tw_since_<handle>`.
- First crawl does not backfill old tweets.
- Later crawls enqueue only tweets newer than the saved checkpoint.

Telegram Bot API behavior:

- Current ingest uses Bot API updates.
- It only sees chats/topics where the bot receives updates.
- Admin commands share the same Bot API update offset as content ingest.

Telegram source guidance:

- For this project, Telegram operations are bot-only.
- The bot can ingest only from chats/topics where it is present and receives updates.
- The bot can publish only to chats/topics listed in `publish_targets`.
- Public Telegram usernames/channels that require user-session crawling are out of scope.

## Demo Flow

For a controlled demo that crawls, rewrites, and posts to Telegram:

1. Confirm settings:

```bash
go run ./cmd/cli settings-get enable_twitter_crawler
go run ./cmd/cli settings-get enable_rewrite_processor
go run ./cmd/cli settings-get auto_publish
go run ./cmd/cli settings-get telegram_runtime
```

2. Enable Telegram auto-publish temporarily:

```bash
go run ./cmd/cli settings-set auto_publish true
```

3. Crawl:

```bash
go run ./cmd/cli crawl-twitter-once
```

4. Rewrite:

```bash
go run ./cmd/cli process-next
```

5. Publish to Telegram:

```bash
go run ./cmd/cli publish-next
```

6. Optional Twitter publish after Telegram publish:

```bash
go run ./cmd/cli publish-twitter-next
```

7. Return to safe develop mode:

```bash
go run ./cmd/cli settings-set auto_publish false
```

The Telegram publish loop reads `auto_publish` at runtime, so changing this key does not require a restart.

## Telegram Troubleshooting

Common Telegram issues:

- `chat not found`:
  - bot is not in the target group/channel
  - `chat_id` is wrong
  - target is a forum topic but `thread_id` is missing
  - bot cannot see or send to that chat yet
- no ingest updates:
  - bot is not a member of the target chat
  - group privacy/settings prevent the bot from receiving updates
  - wrong `ingest_targets`
- publish works to old group but fails to new group:
  - `telegram_runtime.publish_targets` was updated but the bot was not added to the new group
  - the new group has different topic/forum behavior
- content crawled but nothing gets rewritten:
  - item exceeded `pending_stale_after_seconds` and was marked `skipped`
- content rewritten but nothing gets posted:
  - `auto_publish=false`
  - item exceeded `rewritten_stale_after_seconds`
  - Telegram publish failed and item moved to `failed`

Fast checks:

```bash
go run ./cmd/cli check-connections
go run ./cmd/cli settings-get telegram_runtime
go run ./cmd/cli settings-get auto_publish
go run ./cmd/cli settings-get pending_stale_after_seconds
go run ./cmd/cli settings-get rewritten_stale_after_seconds
go run ./cmd/cli probe-telegram-targets
go run ./cmd/cli publish-next
```

## Docker

Build:

```bash
docker build -t content-bot-go:latest .
```

Run with Compose:

```bash
docker compose up -d --build
podman-compose up -d --build
```

The local environment used during development may not always have Docker installed; verify Docker separately on the deployment host.

For Podman compatibility, the compose file no longer relies on a top-level Compose `name:` expression. The default network is pinned as `content-bot-go_default`, which avoids the `__-content-bot-go_default` project-name parsing issue seen in some `podman-compose` releases.

Useful Podman checks after startup:

```bash
podman-compose ps
podman logs go-content-publisher_api_1 --tail=100
podman logs go-content-publisher_worker_1 --tail=100
curl -s http://127.0.0.1:8080/healthz
podman-compose down
```

## Single Binary

The repository now also supports a unified runtime binary in addition to the existing `api`, `worker`, `migrate`, and `cli` binaries.

Build a Windows executable:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o content-bot.exe ./cmd/content-bot
```

Run locally with the unified binary:

```bash
go run ./cmd/content-bot
```

Supported commands:

```bash
go run ./cmd/content-bot
go run ./cmd/content-bot run
go run ./cmd/content-bot api
go run ./cmd/content-bot worker
go run ./cmd/content-bot migrate
```

Default behavior:

- `run` is the default command
- it can start API and worker together in one process
- the old binaries still work exactly as before

## INI Config

The runtime now checks these config files before `.env` files:

```bash
config/config.ini
config.ini
config/.env
.env
```

`config.ini` is bootstrap-friendly and does not replace the DB-first runtime settings model. A good operator pattern is:

- keep only bootstrap fields in `config.ini`, such as database connectivity, AI API keys, Twitter credentials, and unified process flags like `auto_migrate`, `run_api`, and `run_worker`
- keep live runtime behavior in the `settings` table
- do not duplicate DB-managed keys such as `telegram_runtime`, `auto_publish`, `rewrite_provider`, worker feature flags, or scheduler intervals in `config.ini`

Sample file:

```bash
cp config/config.example.ini config/config.ini
```

Important precedence:

- existing shell env vars win over `config.ini`
- `.env` values loaded later still override `config.ini`
- DB `settings` still override the relevant runtime fields after bootstrap, as before

`config/config.example.ini` intentionally omits DB-managed runtime keys so operators do not assume file edits will override `settings`.

Quick operator rule:

- `config.ini` is for bootstrap and secrets
- `settings` is for bot behavior and policy
- when in doubt, do not duplicate the same concern in both places

For a packaged operator handoff, see the `release/` folder.

## Verification

Standard checks:

```bash
go test ./...
go build ./cmd/api ./cmd/cli ./cmd/migrate ./cmd/worker
go run ./cmd/cli check-connections
```

Live checks require valid Supabase, Telegram, Gemini/DeepSeek, and Twitter credentials.

## Parity Roadmap

The next implementation waves are:

1. Bot-only operator polish and troubleshooting.
2. Queue and duplicate-policy tuning for safer demos.
3. README/progress/changelog sync after each wave.

See `docs/superpowers/plans/2026-04-19-content-bot-parity-completion.md` for the execution plan and agent routing.
