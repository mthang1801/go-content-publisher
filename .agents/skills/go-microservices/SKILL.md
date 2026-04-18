---
name: go-microservices
description: Use when preparing a Go service for service extraction, inter-service contracts, outbox/event delivery, or operational boundaries beyond a single API process.
---

# Go Microservices

## Overview
Use this skill to design service boundaries early without forcing distributed complexity too soon. It helps a codebase stay extraction-ready while phase 1 may still run as one sample.

## When To Use
- Bounded contexts may later split into separate services.
- A capability needs its own data ownership or independent scaling path.
- Contracts, idempotency, and event-driven coordination must be planned.

## Design Rules
- One service owns its write model and source of truth.
- Cross-service calls should be explicit in contracts, timeouts, retries, and ownership.
- Prefer asynchronous integration for facts and workflows that do not require immediate consistency.
- Introduce outbox/inbox and idempotency keys before adding brokers.
- Keep shared code minimal; share platform packages, not domain logic.

## Common Mistakes
- Sharing domain packages across services.
- Synchronous chains for every business flow.
- Using a broker to hide unclear ownership.
