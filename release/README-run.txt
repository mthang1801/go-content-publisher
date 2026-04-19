Content Bot Release Layout

Files:
- content-bot.exe
- config.ini

Typical operator flow:
1. Copy `config.example.ini` to `config.ini`
2. Fill the bootstrap values in `config.ini`
3. Run `content-bot.exe`

Default behavior:
- `content-bot.exe` is the same as `content-bot.exe run`
- it can start API and worker together
- if `auto_migrate=true`, migrations run before startup

Useful commands:
- `content-bot.exe`
- `content-bot.exe run`
- `content-bot.exe api`
- `content-bot.exe worker`
- `content-bot.exe migrate`

Quick start:
1. Fill bootstrap values in `config.ini`
2. Run `content-bot.exe migrate`
3. Set `telegram_runtime` and worker flags through the CLI or an existing operator environment
4. Run `content-bot.exe`
5. Check `/healthz` and the worker logs

Minimal runtime settings example:
- `telegram_runtime={"bot_token":"<telegram-bot-token>","publish_targets":[{"chat_id":"<telegram-chat-id>"}],"ingest_targets":[{"chat_id":"<telegram-chat-id>"}],"admin_user_ids":[]}`
- `enable_twitter_crawler=true`
- `enable_rewrite_processor=true`
- `auto_publish=true`

Settings cookbook:
- `demo`: crawl + rewrite + Telegram auto publish on, Twitter outbound publish off
- `safe-production`: crawl + rewrite on, all auto outbound publish off for human review
- `twitter-only intake`: only Twitter/X is used as the upstream source; Telegram ingest off, auto publish off
- if you are running from a source checkout, helper scripts live under `scripts/`
- full preset commands live in the main `README.md`

Fast demo flow:
1. Make sure one active Twitter source exists in the database
2. Run one crawl
3. Run one rewrite
4. Run one Telegram publish
5. Check the operator report and Telegram target

Equivalent one-shot commands in a source checkout:
- `go run ./cmd/cli crawl-twitter-once`
- `go run ./cmd/cli process-next`
- `go run ./cmd/cli publish-next`
- `go run ./cmd/cli report-content-ops --limit=10`

Failure playbook:
- crawl creates nothing: check source activity, source status, and Twitter connectivity
- process skips: inspect duplicate/editorial skip reason and stale thresholds
- publish sends nothing: check `auto_publish`, `telegram_runtime`, rewritten queue state, and Telegram target probe
- Telegram target probe fails: verify bot token, `chat_id`, optional `thread_id`, and bot membership in the target chat

Notes:
- `config.ini` is only for bootstrap secrets and local process flags
- DB-backed `settings` still control live bot behavior after startup
- runtime keys such as `telegram_runtime`, `auto_publish`, `rewrite_provider`, feature flags, and loop intervals belong in `settings`, not in `config.ini`
- only `auto_publish` is designed for live no-restart toggling; most other `settings` changes should be followed by an API/worker restart
- keep secrets private; do not share a filled `config.ini` publicly
