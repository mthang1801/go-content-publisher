# Progress

## Current State
- Phase 1 focus is `.agents` as the reusable source-of-truth operating model for future Go services.
- Entry docs, Go-native skills, workflows, rules, contexts, prompts, hooks, and helper scripts are scaffolded.
- Team-level agent docs have been hardened from short profiles into detailed operating contracts with routing, handoff, checklist, and completion rules.
- `architecture.md` now needs to be treated as a structural spec, and `go-orchestrator.md` as the execution governor for the team.
- Specialist cards now follow the same deeper format across review, debugging, database, devops, test, and technical-writing roles.
- `skills/` now contains a copied local base skill set, while `.agents/skills/` contains local bridge skills and precise sample-specific skills.
- Agent guidance now explicitly forbids falling back to any external root skill tree, so the sample can be moved to a different repository without base-skill reference drift.
- Copied skill docs and `.agents` docs now use repo-local root-relative paths such as `skills/`, `.agents/`, `docs/`, and `changelogs/` instead of home-directory or nested-repo-specific references.
- Each agent role now has a dedicated bundle skill under `.agents/skills/`, and role cards reference that single entry skill instead of expanding the entire stack inline.
- A migration design spec for converting `public/content-bot-share` into a Go/Gin/GORM/Supabase implementation has been written under `docs/superpowers/specs/`.
- A tracked change record and execution plan now exist for the first content-bot implementation wave, including `config/.env`, SQL migrations, Dockerfile, and Docker Compose deliverables.
- The first Go runtime foundation now exists under `cmd/`, `internal/`, `pkg/`, and `db/migrations/`, including `api`, `worker`, `migrate`, config loading, Supabase Postgres wiring, initial bounded contexts, and container artifacts.
- Basic verification has passed for Go formatting, unit tests, binary builds, live Supabase migration execution, and outbound connectivity checks; Docker verification remains blocked on local tooling.
- A real connectivity check path now exists through `cmd/cli`, with live verification completed for Supabase Postgres, Telegram Bot API, and DeepSeek API using the current `config/.env`.
- A live Supabase read/write smoke test has passed through the HTTP API by creating, listing, and deleting a temporary `source` record.
- HTTP presentation now maps domain models into explicit API DTOs, so responses no longer leak exported Go field names directly.
- The content workflow now supports manual enqueue, one-shot processing/publishing, and manual rewrite override, making operator-driven recovery possible when an external AI dependency is degraded or out of credit.
- Telegram publish has been verified end-to-end against the configured supergroup, while DeepSeek rewrite currently reaches the vendor API but is blocked by a `402 Payment Required` response from the upstream account.
- A first real Telegram ingest slice now exists: worker/CLI polling uses Bot API `getUpdates`, persists offset state in `settings`, matches active Telegram sources by username or chat ID, and enqueues matching messages into Supabase-backed `content_items`.
- Live validation shows the ingest path is wired correctly to Supabase and Telegram, but the configured `Coding` supergroup currently yields `updates=0`, so Telegram-side delivery/privacy remains the blocker for automatic ingestion.
- Telegram runtime config is now topic-aware and list-based through `.env`, with independent `TELEGRAM_PUBLISH_TARGETS` and `TELEGRAM_INGEST_TARGETS` JSON arrays.
- Topic-aware publish has been verified live against `chat_id=-1002451344189` and `thread_id=5`; the resulting `published_msg_id` is now stored as structured JSON when publishing to one or more Telegram targets.
- Telegram ingest has now received live updates from the bot API, but the captured update was a `my_chat_member` admin event rather than a topic message, so automatic content ingest is still waiting on real message updates for the configured topic.
- A second rewrite provider now exists alongside DeepSeek: Gemini can be selected through `REWRITE_PROVIDER=gemini`, with dedicated `GEMINI_API_KEY` and `GEMINI_MODEL` configuration.
- Live verification confirmed both Gemini connectivity and end-to-end rewrite execution; `process-next` successfully rewrote a pending content item through Gemini and persisted the rewritten text in Supabase.
- Gemini rewrite has now also been verified through the publish leg into Telegram topic `5`, so the end-to-end path `gemini -> publish-next -> Telegram topic` is confirmed.
- Rewrite execution now supports automatic fallback to the secondary provider when the primary provider errors, though a successful live fallback to DeepSeek is still blocked by the upstream billing condition on that account.
- The `process-next` path now claims pending items atomically in Postgres with row locking, which removes the concrete duplicate-processing race that appeared when multiple rewrite workers were started in parallel.
- The `publish-next` path now also claims rewritten items atomically before Telegram publish, so concurrent publisher workers no longer race on the same row.
- A first Twitter parity slice now exists in Go: worker bootstrap wires a real Twitter crawl action and a real Twitter publish action instead of the previous noop adapter.
- Twitter crawl now uses the configured bearer token to resolve usernames, cache user IDs and `since_id` checkpoints in `settings`, and enqueue new tweets from active `twitter` sources into `content_items`.
- Twitter publish now uses the configured account credentials to post Vietnamese and English tweet variants for already-published content items, persisting `tweet_vi_id` and `tweet_en_id` back into Supabase.
- Live verification now confirms Twitter crawl and Twitter publish are both usable with the current credentials; a real tweet was posted and `tweet_vi_id` persisted back into Supabase.
- Source management now tracks `last_crawled_at`, `last_check_at`, and `last_error`, allowing invalid Telegram/Twitter sources to be marked inactive and later revalidated.
- A new `source_revalidation` worker subaction and `cmd/cli revalidate-sources-once` command now validate newly added active sources once, mark unusable sources inactive, and re-activate inactive sources only after they have been due for recheck for at least 24 hours.
- Live verification confirmed `revalidate-sources-once` checked 29 stored sources, inactivated 9 unusable ones, and a second immediate run checked 0 sources, proving the 24-hour recheck gate is active.
- Twitter publish policy is now runtime-configurable from `settings`, including `twitter_publish_after`, `twitter_publish_source_types`, `twitter_publish_source_tags`, `twitter_publish_source_topics`, and `twitter_publish_topic_keywords`.
- Sources now carry structured `tags` and `topics`, exposed through the source API/report surface and updateable via `PATCH /api/sources/:type/:handle`, so auto-post routing no longer has to rely only on text keyword matching.
- Queue stale policy now supports split runtime settings `pending_stale_after_seconds` and `rewritten_stale_after_seconds`, with backward-compatible fallback to the legacy `queue_stale_after_seconds` key.
- The `settings` table now includes a `description` column, and `settings-get` returns both the current value and the human-readable meaning of each runtime key.
- The legacy `queue_stale_after_seconds` runtime path has now been removed; only `pending_stale_after_seconds` and `rewritten_stale_after_seconds` remain active for queue aging policy.
- Worker feature flags `enable_twitter_crawler` and `enable_rewrite_processor` now come from `settings`, and the duplicate env keys have been removed from `.env` so the runtime stays DB-first for these controls.
- Remaining worker feature flags `enable_telegram_crawler`, `enable_twitter_publish_vi`, and `enable_twitter_publish_en` now also come from `settings`, and the duplicate env keys have been removed.
- Telegram runtime configuration now lives in `settings.json_value` under `telegram_runtime`, including `bot_token`, publish targets, ingest targets, and admin IDs.
- The stale template README has been replaced with content-bot-specific setup, operations, settings, CLI, API, demo, Docker, and parity roadmap documentation.
- README now also includes Mermaid-based system architecture, ERD, processing flows, sequence diagrams, content state machine, and a code-grounded HTTP API specification while keeping the document operator-first.
- A focused parity completion plan now exists at `docs/superpowers/plans/2026-04-19-content-bot-parity-completion.md`, covering structured rewrite, Telegram admin bot, bot-only Telegram operations, and documentation sync.
- Structured rewrite parity is now implemented at the Go contract level: Gemini and DeepSeek request JSON, parse `rewrittenText`, `rewrittenTextEn`, `tweetVI`, `tweetEN`, `factCheckNote`, `shouldPublish`, and `reason`, and the content service persists those fields.
- The rewrite processor now skips expected editorial rejects with `shouldPublish=false` and supports pre-rewrite/post-rewrite similarity duplicate detection with settings-driven thresholds.
- Telegram admin bot parity is now implemented in the existing Bot API ingest loop, including `status`, `queue`, `recent`, `sources`, `addsource`, `removesource`, `retry`, `skip`, `pause`, `resume`, `logs`, and `crawlnow`.
- Telegram admin input parity now also covers `/add <text>` and plain non-command text from authorized admins in `ingest_targets`, which are queued as manual pending content like the TypeScript sample.
- Admin command handling uses `telegram_runtime.admin_user_ids` for allowlisting, keeps a single `getUpdates` offset consumer, and deactivates sources instead of hard deleting them.
- The Telegram publish loop now reads `settings.auto_publish` at runtime, so `/pause` and `/resume` can take effect without restarting the worker.
- Scheduler intervals now come from DB-backed settings keys `crawl_interval_seconds`, `process_interval_seconds`, `publish_interval_seconds`, and `twitter_publish_interval_seconds`; the duplicate env entries were removed from `config/.env`.
- Telegram scope is now intentionally Bot API only for operations and planning; MTProto/App ID/session are no longer part of the recommended runtime contract for this project.
- The dormant Telegram MTProto/public-crawl code path has now been removed from the active runtime, and the obsolete merge/public settings are being retired so the codebase matches the bot-only operating model.
- Manual/admin queue items now use exact-only duplicate skipping during rewrite, which makes demo prompts less likely to be skipped while preserving exact publish dedupe against recent Telegram posts.
- Telegram admin/manual ingest now sanitizes invalid UTF-8 before replying through Bot API, removing the runtime failure that surfaced during live `/add`-style group verification.
- A compact CLI operator report now exists as `report-content-ops`, summarizing `rewritten_ready`, manual duplicate skips, and recently published items for faster demo monitoring.
- An optional `config.ini` bootstrap path and unified `cmd/content-bot` binary now exist alongside the original binaries, preserving the current workflow while enabling single-executable packaging for operators.
- The sample `config.ini` files have been trimmed to bootstrap-only fields so operators configure DB-managed runtime behavior exclusively through the `settings` table instead of split file/database contracts.
- Operator docs now include an explicit config-versus-settings matrix and restart guidance, reducing ambiguity around which values are bootstrap-only and which are runtime policy.
- README and release runbook now include a compact operator quick-start flow that sequences bootstrap config, migrate, `telegram_runtime`, feature flags, and local/Docker startup in the actual order operators use.
- README and release runbook now also include a compact operator demo flow for the common one-shot pipeline `crawl-twitter-once -> process-next -> publish-next -> report-content-ops`.
- README and release runbook now also include a compact failure playbook for the most common demo-stage failures across crawl, rewrite, publish, and Telegram target probing.
- README now includes a short settings cookbook for `demo`, `safe-production`, and `twitter-only intake`, while the release runbook points operators to those presets without overloading the packaged handoff note.
- Executable helper scripts now exist under `scripts/` for the three settings cookbook presets, giving operators a faster path than replaying each `settings-set` command by hand.
- `docker-compose.yml` has been adjusted for better `podman-compose` compatibility by removing the top-level project-name expression and pinning the default network name explicitly.
- A lightweight `release/` layout now exists for packaging the unified binary with a sample `config.ini` and a short operator runbook.

