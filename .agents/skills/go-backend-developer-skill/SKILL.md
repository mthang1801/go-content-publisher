---
name: go-backend-developer-skill
description: Use when implementing Go service behavior across handlers, use cases, domain logic, repositories, cache, and worker seams and you want one entry skill for the implementation stack.
---

# Go Backend Developer Skill

## Overview

This is the primary entry skill for `go-backend-developer`.

It loads the implementation stack in the right order: broad Go/backend guidance first, then repository-local architecture, persistence, error handling, testing, and runtime patterns.

## When To Use

- adding or changing handlers, requests, or responses
- implementing application services, use cases, or domain behavior
- adding repository adapters, transactions, or persistence mapping
- wiring cache, outbox, or worker flows behind ports
- refactoring service code while preserving behavior

## Skill Stack

### Copied Base Skills

- `skills/go-senior-expert/SKILL.md`
- `skills/backend-patterns-skill/SKILL.md`
- `skills/senior-dev/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/go-senior-expert/SKILL.md`
- `.agents/skills/backend-patterns-skill/SKILL.md`
- `.agents/skills/senior-dev/SKILL.md`

### Precise Local Skills

- `.agents/skills/go-backend-patterns/SKILL.md`
- `.agents/skills/go-clean-architecture/SKILL.md`
- `.agents/skills/go-ddd/SKILL.md`
- `.agents/skills/go-gorm-postgres/SKILL.md`
- `.agents/skills/go-error-handling/SKILL.md`
- `.agents/skills/go-testing-tdd/SKILL.md`
- `.agents/skills/go-redis/SKILL.md`
- `.agents/skills/go-microservices/SKILL.md`
- `.agents/skills/go-saga/SKILL.md`
- `.agents/skills/go-stream-pipeline/SKILL.md`
- `.agents/skills/git-workflow/SKILL.md`
- `.agents/skills/security-review/SKILL.md`

## Loading Order

1. load broad Go and backend-delivery base skills
2. load bridge skills that normalize them for this repository
3. load only the precise local skills needed by the active runtime surface

## Workflow Focus

- keep handlers thin and application-owned orchestration explicit
- keep domain invariants inside domain types or domain services
- keep infrastructure details behind ports and adapters
- keep transaction, error, cache, and worker semantics intentional

## Output Contract

Produce implementation-ready guidance or code changes that make clear:

- owning layer
- runtime surface
- transaction and error behavior
- verification steps
- required follow-up review

## Escalation

- route to `go-architect` if package ownership or reusable extraction is unclear
- route to `go-test-writer` when verification design is not obvious yet
- route to `go-db-optimizer` for query, migration, or transaction-heavy work

## Common Mistakes

- placing business logic in handlers or ORM models
- inventing generic helpers before a real second consumer exists
- leaking infrastructure details into domain or application contracts
- coding before choosing the narrowest correct verification seam
