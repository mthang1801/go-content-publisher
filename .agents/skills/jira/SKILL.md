---
name: jira
description: Use when a Go service change needs traceable ticket context, acceptance criteria, implementation notes, or status updates that map engineering work back to planning artifacts.
---

# Jira

## Overview
Use this skill to turn ticket intent into implementation-ready work without letting the ticket become the only source of truth.

## When To Use
- A task starts from Jira or another issue tracker.
- Acceptance criteria need translating into specs, tasks, or tests.
- Progress or risk updates need to be written back clearly.

## Working Pattern
- Extract business goal, constraints, and acceptance criteria.
- Mirror the implementation plan in local change docs when the work is substantial.
- Map ticket terms to bounded contexts, packages, and test scenarios.
- Record deviations or discovered scope changes explicitly.

## Common Mistakes
- Treating vague ticket text as an implementation spec.
- Updating Jira without updating repo-local progress and change records.
- Losing links between acceptance criteria and verification evidence.