## Next Milestones
1. Decide whether crawler/rewrite/Twitter publish feature flags also need runtime live reload, or whether bootstrap-time settings are acceptable for those controls.
2. Keep README, progress, and changelog in sync after each parity wave.
3. Introduce Redis behind explicit ports after content-bot parity stabilizes.
4. Design broker adapter-manager patterns for Kafka and RabbitMQ.

## Risks
- The current boundary hook is intentionally lightweight and should evolve into stronger import/package verification once the codebase exists.
- Future messaging support needs a careful abstraction to avoid premature generic infrastructure.
- The agent roster is backend-first; future product-analysis or frontend-specific roles should only be added if the sample grows beyond service-template scope.
- The content-bot migration now introduces deployment artifacts and a Supabase-backed runtime path, so config discipline and migration verification need to stay ahead of feature work.
- Docker-based verification is not currently runnable in this environment because the `docker` CLI is unavailable.
- Telegram ingest for supergroups depends on Telegram actually delivering updates to the bot; privacy mode, admin rights, and source type semantics can still block visibility even when outbound `sendMessage` works.
- Because publish can now target multiple Telegram destinations, `published_msg_id` is no longer guaranteed to be a single scalar Telegram message id; downstream consumers should treat it as an opaque string or parse JSON if they need per-target delivery detail.
- Telegram username existence is highly uneven across the imported advisory list; several Telegram handles resolved as nonexistent or suspicious channels and have therefore been pushed inactive until a future recheck succeeds.
