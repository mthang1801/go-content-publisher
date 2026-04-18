# .agents — Go Service Agent Instructions

This file is the primary entrypoint for AI agents working inside this repository.

## Universal Safety Rules

- No auto-commit or auto-push unless the user explicitly asks.
- Do not execute destructive filesystem commands without explicit user approval.
- If a task maps to a skill, read the skill first.
- If a task is non-trivial, follow the planning and change-tracking workflow instead of jumping straight to code.

## Required Reading Order

Before planning or implementing meaningful work, read in this order:

1. `docs/plan/progress.md`
2. `changelogs/CHANGELOG.md`
3. `.agents/project.md`
4. `.agents/agents/GO-TEAM.md`
5. `.agents/agents/interaction-protocol.md`
6. `.agents/agents/architecture.md`
7. relevant files under `.agents/workflows/`, `.agents/skills/`, and `.agents/rules/`

## Core Workflow

### 1. Understand State First

Read project progress and recent changes before deciding what to do next.

### 2. Use Spec-First Changes

For non-trivial work, create:

```text
changelogs/changes/<slug>/
├── proposal.md
├── tasks.md
├── design.md        # optional
└── evidence/        # optional
```

### 3. Use Skills Before Coding

If work touches architecture, Go persistence, debugging, testing, Redis, or microservices, load the corresponding skill first.

### 4. Follow Workflows

Use:

- `.agents/workflows/orchestration.md`
- `.agents/workflows/new-feature.md`
- `.agents/workflows/debugging.md`
- `.agents/workflows/code-review.md`

### 5. Keep Documentation In Sync

After meaningful work:

- update `docs/plan/progress.md`
- update `changelogs/CHANGELOG.md`
- update the relevant `changelogs/changes/<slug>/...` artifacts

## Go-Specific Expectations

- Prefer explicit constructor injection.
- Keep handlers thin and domain framework-free.
- Use `context.Context` correctly.
- Respect package boundaries and import direction.
- Treat `pkg/` as the reusable platform layer inside this sample.

## Agent System

Specialist roles are defined under `.agents/agents/`.

Start with:

- `.agents/agents/GO-TEAM.md`
- `.agents/agents/interaction-protocol.md`

The generic starting team is:

- `go-orchestrator`
- `go-architect`
- `go-backend-developer`
- `go-code-reviewer`
- `go-debugger`
- `go-test-writer`
- `go-db-optimizer`
- `go-devops-engineer`
- `go-technical-writer`

## Skill System

There are two skill roots:

1. `skills/`
   Copied local base skills bundled with this sample.
2. `.agents/skills/`
   Repo-local role bundle skills plus bridge and precision skills for this Go sample.

When work maps to an agent role, load that role's bundle skill from `.agents/skills/` first.
Then load copied base skills from `skills/`, then bridge and precise local skills from `.agents/skills/`.
Do not fall back to any external root skill tree for this repository.

## Source Of Truth

`.agents/` is the only source of truth in this repository. There is no secondary mirrored agent tree.
