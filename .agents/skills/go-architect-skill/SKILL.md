---
name: go-architect-skill
description: Use when deciding package boundaries, reusable patterns, extraction seams, runtime topology, or clean-architecture trade-offs for Go services and you want one entry skill for the architecture stack.
---

# Go Architect Skill

## Overview

This is the primary entry skill for `go-architect`.

It loads the broad Go and system-design judgment first, then narrows into repository-specific clean-architecture, DDD, and extraction guidance.

## When To Use

- a new package, module, or bounded context is being introduced
- a change crosses `cmd/`, `internal/`, and `pkg/`
- architectural trade-offs must be evaluated explicitly
- service extraction or future distribution seams are under discussion
- a refactor risks circular dependencies or layer leaks

## Skill Stack

### Copied Base Skills

- `skills/go-senior-expert/SKILL.md`
- `skills/backend-patterns-skill/SKILL.md`
- `skills/senior-sysdesign/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/go-senior-expert/SKILL.md`
- `.agents/skills/backend-patterns-skill/SKILL.md`
- `.agents/skills/senior-sysdesign/SKILL.md`

### Precise Local Skills

- `.agents/skills/go-clean-architecture/SKILL.md`
- `.agents/skills/go-ddd/SKILL.md`
- `.agents/skills/go-backend-patterns/SKILL.md`
- `.agents/skills/go-microservices/SKILL.md`
- `.agents/skills/go-saga/SKILL.md`
- `.agents/skills/security-review/SKILL.md`

## Loading Order

1. load broad Go and system-design base skills
2. load bridge skills that reinterpret them for this repository
3. load precise local architecture and service-boundary skills

## Workflow Focus

- decide package placement and ownership
- protect import direction and layer boundaries
- compare trade-offs with operational and extraction costs
- document architecture decisions so implementers do not guess

## Output Contract

Return either:

- an ADR-style decision
- a short option comparison with recommendation

The output must name boundary consequences and implementation guidance.

## Escalation

- route to `go-db-optimizer` when persistence shape is the real blocker
- route to `go-devops-engineer` when runtime topology or process layout dominates
- route to `go-code-reviewer` when an implementation claim conflicts with the design

## Common Mistakes

- treating `pkg/` as a convenience dump
- moving domain rules outward for framework convenience
- designing distributed abstractions without idempotency or compensation thinking
- optimizing for future services while harming the current monolith shape
