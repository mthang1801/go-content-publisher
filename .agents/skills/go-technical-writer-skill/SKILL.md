---
name: go-technical-writer-skill
description: Use when updating reusable Go service documentation, progress logs, changelogs, specs, or onboarding guidance and you want one entry skill for the documentation and change-tracking stack.
---

# Go Technical Writer Skill

## Overview

This is the primary entry skill for `go-technical-writer`.

It loads broad documentation guidance first, then narrows into repository-local tracking, governance, and security-aware writing constraints.

## When To Use

- README or onboarding guidance changes
- progress or changelog entries must be updated
- a spec, proposal, tasks file, or implementation summary needs polishing
- architecture guidance or role descriptions drift from current reality

## Skill Stack

### Copied Base Skills

- `skills/doc-writer/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/doc-writer/SKILL.md`

### Precise Local Skills

- `.agents/skills/jira/SKILL.md`
- `.agents/skills/git-workflow/SKILL.md`
- `.agents/skills/security-review/SKILL.md`

## Loading Order

1. load the broad documentation base skill
2. load the bridge skill that translates it into this repository
3. load precise local governance and tracking skills

## Workflow Focus

- identify the audience and truth source first
- update the smallest correct doc set
- keep current-state docs accurate and durable-history docs additive
- avoid implying code or runtime that does not exist

## Output Contract

Produce a documentation sync summary that states:

- audience
- files updated
- why they changed
- remaining gaps or follow-up docs needed

## Escalation

- route to `go-architect` when docs need an actual architecture decision
- route to `go-orchestrator` when doc drift exposes wider planning drift
- route to `go-code-reviewer` when documentation may hide a verification gap

## Common Mistakes

- rewriting durable history instead of extending it
- documenting intended future state as if it already exists
- updating README while leaving progress or changelog stale
- adding filler text that does not help future contributors or agents
