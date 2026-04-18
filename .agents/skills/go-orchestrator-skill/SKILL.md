---
name: go-orchestrator-skill
description: Use when routing ambiguous, multi-step, or cross-surface Go service work and you need one entry skill that loads planning, repo-discovery, system-design, and tracking guidance in the correct order.
---

# Go Orchestrator Skill

## Overview

This is the primary entry skill for `go-orchestrator`.

Use it to load the smallest correct planning stack before emitting routing, sequencing, verification, and documentation requirements for a Go service task.

## When To Use

- the request is ambiguous or spans multiple steps
- more than one specialist may be needed
- the blast radius is unclear
- the task touches both code and operating-model concerns
- change tracking, verification, and routing all need to be explicit

## Skill Stack

### Copied Base Skills

- `skills/strategic-compact/SKILL.md`
- `skills/repo-scan/SKILL.md`
- `skills/search-first/SKILL.md`
- `skills/senior-sysdesign/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/strategic-compact/SKILL.md`
- `.agents/skills/repo-scan/SKILL.md`
- `.agents/skills/search-first/SKILL.md`
- `.agents/skills/senior-sysdesign/SKILL.md`

### Precise Local Skills

- `.agents/skills/jira/SKILL.md`
- `.agents/skills/git-workflow/SKILL.md`

## Loading Order

1. load repo-discovery and planning base skills
2. load the matching bridge skills to translate them into this repository
3. load tracking and workflow precision skills needed for the actual plan

## Workflow Focus

- bootstrap repo context
- read progress, changelog, architecture, and team protocol
- classify task type, runtime surfaces, and blast radius
- choose the smallest correct specialist chain
- emit a verification-aware plan with required doc updates

## Output Contract

Return one concise execution brief that states:

- task summary
- task type
- current state
- specialist chain
- required reads
- risks
- verification
- doc and change-record updates

## Escalation

- route to `go-architect` for boundary or design disputes
- route to `go-code-reviewer` for acceptance or severity disputes
- escalate to the user when scope or contract changes materially

## Common Mistakes

- routing implementation work without reading current repo state
- assigning too many specialists when one primary owner is enough
- skipping change-record or doc updates for meaningful work
- pretending a runtime surface already exists when it does not
