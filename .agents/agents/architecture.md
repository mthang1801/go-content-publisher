---
name: architecture
---

# Go Service Template Architecture

This file is the structural source of truth for this repository. It defines how the sample should be organized once service code is added and how agents should decide file placement before implementation begins.

## Root Structure

```text
./
├── .agents/
│   ├── agents/                 # team model, routing, operating contracts
│   ├── skills/                 # reusable Go-service guidance
│   ├── workflows/              # required execution sequences
│   ├── rules/                  # hard constraints
│   ├── contexts/               # operating modes
│   ├── prompts/                # reusable prompt scaffolds
│   ├── hooks/                  # optional automation hooks
│   ├── scripts/                # local helper scripts
│   └── specs/                  # agent-system evolution docs
├── cmd/
│   ├── api/                    # HTTP API process
│   ├── worker/                 # async worker / outbox / consumers / jobs
│   ├── migrate/                # schema migration and controlled data fixes
│   └── cli/                    # admin or operator-oriented commands
├── internal/
│   ├── <bounded-context>/
│   │   ├── domain/
│   │   │   ├── entities/
│   │   │   ├── valueobjects/
│   │   │   ├── events/
│   │   │   ├── services/
│   │   │   ├── policies/
│   │   │   └── repositories/
│   │   ├── application/
│   │   │   ├── commands/
│   │   │   ├── queries/
│   │   │   ├── usecases/
│   │   │   ├── services/
│   │   │   ├── dto/
│   │   │   └── ports/
│   │   ├── infrastructure/
│   │   │   ├── persistence/
│   │   │   │   ├── postgres/
│   │   │   │   ├── models/
│   │   │   │   └── repositories/
│   │   │   ├── cache/
│   │   │   ├── messaging/
│   │   │   ├── jobs/
│   │   │   └── clients/
│   │   └── presentation/
│   │       └── http/
│   │           ├── handlers/
│   │           ├── requests/
│   │           ├── responses/
│   │           ├── routes.go
│   │           └── middleware.go
│   └── <service-private-shared>/ # only when reuse is local to this service
├── pkg/
│   ├── app/
│   │   ├── bootstrap/
│   │   ├── config/
│   │   ├── lifecycle/
│   │   └── server/
│   ├── ddd/
│   │   ├── aggregate/
│   │   ├── entity/
│   │   ├── event/
│   │   ├── repository/
│   │   └── valueobject/
│   ├── database/
│   │   ├── postgres/
│   │   ├── gormx/
│   │   ├── tx/
│   │   └── migration/
│   ├── cache/
│   │   └── redis/
│   ├── messaging/
│   │   ├── outbox/
│   │   ├── inbox/
│   │   ├── kafka/
│   │   ├── rabbitmq/
│   │   └── consumer/
│   ├── transport/
│   │   └── http/
│   │       ├── errors/
│   │       ├── middleware/
│   │       ├── pagination/
│   │       └── response/
│   ├── observability/
│   │   ├── logger/
│   │   ├── tracing/
│   │   ├── metrics/
│   │   └── health/
│   ├── worker/
│   │   ├── runner/
│   │   ├── retry/
│   │   └── idempotency/
│   ├── testing/
│   │   ├── builders/
│   │   ├── containers/
│   │   └── fixtures/
│   └── util/
├── test/
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
├── docs/
│   └── plan/
│       └── progress.md
└── changelogs/
    ├── CHANGELOG.md
    └── changes/
```

## Architectural Direction

```text
Presentation -> Application -> Domain <- Infrastructure
```

This rule is strict. When code does not fit cleanly, revisit the design rather than forcing a shortcut.

## Root Folder Responsibilities

### `cmd/`

- owns process entrypoints and composition roots
- wires dependencies, config, transports, and graceful shutdown
- contains startup logic, not business decisions

### `internal/`

- owns service-specific code and bounded contexts
- is the default home for business behavior
- may contain local shared helpers only when reuse is real and still private to this service

### `pkg/`

- owns reusable platform/kernel code inside the sample
- may later be extracted or shared across services
- must not depend on service-specific `internal/*` packages

### `docs/` and `changelogs/`

- `docs/plan/progress.md` tracks current phase and near-term milestones
- `changelogs/CHANGELOG.md` records durable history
- `changelogs/changes/<slug>/` holds proposal/tasks/design/evidence for non-trivial changes

## Bounded Context Shape

Each bounded context under `internal/` should be understandable in isolation.

Recommended shape:

```text
internal/<bounded-context>/
├── domain/
├── application/
├── infrastructure/
└── presentation/
```

### `domain/`

Put here:

