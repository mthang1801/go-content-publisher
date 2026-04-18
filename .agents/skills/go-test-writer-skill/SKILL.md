---
name: go-test-writer-skill
description: Use when designing or writing Go service tests, regression coverage, or verification commands and you want one entry skill for the testing and TDD stack.
---

# Go Test Writer Skill

## Overview

This is the primary entry skill for `go-test-writer`.

It loads broad Go testing judgment first, then narrows into repository-local TDD, architecture-aware seam selection, persistence coverage, and failure-path verification.

## When To Use

- implementing new behavior with TDD
- fixing a bug that needs regression coverage
- deciding test layer or harness shape
- reviewing weak or missing verification
- choosing whether a seam needs real dependency-backed coverage

## Skill Stack

### Copied Base Skills

- `skills/go-senior-expert/SKILL.md`
- `skills/senior-dev/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/go-senior-expert/SKILL.md`
- `.agents/skills/senior-dev/SKILL.md`

### Precise Local Skills

- `.agents/skills/go-testing-tdd/SKILL.md`
- `.agents/skills/go-backend-patterns/SKILL.md`
- `.agents/skills/go-clean-architecture/SKILL.md`
- `.agents/skills/go-gorm-postgres/SKILL.md`
- `.agents/skills/go-error-handling/SKILL.md`
- `.agents/skills/go-debugging/SKILL.md`

## Loading Order

1. load broad Go and engineering-quality base skills
2. load bridge skills for repository-specific interpretation
3. load precise local testing and failure-path skills for the active seam

## Workflow Focus

- choose the narrowest useful test layer
- write RED cases first when applicable
- cover success and failure paths intentionally
- make verification commands concrete and reproducible

## Output Contract

Return a test plan or test changes that state:

- owning layer
- target cases
- fixtures or builders needed
- exact verification commands
- remaining execution gaps if any

## Escalation

- route to `go-backend-developer` once the seam and cases are clear
- route to `go-db-optimizer` when persistence-backed verification is the real risk
- route to `go-debugger` when intermittent failure or race evidence is needed first

## Common Mistakes

- starting from end-to-end for every change
- proving implementation details instead of behavior
- skipping dependency-failure paths
- claiming coverage without naming runnable commands
