---
name: go-backend-patterns
description: Use when designing or implementing Go service features that span HTTP handlers, application services, domain logic, repositories, and platform packages.
---

# Go Backend Patterns

## Overview
Use this skill to keep Go services consistent across `cmd/`, `internal/`, and `pkg/`. It favors explicit wiring, constructor injection, `context.Context`, and small interfaces defined near their consumers.

## When To Use
- New endpoints or business flows touch more than one layer.
- A change risks leaking Gin, GORM, Redis, or transport details into domain code.
- A team needs one repeatable shape for handlers, use cases, repositories, and platform helpers.

## Core Pattern
Dependency rule:

```text
Presentation -> Application -> Domain <- Infrastructure
                         ^
                         |
                        pkg
```

- `cmd/`: process entrypoints and composition roots.
- `internal/domain`: entities, value objects, domain services, domain events, repository ports.
- `internal/application`: commands, queries, use cases, DTOs, transactions, orchestration.
- `internal/infrastructure`: persistence, messaging, cache, external SDK adapters.
- `internal/presentation`: Gin handlers, request/response DTO mapping, middleware registration.
- `pkg/`: reusable platform utilities that are not business-specific.

## Implementation Rules
- Pass `context.Context` through every boundary that can block, call I/O, or emit telemetry.
- Keep interfaces on the consumer side. Repository ports belong in domain or application, not in infrastructure.
- Map transport DTOs to application inputs at the edge. Domain entities should not know about JSON, HTTP, or SQL tags unless there is a deliberate tradeoff.
- Prefer explicit constructor graphs in `cmd/*` over hidden registries or global state.
- Keep transactions in application services or dedicated unit-of-work helpers, not in handlers.
- Emit domain events from aggregates or application services, then hand them to outbox/messaging adapters.

## Common Mistakes
- Putting business rules in Gin handlers or GORM hooks.
- Reusing persistence models as domain entities by default.
- Making `pkg/` a dumping ground for arbitrary helpers.
- Returning transport-specific error payloads from application services.
