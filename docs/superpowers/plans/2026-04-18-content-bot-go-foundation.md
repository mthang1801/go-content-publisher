# Content Bot Go Foundation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first Go-native foundation for the content-bot sample using Gin, GORM, Supabase Postgres, SQL migrations, and container deployment artifacts.

**Architecture:** The implementation starts infrastructure-first. It establishes `cmd/api`, `cmd/worker`, and `cmd/migrate`; models the `source` and `content` bounded contexts; uses SQL migrations as schema source of truth; and keeps Telegram, Twitter, and DeepSeek integrations behind ports or noop adapters in wave 1.

**Tech Stack:** Go, Gin, GORM, PostgreSQL (Supabase), SQL migrations, Docker, Docker Compose

---

## File Structure Map

### New Directories

- `cmd/api`
- `cmd/worker`
- `cmd/migrate`
- `config`
- `internal/source/domain`
- `internal/source/application`
- `internal/source/infrastructure/persistence/models`
- `internal/source/infrastructure/persistence/repositories`
- `internal/source/presentation/http`
- `internal/content/domain`
- `internal/content/application`
- `internal/content/infrastructure/persistence/models`
- `internal/content/infrastructure/persistence/repositories`
- `internal/content/presentation/http`
- `pkg/app/bootstrap`
- `pkg/config`
- `pkg/database/postgres`
- `pkg/migration`
- `pkg/observability/logger`
- `db/migrations`

### New Root Files

- `go.mod`
- `go.sum`
- `Dockerfile`
- `docker-compose.yml`
- `config/.env`
- `config/.env.example`

### New Runtime Files

- `cmd/api/main.go`
- `cmd/worker/main.go`
- `cmd/migrate/main.go`
- `pkg/config/config.go`
- `pkg/app/bootstrap/app.go`
- `pkg/database/postgres/postgres.go`
- `pkg/migration/runner.go`

### New Context Files

- `internal/source/domain/source.go`
- `internal/source/domain/repository.go`
- `internal/source/application/service.go`
- `internal/source/infrastructure/persistence/models/source_model.go`
- `internal/source/infrastructure/persistence/repositories/source_repository.go`
- `internal/source/presentation/http/routes.go`
- `internal/source/presentation/http/handler.go`
- `internal/content/domain/content_item.go`
- `internal/content/domain/repository.go`
- `internal/content/application/service.go`
- `internal/content/infrastructure/persistence/models/content_item_model.go`
- `internal/content/infrastructure/persistence/repositories/content_repository.go`
- `internal/content/presentation/http/routes.go`
- `internal/content/presentation/http/handler.go`

### New Migration Files

- `db/migrations/0001_init_content_bot_schema.sql`
- `db/migrations/0002_indexes_and_uniqueness.sql`
- `db/migrations/0003_seed_settings.sql`

## Chunk 1: Bootstrap And Config

### Task 1: Initialize Go module and runtime dependency skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/api/main.go`
- Create: `cmd/worker/main.go`
- Create: `cmd/migrate/main.go`
- Create: `pkg/app/bootstrap/app.go`
- Create: `pkg/observability/logger/logger.go`

- [ ] **Step 1: Initialize the module**

Run: `go mod init github.com/mvt/go-content-bot`
Expected: a new `go.mod` exists at the repository root

- [ ] **Step 2: Add initial dependencies**

Run:
```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/joho/godotenv
```

Expected: `go.mod` and `go.sum` include the minimum runtime dependencies

- [ ] **Step 3: Create minimal process entrypoints**

Use these initial shapes:

```go
func main() {
    app, err := bootstrap.New()
    if err != nil {
        log.Fatal(err)
    }
    if err := app.RunAPI(); err != nil {
        log.Fatal(err)
    }
}
```

```go
func main() {
    app, err := bootstrap.New()
    if err != nil {
        log.Fatal(err)
    }
    if err := app.RunWorker(); err != nil {
        log.Fatal(err)
    }
}
```

```go
func main() {
    app, err := bootstrap.New()
    if err != nil {
        log.Fatal(err)
    }
    if err := app.RunMigrations(); err != nil {
        log.Fatal(err)
    }
}
```

- [ ] **Step 4: Build the three binaries**

Run:
```bash
go build ./cmd/api
go build ./cmd/worker
go build ./cmd/migrate
```

Expected: all three builds succeed, even if behavior is still skeletal

### Task 2: Create the config contract and placeholder env files

**Files:**
- Create: `config/.env`
- Create: `config/.env.example`
- Create: `pkg/config/config.go`

- [ ] **Step 1: Define the config struct**

Use explicit sections and typed values:

