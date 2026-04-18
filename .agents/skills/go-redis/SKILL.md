---
name: go-redis
description: Use when introducing Redis into a Go service for caching, locking, coordination, or ephemeral state while keeping clean architecture boundaries intact.
---

# Go Redis

## Overview
This skill guides Redis adoption after the core HTTP and Postgres path is stable. Redis should be introduced as an infrastructure adapter behind business-oriented ports.

## When To Use
- Adding cache-aside or read-model acceleration.
- Introducing distributed locks, rate limits, or ephemeral coordination state.
- Reviewing TTL strategy, invalidation, or cache key design.

## Rules
- Define cache or coordination interfaces from the consumer side.
- Treat Redis as optional acceleration unless the business model explicitly depends on it.
- Keep serialization formats versionable and observable.
- Use namespaced keys and explicit TTLs.
- Design invalidation from business events, not ad-hoc deletes.

## Current Phase Guidance
- Prioritize read-heavy use cases and idempotency support before advanced patterns.
- Do not let Redis become a hidden source of truth for critical transactional data.
- Document fallback behavior when Redis is unavailable.

## Common Mistakes
- Caching ORM structs directly without versioning.
- No TTL, no invalidation plan, and no metrics.
- Domain code knowing Redis key formats.
