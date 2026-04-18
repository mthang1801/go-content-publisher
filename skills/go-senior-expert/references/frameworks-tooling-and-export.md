# Frameworks, Tooling & Export in Go

Derived from:
- `documents/assets/go/gin/README.md`
- `documents/assets/go/fiber/README.md`
- `documents/assets/go/orm/README.md`
- `documents/assets/go/cli/README.md`
- `documents/assets/go/export/README.md`

## When To Read

Load this reference when the task is about HTTP framework behavior, persistence adapters, CLI ergonomics, or export pipelines rather than core language semantics.

## HTTP Delivery

- Gin and Fiber are delivery adapters. Keep handlers thin: bind, validate, authorize, dispatch, marshal.
- Use middleware for logging, recovery, request IDs, auth, rate limits, and cross-cutting concerns.
- Treat file upload/download, SSE, WebSocket, and streaming responses as lifecycle-sensitive paths; cancellation and shutdown must be explicit.
- If the user problem is request flow, binding/validation, auth edge, or shutdown/streaming, route deeper into the matching `gin/` or `fiber/` lane.

## ORM & Persistence

- Keep GORM in infrastructure. Domain aggregates should not depend on ORM tags or lazy-loading behavior.
- Separate persistence models from domain entities when invariants or lifecycle semantics differ.
- Review transactions, associations, and hooks carefully; they are correctness boundaries, not convenience features.
- If the symptom is query shape, preload behavior, CRUD workflow, or transaction semantics, route into `orm/`.

## CLI Tooling

- Use Cobra/Viper when there is a real command tree and layered config problem, not for a trivial one-command binary.
- Make precedence explicit: flags over env over config file over defaults, unless the product contract says otherwise.
- Treat secrets handling and shell completion as operator UX and safety work, not afterthoughts.

## Export Architecture

- First choose the format: CSV, Excel, PDF, or raw streaming.
- Then choose the execution model: inline streaming for small/fast exports, background jobs for large or long-running jobs.
- Large exports need progress tracking, retry-aware workers, object storage, and signed delivery URLs.
- Never accumulate million-row datasets in memory just because the format is "simple".
