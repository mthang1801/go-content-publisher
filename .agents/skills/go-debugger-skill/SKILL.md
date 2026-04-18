---
name: go-debugger-skill
description: Use when diagnosing Go service failures such as panics, startup problems, timeouts, cache bugs, race conditions, or broken worker flows and you want one entry skill for the debugging stack.
---

# Go Debugger Skill

## Overview

This is the primary entry skill for `go-debugger`.

It loads broad Go, backend, database, and runtime-ops judgment first, then narrows into repository-local root-cause, error-handling, persistence, cache, and regression-protection skills.

## When To Use

- the service crashes or refuses to start
- requests hang, time out, or behave inconsistently
- race or goroutine-lifecycle symptoms appear
- cache, worker, or background processing behaves incorrectly
- startup, env, or script behavior diverges from expectations

## Skill Stack

### Copied Base Skills

- `skills/go-senior-expert/SKILL.md`
- `skills/backend-patterns-skill/SKILL.md`
- `skills/senior-dev/SKILL.md`
- `skills/senior-dba/SKILL.md`
- `skills/devops-senior-expert/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/go-senior-expert/SKILL.md`
- `.agents/skills/backend-patterns-skill/SKILL.md`
- `.agents/skills/senior-dev/SKILL.md`
- `.agents/skills/senior-dba/SKILL.md`
- `.agents/skills/devops-senior-expert/SKILL.md`

### Precise Local Skills

- `.agents/skills/go-debugging/SKILL.md`
- `.agents/skills/go-error-handling/SKILL.md`
- `.agents/skills/go-backend-patterns/SKILL.md`
- `.agents/skills/go-gorm-postgres/SKILL.md`
- `.agents/skills/go-redis/SKILL.md`
- `.agents/skills/go-stream-pipeline/SKILL.md`
- `.agents/skills/go-microservices/SKILL.md`
- `.agents/skills/go-testing-tdd/SKILL.md`

## Loading Order

1. load broad Go, backend, DBA, and runtime base skills
2. load bridge skills to align them with this repository
3. load precise local debugging and regression-protection skills

## Workflow Focus

- stabilize reproduction
- classify the failure by seam and ownership
- prove root cause before fixing
- add regression protection before closing the issue

## Output Contract

Produce a debug report that states:

- symptom
- reproduction
- root cause
- fix
- regression protection
- verification

## Escalation

- route to `go-db-optimizer` when query, transaction, or lock behavior is the real cause
- route to `go-devops-engineer` when startup or runtime-contract issues dominate
- route to `go-test-writer` when regression protection needs a separate plan

## Common Mistakes

- fixing the symptom before reproduction is reliable
- hiding logic bugs behind retries or sleeps
- blaming the wrong layer without evidence
- closing an issue without fresh verification
