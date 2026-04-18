---
name: go-devops-engineer-skill
description: Use when shaping Go service runtime, Docker, CI, migration execution, worker bootstrap, environment contracts, or release safety and you want one entry skill for the operational stack.
---

# Go DevOps Engineer Skill

## Overview

This is the primary entry skill for `go-devops-engineer`.

It loads broad runtime and delivery knowledge first, then narrows into repository-local topology, security, messaging, and governance constraints.

## When To Use

- local bootstrap or runtime scripts are added or changed
- Docker or Compose configuration changes
- CI or verification pipelines need updates
- migrations, workers, or multi-process startup need coordination
- environment validation, config sourcing, or observability defaults need work

## Skill Stack

### Copied Base Skills

- `skills/devops-senior-expert/SKILL.md`
- `skills/go-senior-expert/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/devops-senior-expert/SKILL.md`
- `.agents/skills/go-senior-expert/SKILL.md`

### Precise Local Skills

- `.agents/skills/git-workflow/SKILL.md`
- `.agents/skills/security-review/SKILL.md`
- `.agents/skills/go-microservices/SKILL.md`
- `.agents/skills/go-redis/SKILL.md`
- `.agents/skills/go-stream-pipeline/SKILL.md`

## Loading Order

1. load broad runtime and delivery base skills
2. load bridge skills that reinterpret them for this repository
3. load precise local runtime-topology, security, messaging, and governance skills

## Workflow Focus

- make environment contracts explicit
- keep local, CI, and release paths reproducible
- coordinate process topology across `api`, `worker`, `migrate`, and `cli`
- reveal missing infrastructure assumptions instead of hiding them

## Output Contract

Produce a runtime plan that states:

- scope
- environment contract
- changes
- verification
- operational risks

## Escalation

- route to `go-backend-developer` when runtime issues are really code-shape issues
- route to `go-architect` when process topology forces structural decisions
- route to `go-code-reviewer` for final operational risk validation

## Common Mistakes

- encoding hidden local assumptions in scripts
- making CI diverge from local verification without reason
- inventing production detail for infrastructure that does not exist yet
- treating startup and shutdown behavior as separate from correctness
