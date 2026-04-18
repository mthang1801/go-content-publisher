# Content Bot Go Migration Design

**Date:** 2026-04-18

**Status:** Proposed and approved at design level, pending spec review before implementation planning.

## Goal

Convert the sample project under `public/content-bot-share` from TypeScript + Prisma + SQLite into a Go-native foundation in the main repository using Gin, GORM, and Supabase Postgres.

The first delivery wave focuses on infrastructure and migration readiness rather than full integration parity with Telegram, Twitter, and DeepSeek.

## Scope

### In Scope

- Define the Go runtime topology for the content-bot sample.
- Define bounded contexts and package placement for the first implementation wave.
- Replace Prisma + SQLite persistence assumptions with GORM + SQL migrations targeting Supabase Postgres.
- Create a configuration contract covering app/server, Supabase, scheduler, and external integrations.
- Define container build and local deployment artifacts for the Go application.
- Define the first migration plan and verification strategy for Go.
- Create implementation planning artifacts for the Go conversion.

### Out Of Scope For Wave 1

- Full Telegram bot command implementation.
- Real Telegram realtime crawler implementation.
- Real Twitter crawler and publisher implementation.
- Real DeepSeek rewrite implementation.
- Content deduplication heuristics, aggregation windows, and fact-checking logic.
- Production deployment automation.

## Recommended Approach

Use an infrastructure-first migration.

This repository currently contains the Go operating model and documentation but does not yet contain runnable service code under `cmd/`, `internal/`, or `pkg/`. Because of that, the first conversion wave should establish the Go runtime shape, persistence, migrations, config contract, and worker skeleton before attempting full feature parity with the TypeScript sample.

This approach keeps the blast radius controlled, allows early verification against Supabase Postgres, and avoids coupling HTTP contracts to assumptions that should instead be owned by the worker pipeline.

## Runtime Topology

### `cmd/api`

Owns:

- Gin HTTP server startup
- health and readiness endpoints
- admin/read endpoints for sources, content queue, and status views
- wiring of application services into HTTP handlers

Does not own:

- scheduler loops
- long-running background processing
- direct GORM model exposure

### `cmd/worker`

Owns:

- scheduler loop skeleton
- orchestration hooks for `crawl -> process -> publish`
- startup of background jobs and graceful shutdown
- adapter availability reporting for integrations that are not implemented yet

In wave 1, external integrations stay behind ports/interfaces and may use noop or stub implementations.

### `cmd/migrate`

Owns:

- applying SQL migrations to Supabase Postgres
- controlled schema evolution
- optional controlled data fixes when needed

Migrations are the schema source of truth. GORM auto-migrate is not part of the production workflow.

## Container And Deployment Shape

### `Dockerfile`

Wave 1 should include a production-oriented multi-stage `Dockerfile` that:

- builds the Go binaries in a builder stage
- produces a slim runtime image
- supports at least the API process as the default container target
- can be extended to run the worker and migrate binaries with command overrides

The image should avoid bundling development-only tooling and should load configuration entirely from environment variables.

### `docker-compose.yml`

Wave 1 should include a `docker-compose.yml` that supports deployment and operator workflows for the initial Go runtime.

Expected services:

- `api`
- `worker`
- `migrate` as an on-demand or profile-gated service

Because Supabase is the system of record database, Compose should not provision a local Postgres by default for the main deployment path. Instead, the services should connect to Supabase Postgres through environment configuration.

Compose should:

- share the same built image where practical
- allow command overrides for `api`, `worker`, and `migrate`
- load values from `config/.env`
- express service dependencies only where they are real
- remain compatible with simple VPS deployment via Docker Compose

## Bounded Contexts

### `internal/source`

Owns:

- source domain model
- source repository ports
- source administration use cases
- source HTTP handlers

Primary entity from the sample:

- `Source`

### `internal/content`

Owns:

- content item domain model
- pipeline status transitions
- repository ports and persistence mapping
- queue and recent-item query use cases
- worker-facing orchestration primitives

Primary entity from the sample:

- `ContentItem`

### `internal/system`

Owns:

- settings access
- logs/read-side operational support when needed
- health/reporting utilities that are local to this sample

This context should remain small. Shared reusable runtime concerns still belong in `pkg/`.

## Persistence Strategy

### Database Role

Supabase is used as the storage backend through its PostgreSQL database. The Go application connects to Supabase Postgres using a standard DSN.

Supabase is not treated as an MCP transport for the runtime application path.

### Schema Source

Schema evolution is defined through SQL migration files and executed by `cmd/migrate`.

GORM is used for:

- persistence model mapping
- repository implementations
- transactional access patterns

GORM is not used as the authoritative schema migration engine.

### Entity Mapping From Prisma

The TypeScript sample currently defines:

- `Source`
- `ContentItem`
- `Setting`
- `Log`

These map to PostgreSQL tables:

- `sources`
- `content_items`
- `settings`
- `logs`

### Identifier Strategy

The Prisma sample uses `cuid()`. The Go version should standardize on PostgreSQL UUIDs for consistency with Supabase and Go tooling.

Decision:

- use UUID primary keys
- keep external content IDs as explicit business keys
- preserve uniqueness through database constraints

### Required Constraints And Indexes

From the sample schema and likely query paths:

