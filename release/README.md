# Content Bot Release

## Files

- `content-bot.exe`
- `cli.exe`
- `config.ini`
- `run-content-bot.bat`
- `run-cli.bat`
- `db/migrations/`

## Tool roles

- `content-bot.exe` is the main runtime process. Use it to run the API, worker, or both.
- `cli.exe` is an operator tool. Use it for checks, reports, one-shot jobs, and settings updates.
- `cli.exe` does not need to stay running. If you never open it, the main app can still run normally once config, migrations, and DB settings are in place.
- `run-content-bot.bat` starts `content-bot.exe run` from the package folder and keeps the console window open after the app stops.
- `run-cli.bat` runs `cli.exe` commands from the package folder and keeps the console window open after the command ends.

## Typical operator flow

1. Fill the Supabase/bootstrap values in `config.ini`.
2. Run `cli.exe check-connections`.
3. Run `content-bot.exe migrate`.
4. Run `content-bot.exe show-runtime`.
5. Run `content-bot.exe`.

## Default behavior

- `content-bot.exe` is the same as `content-bot.exe run`
- it can start API and worker together
- if `auto_migrate=true`, migrations run before startup

## Useful commands

- `content-bot.exe help`
- `content-bot.exe`
- `content-bot.exe run`
- `content-bot.exe api`
- `content-bot.exe worker`
- `content-bot.exe migrate`
- `content-bot.exe version`
- `content-bot.exe show-runtime`
- `content-bot.exe check-connections`
- `cli.exe help`
- `cli.exe check-connections`
- `cli.exe version`
- `cli.exe show-runtime`
- `cli.exe settings-get <key>`
- `cli.exe settings-set <key> <value>`

## Quick start

1. Fill bootstrap values in `config.ini`.
2. Point `database.url` at your Supabase Postgres connection string.
3. If the Windows host or network is IPv4-only, use the Supabase Session pooler URL from the Supabase Connect panel; direct DB URLs are IPv6-first.
4. Run `cli.exe help` if you want the operator command list inside the package.
5. Run `cli.exe version` or `content-bot.exe version` to verify the release build metadata.
6. Run `cli.exe check-connections`.
7. Run `content-bot.exe migrate`.
8. Run `content-bot.exe show-runtime`.
9. Set `telegram_runtime` and worker flags through `cli.exe` or an existing operator environment.
10. Run `run-content-bot.bat` or `content-bot.exe run`.
11. Check `/healthz` and the worker logs.

When you build a Windows package from the source repository, `release/build-windows.sh` now prefers `config/config.ini` if that file exists. If it does not exist, the script falls back to `release/config.example.ini`.

`content-bot.exe show-runtime` prints the effective runtime shape without touching DB-backed `settings`, including:

- app env and run flags
- HTTP host, configured port, and whether the port mode is `fixed` or `auto`
- database target
- inferred database mode such as `local`, `supabase_direct`, `supabase_session_pooler`, or `supabase_transaction_pooler`
- config files that were actually loaded

When `http.port=auto`, `http.port=0`, or `http.port=` the API log prints the actual bound port and a local health URL, for example:

```text
time=... level=INFO msg="api listening" addr=[::]:40151 port=40151 healthz=http://127.0.0.1:40151/healthz
```

By default the app writes logs to both the console and `logs/app.log` in the package working directory.
When `logs/app.log` grows past the built-in size limit, it rotates to `logs/app.log.1` and starts a fresh `logs/app.log`.
Set `log_max_size_mb` in `config.ini` if you want to change that limit.

## Minimal runtime settings example

- `telegram_runtime={"bot_token":"<telegram-bot-token>","publish_targets":[{"chat_id":"<telegram-chat-id>"}],"ingest_targets":[{"chat_id":"<telegram-chat-id>"}],"admin_user_ids":[]}`
- `enable_twitter_crawler=true`
- `enable_rewrite_processor=true`
- `auto_publish=true`

## Settings cookbook

- `demo`: crawl + rewrite + Telegram auto publish on, Twitter outbound publish off
- `safe-production`: crawl + rewrite on, all auto outbound publish off for human review
- `twitter-only intake`: only Twitter/X is used as the upstream source; Telegram ingest off, auto publish off
- if you are running from a source checkout, helper scripts live under `scripts/`
- full preset commands live in the main repository `README.md`

## Fast demo flow

1. Make sure one active Twitter source exists in the database.
2. Run one crawl.
3. Run one rewrite.
4. Run one Telegram publish.
5. Check the operator report and Telegram target.

## Equivalent one-shot commands in a source checkout

- `go run ./cmd/cli crawl-twitter-once`
- `go run ./cmd/cli process-next`
- `go run ./cmd/cli publish-next`
- `go run ./cmd/cli report-content-ops --limit=10`

## Failure playbook

- crawl creates nothing: check source activity, source status, and Twitter connectivity
- process skips: inspect duplicate/editorial skip reason and stale thresholds
- publish sends nothing: check `auto_publish`, `telegram_runtime`, rewritten queue state, and Telegram target probe
- Telegram target probe fails: verify bot token, `chat_id`, optional `thread_id`, and bot membership in the target chat

## Notes

- `config.ini` is only for bootstrap secrets and local process flags
- DB-backed `settings` still control live bot behavior after startup
- runtime keys such as `telegram_runtime`, `auto_publish`, `rewrite_provider`, feature flags, and loop intervals belong in `settings`, not in `config.ini`
- only `auto_publish` is designed for live no-restart toggling; most other `settings` changes should be followed by an API/worker restart
- set `http.port=auto`, `http.port=0`, or leave it blank if you want each package instance to bind its own free API port
- logs are written both to the console and to `logs/app.log`
- log rotation is intentionally simple: one active file plus one backup file `logs/app.log.1`
- configure log rotation size with `log_max_size_mb` in the `[app]` section
- `content-bot.exe migrate` reads migrations from `db/migrations` next to the executable package, so keep that folder with the release files
- keep secrets private; do not share a filled `config.ini` publicly