```go
type Config struct {
    App       AppConfig
    HTTP      HTTPConfig
    Database  DatabaseConfig
    Scheduler SchedulerConfig
    Feature   FeatureConfig
    Telegram  TelegramConfig
    Twitter   TwitterConfig
    DeepSeek  DeepSeekConfig
}
```

- [ ] **Step 2: Implement config loading from environment**

Requirements:
- support `DATABASE_URL` as the primary DSN
- validate required app and database values
- parse durations and integer intervals explicitly
- never log secret values

- [ ] **Step 3: Create `config/.env.example`**

Include placeholders for:
- app/server settings
- Supabase Postgres settings
- scheduler intervals
- feature flags
- Telegram, Twitter, and DeepSeek credentials
- Docker/Compose naming values

- [ ] **Step 4: Create `config/.env` from the example**

Populate with empty or obvious placeholder values so the user can fill them manually.

- [ ] **Step 5: Verify config parsing**

Run: `go test ./pkg/config/...`
Expected: parsing and required-key validation tests pass

## Chunk 2: Database And Migration Path

### Task 3: Add PostgreSQL bootstrap and migration runner

**Files:**
- Create: `pkg/database/postgres/postgres.go`
- Create: `pkg/migration/runner.go`
- Modify: `pkg/app/bootstrap/app.go`

- [ ] **Step 1: Implement Postgres connection setup**

Required behavior:
- connect using Supabase Postgres
- set max open conns, max idle conns, and conn max lifetime
- expose `*gorm.DB` and close behavior via the underlying `sql.DB`

- [ ] **Step 2: Implement migration discovery**

Use a simple ordered file runner:

```go
type Runner interface {
    Up(ctx context.Context, db *sql.DB) error
}
```

The runner should apply `.sql` files from `db/migrations` in lexical order.

- [ ] **Step 3: Track applied migrations**