- unique `(type, handle)` on `sources`
- unique `(source_id, external_id)` on `content_items`
- index on `content_items.status`
- index on `content_items.crawled_at`
- index on `logs.created_at`

Additional index work may be added after actual list/read endpoints are finalized.

## Initial Migration Plan

### `0001_init_content_bot_schema.sql`

Creates:

- extensions required for UUID generation if needed
- `sources`
- `content_items`
- `settings`
- `logs`

Defines core columns, nullability, defaults, and foreign keys.

### `0002_indexes_and_uniqueness.sql`

Creates:

- unique constraints
- secondary indexes aligned to queue/status access patterns

### `0003_seed_settings.sql`

Optional seed migration for default setting rows only if the application requires them at bootstrap time.

## Configuration Contract

Create `config/.env` with placeholders for:

### App And Server

- `APP_NAME`
- `APP_ENV`
- `APP_LOG_LEVEL`
- `HTTP_HOST`
- `HTTP_PORT`
- `HTTP_READ_TIMEOUT`
- `HTTP_WRITE_TIMEOUT`

### Supabase Postgres

- `DATABASE_URL`
- `DATABASE_HOST`
- `DATABASE_PORT`
- `DATABASE_NAME`
- `DATABASE_USER`
- `DATABASE_PASSWORD`
- `DATABASE_SSL_MODE`
- `DATABASE_MAX_OPEN_CONNS`
- `DATABASE_MAX_IDLE_CONNS`
- `DATABASE_CONN_MAX_LIFETIME`

The app may use `DATABASE_URL` as the primary DSN with the split fields kept for clarity and operator convenience.

### Container Runtime

- `DOCKER_IMAGE_NAME`
- `DOCKER_IMAGE_TAG`
- `COMPOSE_PROJECT_NAME`

### Scheduler

- `CRAWL_INTERVAL_SECONDS`
- `PROCESS_INTERVAL_SECONDS`
- `PUBLISH_INTERVAL_SECONDS`
- `TWITTER_PUBLISH_INTERVAL_SECONDS`

### Feature Flags

- `AUTO_PUBLISH`
- `ENABLE_TELEGRAM_CRAWLER`
- `ENABLE_TWITTER_CRAWLER`
- `ENABLE_DEEPSEEK_PROCESSOR`
- `ENABLE_TWITTER_PUBLISH_VI`
- `ENABLE_TWITTER_PUBLISH_EN`

### Telegram

- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_TARGET_CHANNEL`
- `ADMIN_USER_IDS`
- `TELEGRAM_API_ID`
- `TELEGRAM_API_HASH`
- `TELEGRAM_SESSION`

### Twitter

- `TWITTER_BEARER_TOKEN`
- `TWITTER_VI_API_KEY`
- `TWITTER_VI_API_SECRET`
- `TWITTER_VI_ACCESS_TOKEN`
- `TWITTER_VI_ACCESS_SECRET`
- `TWITTER_EN_API_KEY`
- `TWITTER_EN_API_SECRET`
- `TWITTER_EN_ACCESS_TOKEN`
- `TWITTER_EN_ACCESS_SECRET`

### DeepSeek

- `DEEPSEEK_API_KEY`
- `DEEPSEEK_MODEL`

## Delivery Sequencing

### Wave 1

- config contract
- migration files
- repository layer
- initial domain/application model for source and content
- `cmd/api`, `cmd/worker`, `cmd/migrate` skeletons
- `Dockerfile` for multi-stage build
- `docker-compose.yml` for API, worker, and migrate workflows
- health endpoint and basic read/admin API

### Wave 2

- Telegram bot
- Telegram crawler
- initial ingest flow

### Wave 3

- DeepSeek processing
- Telegram publish
- Twitter publish
- retry and failure handling expansion

## Risks

- Porting Prisma/SQLite assumptions directly into GORM/Postgres would create incorrect nullability, ID handling, and migration habits.
- Implementing external integrations before the worker and schema contracts exist would create unstable boundaries.
- Treating Supabase as something other than plain Postgres for runtime access would complicate the architecture without benefit.
- Exposing GORM models directly through HTTP would violate the repository's clean architecture rules.
- Overloading Compose with local-only assumptions would create drift from the intended Supabase-backed deployment path.
- Treating the migrate container as a long-running service instead of a controlled job would make schema changes harder to reason about.

## Verification Strategy

Wave 1 should be considered ready only when the following can be demonstrated:

- configuration loads successfully from `config/.env`
- migration command can apply the schema to Supabase Postgres
- API process boots and exposes a working health endpoint
- repository smoke checks validate inserts and uniqueness constraints for `sources` and `content_items`
- worker process boots and reports the state of not-yet-implemented adapters explicitly
- Docker image builds successfully from the root `Dockerfile`
- `docker compose` can start the API and worker services using `config/.env`

## Change Tracking Requirements

Implementation planning and execution for this design should create:

- `changelogs/changes/<slug>/proposal.md`
- `changelogs/changes/<slug>/tasks.md`
- migration plan documentation under `docs/`
- progress and changelog updates after meaningful changes

## Open Decisions Deferred To Implementation Planning

- exact package names under `pkg/` for config/bootstrap/database helpers
- whether `logs` should be a table from day 1 or replaced by external logging only
- whether default `settings` rows are required at bootstrap
- whether queue endpoints should be read-only in wave 1 or include manual enqueue/retry actions
