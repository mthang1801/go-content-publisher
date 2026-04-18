---
name: go-code-reviewer-skill
description: Use when reviewing Go service changes for correctness, regressions, boundary leaks, persistence risk, operational safety, and verification quality and you want one entry skill for the review stack.
---

# Go Code Reviewer Skill

## Overview

This is the primary entry skill for `go-code-reviewer`.

It loads broad senior engineering judgment first, then narrows into repository-local architecture, persistence, runtime, and verification standards so findings stay severity-first and concrete.

## When To Use

- implementation is claimed to be complete
- the user asks for a review
- a risky refactor or bugfix needs an independent pass
- multiple specialists touched adjacent boundaries
- code or docs changed across more than one layer or runtime surface

## Skill Stack

### Copied Base Skills

- `skills/go-senior-expert/SKILL.md`
- `skills/backend-patterns-skill/SKILL.md`
- `skills/senior-dev/SKILL.md`
- `skills/senior-sysdesign/SKILL.md`
- `skills/senior-dba/SKILL.md`

### Repo-Local Bridge Skills

- `.agents/skills/go-senior-expert/SKILL.md`
- `.agents/skills/backend-patterns-skill/SKILL.md`
- `.agents/skills/senior-dev/SKILL.md`
- `.agents/skills/senior-sysdesign/SKILL.md`
- `.agents/skills/senior-dba/SKILL.md`

### Precise Local Skills

- `.agents/skills/go-backend-patterns/SKILL.md`
- `.agents/skills/go-clean-architecture/SKILL.md`
- `.agents/skills/go-ddd/SKILL.md`
- `.agents/skills/go-gorm-postgres/SKILL.md`
- `.agents/skills/go-testing-tdd/SKILL.md`
- `.agents/skills/go-error-handling/SKILL.md`
- `.agents/skills/go-redis/SKILL.md`
- `.agents/skills/go-microservices/SKILL.md`
- `.agents/skills/go-saga/SKILL.md`
- `.agents/skills/go-stream-pipeline/SKILL.md`
- `.agents/skills/security-review/SKILL.md`
- `.agents/skills/git-workflow/SKILL.md`

## Loading Order

1. load broad engineering, architecture, and DBA base skills
2. load bridge skills to align them with this repository
3. load precise local review-enforcement skills for changed surfaces

## Workflow Focus

- review correctness before style
- review boundaries before convenience
- review persistence and runtime risk before minor maintainability notes
- review verification credibility before approving anything

## Output Contract

Produce severity-ranked findings with:

- title
- file and line
- why it matters
- smallest safe fix
- explicit verdict

## Escalation

- route to `go-architect` if findings expose design disputes
- route to `go-db-optimizer` if data-path risk dominates the review
- route to `go-devops-engineer` if operational changes need independent scrutiny

## Common Mistakes

- leading with summaries instead of findings
- approving behavior that lacks verification evidence
- calling out style when the real problem is boundary or runtime risk
- missing documentation drift on a reusable template repository
