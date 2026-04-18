---
name: go-db-optimizer-skill
description: Use when designing or reviewing Postgres and GORM behavior for Go services, including schema shape, query paths, transactions, indexes, and migration risk, and you want one entry skill for the persistence stack.
---

# Go DB Optimizer Skill

## Overview

This is the primary entry skill for `go-db-optimizer`.

It loads broad DBA and PostgreSQL judgment first, then narrows into repository-local persistence, architecture, verification, and security constraints.

## When To Use

- designing or reviewing GORM repositories
- adding or changing tables, indexes, or constraints
- investigating slow list, projection, or reporting queries
- designing transaction boundaries or lock behavior
- evaluating migration safety and rollout risk

## Skill Stack

### Copied Base Skills

- `skills/senior-dba/SKILL.md`
- `skills/supabase-postgres-best-practices/SKILL.md`
- `skills/go-senior-expert/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/senior-dba/SKILL.md`
- `.agents/skills/supabase-postgres-best-practices/SKILL.md`
- `.agents/skills/go-senior-expert/SKILL.md`

### Precise Local Skills

- `.agents/skills/go-gorm-postgres/SKILL.md`
- `.agents/skills/go-clean-architecture/SKILL.md`
- `.agents/skills/go-ddd/SKILL.md`
- `.agents/skills/go-testing-tdd/SKILL.md`
- `.agents/skills/security-review/SKILL.md`

## Loading Order

1. load broad DBA and Postgres base skills
2. load bridge skills that normalize them for this repository
3. load precise local persistence, architecture, and verification skills

## Workflow Focus

- understand the real access pattern
- evaluate schema, index, transaction, and query shape together
- separate read models from write models when needed
- make migration safety as explicit as performance

## Output Contract

Produce a persistence review or optimization note that states:

- workload
- risks
- recommended changes
- verification

## Escalation

- route to `go-backend-developer` when repository ownership or adapter shape is the next step
- route to `go-architect` when persistence decisions create boundary disputes
- route to `go-code-reviewer` when data-path risk blocks approval

## Common Mistakes

- optimizing a guessed workload
- hiding transaction ownership inside generic helpers
- indexing by habit rather than actual query shape
- improving speed while weakening consistency or migration safety
