# Proposal: Content Bot Go Foundation

## Summary

Create the first Go-native implementation foundation for the `public/content-bot-share` sample by introducing Gin, GORM, Supabase Postgres, SQL migrations, and container deployment artifacts in the main repository.

## Why

The current repository contains the Go operating model and documentation but does not yet contain runnable service code. The TypeScript sample proves business intent and schema shape, but it relies on Prisma and SQLite. A Go-first runtime foundation is required before Telegram, Twitter, and DeepSeek integrations can be ported safely.

## Scope

- add `cmd/api`, `cmd/worker`, and `cmd/migrate`
- define config loading and placeholder env files
- introduce Supabase Postgres access through GORM
- create SQL migrations for `sources`, `content_items`, `settings`, and `logs`
- implement initial `source` and `content` bounded contexts
- add `Dockerfile` and `docker-compose.yml` for deployment and operator workflows

## Out Of Scope

- full Telegram bot behavior
- real Telegram/Twitter crawling
- real DeepSeek processing
- production CI/CD and infrastructure automation

## Risks

- schema assumptions from Prisma/SQLite may not map cleanly without explicit migration design
- worker and API boundaries may drift if queue/status semantics are not defined early
- Compose can become misleading if it models a local database that the deployment path does not actually use

## Verification

- config parsing tests pass
- migrations apply to Supabase Postgres
- `cmd/api`, `cmd/worker`, and `cmd/migrate` build
- health endpoint boots
- container build and Compose validation succeed

