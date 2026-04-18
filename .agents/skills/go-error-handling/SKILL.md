---
name: go-error-handling
description: Use when defining, propagating, mapping, or reviewing errors in Go services across domain, application, HTTP, database, cache, and messaging boundaries.
---

# Go Error Handling

## Overview
Use this skill to make errors explicit, classifiable, and safe to expose. Go services should return rich internal errors and map them intentionally at the transport boundary.

## When To Use
- A new use case introduces validation or business failure paths.
- Error responses need consistent HTTP mapping.
- External dependency failures must be wrapped without losing context.

## Pattern
- Domain errors describe business conditions such as `ErrInsufficientInventory`.
- Application wraps lower-level failures with operation context.
- Presentation maps known classes to HTTP status and response codes.
- Infrastructure preserves root cause details for logs and telemetry without leaking secrets.

## Rules
- Use sentinel or typed errors only when callers must branch on them.
- Wrap with `%w` so callers can inspect causes.
- Distinguish business rejection, not-found, conflict, transient dependency, and programmer errors.
- Never build error handling around string matching alone.
- Log enough context to debug, but do not log secrets, tokens, or full sensitive payloads.

## Common Mistakes
- Returning raw SQL or Redis errors to API clients.
- Using one catch-all `internal server error` path for every failure.
- Losing stack or causal context by recreating errors without wrapping.
