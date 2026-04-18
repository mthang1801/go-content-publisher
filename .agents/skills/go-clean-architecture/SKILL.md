---
name: go-clean-architecture
description: Use when structuring Go packages, reviewing layer boundaries, or preventing framework and persistence details from leaking into business logic.
---

# Go Clean Architecture

## Overview
This skill protects import direction, layer ownership, and package shape. Use it whenever code needs to be placed in the right boundary or when a refactor risks circular dependencies.

## When To Use
- A new package is being introduced.
- Imports feel tangled or circular.
- Handlers, services, or repositories are growing mixed responsibilities.
- A review needs a hard check on dependency direction.

## Boundary Rules
- Domain imports only standard library and stable internal primitives that do not depend on infrastructure.
- Application imports domain and package-level utilities needed for orchestration.
- Infrastructure imports domain/application ports and implements them.
- Presentation imports application contracts and maps HTTP input/output.
- `pkg/` may be reused by multiple services, but it must not depend on any service-specific domain package.

## Placement Guide
- Put domain invariants and state transitions in `internal/domain`.
- Put use-case orchestration, transactions, and policy coordination in `internal/application`.
- Put DB/cache/broker/SDK code in `internal/infrastructure`.
- Put router setup, middleware binding, request validation, and response mapping in `internal/presentation`.
- Put logging, config, HTTP server bootstrapping, middleware primitives, tracing, and reusable adapters in `pkg/`.

## Review Checklist
- Can domain code compile without Gin, GORM, Redis, Kafka, or RabbitMQ imports?
- Would moving infrastructure packages out of the repo leave the domain intact?
- Does `cmd/*` wire dependencies without embedding business logic?
- Are package names based on business capability, not technical layers only?

## Common Mistakes
- Shared models imported everywhere because they are convenient.
- `pkg/` importing `internal/*` or domain packages.
- Application services returning Gin contexts, SQL errors, or ORM entities.