- aggregates and entities
- value objects
- domain events
- domain services
- policies and specifications
- repository ports that belong to the business model

Do not put here:

- Gin handlers
- GORM models
- SQL, Redis, Kafka, RabbitMQ, HTTP-client code

### `application/`

Put here:

- use cases and handlers for commands/queries
- orchestration across aggregates or ports
- transaction coordination
- DTOs used between presentation and application
- ports owned by application-level workflows

Do not put here:

- transport binding logic
- ORM schemas or broker clients
- domain invariants that belong inside domain types

### `infrastructure/`

Put here:

- Postgres/GORM repositories
- Redis adapters
- outbox/inbox publishers and consumers
- external service clients
- job runners and broker-specific code

Do not put here:

- core business decisions that should remain true without infrastructure

### `presentation/`

Put here:

- Gin route registration
- request parsing and validation
- response mapping
- transport-scoped middleware binding

Do not put here:

- business workflows
- transaction coordination
- direct persistence access that bypasses application

## Placement Matrix

| Concern | Default Location | Notes |
| --- | --- | --- |
| Aggregate behavior | `internal/<bc>/domain` | state transitions and invariants stay here |
| Use-case orchestration | `internal/<bc>/application` | coordinates ports, transactions, events |
| GORM model | `internal/<bc>/infrastructure/persistence/models` | never treat as domain entity by default |
| Repository adapter | `internal/<bc>/infrastructure/persistence/repositories` | implements business-owned ports |
| Gin handler | `internal/<bc>/presentation/http/handlers` | transport edge only |
| Response envelope / middleware primitive | `pkg/transport/http` | only if reusable across contexts |
| Config/bootstrap/lifecycle | `pkg/app` and `cmd/*` | `cmd/*` composes, `pkg/app` provides reusable bootstrap blocks |
| Logger/tracing/health | `pkg/observability` | reusable runtime concerns |
| Outbox, inbox, broker manager | `pkg/messaging` plus service adapters | generic mechanism in `pkg`, service binding in `internal/*/infrastructure` |
| Test fixtures/builders | nearest owning package or `pkg/testing` | keep local unless clearly cross-context |

## Runtime Topology

The sample is expected to evolve into multiple binaries even if early phases run lightly.

### `cmd/api`

- HTTP server
- health/readiness endpoints
- route registration
- request-scoped middleware and observability bootstrapping

### `cmd/worker`

- outbox processors
- async consumers
- scheduled jobs
- long-running background workflows

### `cmd/migrate`

- schema migration
- controlled data backfill or repair tasks when explicitly designed

### `cmd/cli`

- operator/admin utilities
- one-off commands that should not live inside API or worker processes

## Persistence And Messaging Direction

### Postgres + GORM

- Postgres is the source of truth for transactional state in phase 2
- GORM is an adapter convenience, not a domain model
- migrations are the schema source of truth
- transactions should be explicit in application services or transaction helpers

### Redis

- introduced behind ports for cache, locking, idempotency, or ephemeral coordination
- should not become a hidden source of truth for critical data

### Kafka and RabbitMQ

- introduced later through adapter-manager patterns under `pkg/messaging`
- service-specific publishers/consumers live under `internal/<bc>/infrastructure/messaging`
- transport envelopes and broker payloads must not leak into domain types

### Outbox / Inbox

- preferred for reliable async delivery and idempotent consumption
- mechanism belongs in reusable platform code
- service-specific event mapping belongs near the bounded context that emits or consumes

## Testing Layout

Testing should be colocated by default, with shared harness helpers only when justified.

- package-local `_test.go` files for domain, application, and simple handler logic
- `test/integration/` for dependency-backed verification that spans package seams
- `test/e2e/` for critical end-to-end flows only
- `pkg/testing/` for reusable builders, fixtures, or container helpers

## Documentation And Change Layout

- `.agents/` documents how agents should work
- `docs/plan/progress.md` captures current state, milestones, and risks
- `changelogs/CHANGELOG.md` summarizes durable repo evolution
- `changelogs/changes/<slug>/` captures non-trivial change records

## Blocked Anti-Patterns

- putting business logic in Gin handlers
- placing GORM models in domain packages
- creating a generic `shared` or `common` bucket with no ownership
- letting `pkg/` import service-specific code
- hiding transactions inside generic repository helpers
- treating worker or broker payloads as domain events without translation
- inventing abstractions for Kafka/RabbitMQ before the second real use case exists

## Go-Native Constraints

- no framework imports in domain packages
- no hidden DI-container assumptions
- `context.Context` should cross all blocking or request-scoped boundaries
- GORM models must not become domain entities by convenience
- handlers and consumers must stay thin
- interfaces should usually live near the code that consumes them
