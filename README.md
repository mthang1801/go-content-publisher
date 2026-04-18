# Go Sample Service Template

This repository is the first reference implementation of a reusable Go service template.

It exists for two purposes:

1. prove a Go-native Clean Architecture + DDD service layout that can evolve into real services
2. provide a reusable `.agents` operating system for future Go service repositories

## Principles

### Clean Architecture

Dependency direction is fixed:

```text
Presentation -> Application -> Domain <- Infrastructure
```

- `domain` holds business rules and imports no framework code
- `application` orchestrates use cases and domain interactions
- `infrastructure` implements domain ports and external adapters
- `presentation` accepts requests, validates format, delegates, and responds

### Go-Native Structure

This sample follows Go conventions rather than importing NestJS mental models:

- `cmd/` contains runnable entrypoints
- `internal/` contains service-private application code
- `pkg/` contains reusable platform/kernel code local to this sample
- `context.Context` is propagated explicitly
- constructor injection is preferred over container magic

### Why `pkg/` Is The Platform Layer Here

Inside this sample, `pkg/` acts as the internal platform layer that hosts reusable building blocks for:

- CQRS interfaces and helpers
- DDD primitives
- transport-neutral response and error shapes
- configuration and wiring helpers
- shared adapters that may later be promoted into standalone reusable modules

This keeps the sample self-contained and easy to evolve independently from the root repository.

## `.agents` Operating Model

`.agents/` is the only source of truth for AI-agent collaboration in this sample.

It provides:

- generic Go-service agent roles
- Go-native skills
- reusable workflows
- architecture and coding rules
- contexts, prompts, hooks, and helper scripts
- planning and changelog conventions

The template is inspired by `examples/nestjs-sample/.agents`, but rewritten for Go. It is not a direct port.

## Folder Structure

```text
./
├── .agents/          # source-of-truth AI operating system for Go services
├── changelogs/       # tracked implementation history and change proposals
├── docs/             # planning and service-local documentation
├── cmd/              # binaries (phase 2)
├── internal/         # service-private code (phase 2)
└── pkg/              # reusable platform/kernel code (phase 2)
```

Current phase focuses on `.agents`, `docs/plan`, and `changelogs` first. Code scaffolding comes after the operating model is stable.

## Development Workflow

1. Read `.agents/AGENTS.md` first.
2. Read `docs/plan/progress.md` and `changelogs/CHANGELOG.md` before planning or implementation.
3. For non-trivial work, create `changelogs/changes/<slug>/proposal.md` and `tasks.md`.
4. Follow the relevant workflow in `.agents/workflows/`.
5. Load broad guidance from `skills/` first, then repo-local precision skills from `.agents/skills/`.
6. Keep `progress.md` and `CHANGELOG.md` updated after meaningful work.

This sample keeps the full operating model:

- orchestrator-first planning
- spec-first changes
- TDD and review gates
- reusable specialist agents

## Roadmap

### Phase 1

Build `.agents` as a reusable Go service template.

### Phase 2

Scaffold the code sample:

- `cmd/api`
- `cmd/worker`
- `cmd/migrate`
- `cmd/cli`
- `internal/...`
- `pkg/...`

### Phase 3

Add Redis support.

### Phase 4

Add Kafka and RabbitMQ via a more deliberate adapter-manager design.

## Scope Notes

This sample will eventually demonstrate a commerce flow built around:

- customer
- order
- product
- inventory
- promotion

The first end-to-end business flow will be `Place Order`, but that comes after the `.agents` system is in place.
