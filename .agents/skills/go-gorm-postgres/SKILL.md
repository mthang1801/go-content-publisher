---
name: go-gorm-postgres
description: Use when implementing or reviewing PostgreSQL persistence in Go services that use GORM, including schema mapping, transactions, indexes, and repository adapters.
---

# Go GORM Postgres

## Overview
Use this skill for repository adapters backed by PostgreSQL and GORM. It focuses on explicit schema control, predictable transactions, and clean mapping between domain models and persistence models.

## When To Use
- Adding or refactoring GORM repositories.
- Designing tables, indexes, and migrations for a Go service.
- Investigating slow queries, transaction issues, or N+1 behavior.
- Reviewing whether ORM models have leaked across boundaries.

## Implementation Rules
- Keep persistence structs in infrastructure packages unless there is a conscious shared-model tradeoff.
- Use explicit transactions in application services or unit-of-work helpers for multi-write flows.
- Create indexes and unique constraints based on domain lookup patterns, not only foreign keys.
- Prefer repository methods named by business intent, such as `FindAvailableInventoryBySKU`, over generic ORM helpers.
- Use migrations as source of truth for schema evolution. Do not rely on auto-migrate in production workflows.

## Query Guidance
- Select only needed columns for projections and list endpoints.
- Use preloading carefully; avoid blind eager loading.
- Encode idempotency and business uniqueness with constraints where possible.
- Treat soft deletes as a business choice, not a default.

## Common Mistakes
- Returning GORM models directly to handlers.
- Hiding transaction boundaries inside repositories.
- Building generic repositories that ignore aggregate semantics.
- Depending on auto-generated timestamps and hooks for core business rules.