Create and use a schema history table such as:

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version text PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
);
```

- [ ] **Step 4: Verify database connectivity and migration bootstrap**

Run:
```bash
go test ./pkg/database/...
go build ./cmd/migrate
```

Expected: connection/bootstrap code builds and tests pass

### Task 4: Write SQL migrations for the Prisma-to-Postgres conversion

**Files:**
- Create: `db/migrations/0001_init_content_bot_schema.sql`
- Create: `db/migrations/0002_indexes_and_uniqueness.sql`
- Create: `db/migrations/0003_seed_settings.sql`

- [ ] **Step 1: Create the base schema migration**

`0001_init_content_bot_schema.sql` must create:
- `sources`
- `content_items`
- `settings`
- `logs`

Use:
- UUID primary keys
- timestamptz for timestamps
- nullable columns only where the original pipeline requires them
- foreign key from `content_items.source_id` to `sources.id`

- [ ] **Step 2: Create the index and uniqueness migration**

`0002_indexes_and_uniqueness.sql` must create:
- unique `(type, handle)` on `sources`
- unique `(source_id, external_id)` on `content_items`
- indexes on `content_items.status`, `content_items.crawled_at`, and `logs.created_at`

- [ ] **Step 3: Create the optional seed migration**

`0003_seed_settings.sql` should seed only the smallest required bootstrap settings and remain idempotent.

- [ ] **Step 4: Verify migrations against Supabase**

Run: `go run ./cmd/migrate`
Expected: schema applies successfully to the target Supabase database

## Chunk 3: Bounded Context Foundations

### Task 5: Implement the `source` context

**Files:**
- Create: `internal/source/domain/source.go`
- Create: `internal/source/domain/repository.go`
- Create: `internal/source/application/service.go`
- Create: `internal/source/infrastructure/persistence/models/source_model.go`
- Create: `internal/source/infrastructure/persistence/repositories/source_repository.go`

- [ ] **Step 1: Define the domain entity and repository port**

Use a focused API:

```go
type Repository interface {
    Create(ctx context.Context, source Source) error
    ListActive(ctx context.Context) ([]Source, error)
    DeleteByHandle(ctx context.Context, sourceType, handle string) error
}
```

- [ ] **Step 2: Map the persistence model**

Keep the GORM model in infrastructure, not in domain.

- [ ] **Step 3: Implement the repository adapter**

Requirements:
- translate unique constraint violations into domain/application errors
- avoid leaking GORM models upward
- keep queries aligned with business intent

- [ ] **Step 4: Write repository tests**

Run: `go test ./internal/source/...`
Expected: create/list/delete behaviors pass against the chosen test setup

### Task 6: Implement the `content` context and pipeline state model

**Files:**
- Create: `internal/content/domain/content_item.go`
- Create: `internal/content/domain/repository.go`
- Create: `internal/content/application/service.go`
- Create: `internal/content/infrastructure/persistence/models/content_item_model.go`
- Create: `internal/content/infrastructure/persistence/repositories/content_repository.go`

- [ ] **Step 1: Define pipeline statuses explicitly**

Represent statuses from the sample:
- `pending`
- `processing`
- `rewritten`
- `published`
- `failed`
- `skipped`

- [ ] **Step 2: Encode allowed state transitions**

Use domain methods such as:

```go
func (c *ContentItem) MarkProcessing() error
func (c *ContentItem) MarkRewritten(text string) error
func (c *ContentItem) MarkPublished(messageID string, publishedAt time.Time) error
func (c *ContentItem) MarkFailed(reason string) error
```

- [ ] **Step 3: Implement queue-oriented repository methods**

Examples:

```go
type Repository interface {
    CreatePending(ctx context.Context, item ContentItem) error
    FindNextPending(ctx context.Context) (*ContentItem, error)
    Save(ctx context.Context, item ContentItem) error
    ListRecent(ctx context.Context, limit int) ([]ContentItem, error)
}
```

- [ ] **Step 4: Write tests for state transitions and uniqueness**

Run:
```bash
go test ./internal/content/...
```

Expected: invalid status transitions fail and uniqueness is enforced

## Chunk 4: API And Worker Processes

### Task 7: Add the initial Gin HTTP surface

**Files:**
- Create: `internal/source/presentation/http/routes.go`
- Create: `internal/source/presentation/http/handler.go`
- Create: `internal/content/presentation/http/routes.go`
- Create: `internal/content/presentation/http/handler.go`
- Modify: `cmd/api/main.go`
- Modify: `pkg/app/bootstrap/app.go`

- [ ] **Step 1: Add a health endpoint**

Required route:
- `GET /healthz`

Expected response:

```json
{"status":"ok"}
```

- [ ] **Step 2: Add source management endpoints**

Initial wave:
- `GET /api/sources`
- `POST /api/sources`
- `DELETE /api/sources/:type/:handle`

- [ ] **Step 3: Add read-only content endpoints**

Initial wave:
- `GET /api/content/queue`
- `GET /api/content/recent`

- [ ] **Step 4: Verify the API process**

Run:
```bash
go build ./cmd/api
go test ./internal/... ./pkg/...
```

Expected: API compiles and tests pass

### Task 8: Add the worker skeleton with noop integration adapters

**Files:**
- Modify: `cmd/worker/main.go`
- Create: `internal/content/application/worker_service.go`
- Create: `internal/content/application/ports.go`

- [ ] **Step 1: Define external integration ports**

Ports should cover:
- crawl ingestion
- rewrite processing
- publish dispatch

Wave 1 implementations can be noop adapters returning explicit "not enabled" behavior.

- [ ] **Step 2: Implement scheduler loops**

Use config-driven intervals for:
- crawl
- process
- publish
- Twitter publish

- [ ] **Step 3: Make worker startup explicit**

Worker logs must clearly state:
- which loops are enabled
- which adapters are disabled/noop
- whether auto-publish is enabled

- [ ] **Step 4: Verify the worker binary**

Run:
```bash
go build ./cmd/worker
```

Expected: worker binary builds and starts without panicking when adapters are not configured

## Chunk 5: Containerization And Delivery

### Task 9: Add container build and Compose deployment files

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`

- [ ] **Step 1: Write a multi-stage Dockerfile**

Requirements:
- build Go binaries in a builder stage
- produce a minimal runtime image
- default to the API process
- allow command override for worker and migrate

- [ ] **Step 2: Write `docker-compose.yml`**

Services:
- `api`
- `worker`
- `migrate`

Requirements:
- load `config/.env`
- connect to Supabase Postgres via environment variables
- avoid provisioning a local Postgres for the main deployment path
- reuse the same built image where practical

- [ ] **Step 3: Validate Compose configuration**

Run:
```bash
docker compose config
docker build -t content-bot-go:dev .
```

Expected: Compose is valid and the image builds successfully

## Chunk 6: Tracking, Docs, And Final Verification

### Task 10: Sync docs and run end-to-end verification

**Files:**
- Modify: `docs/plan/progress.md`
- Modify: `changelogs/CHANGELOG.md`
- Modify: `changelogs/changes/content-bot-go-foundation/tasks.md`

- [ ] **Step 1: Update progress and changelog with actual implemented scope**

Do not describe unimplemented integration features as complete.

- [ ] **Step 2: Run the verification suite**

Run:
```bash
go test ./...
go build ./cmd/api ./cmd/worker ./cmd/migrate
docker compose config
docker build -t content-bot-go:dev .
```

Expected: tests, builds, and container validation pass

- [ ] **Step 3: Record evidence in the change record**

Add the exact verification commands and outcomes to the change record or its evidence notes.

