---
name: search-first
description: Use when a Go sample task needs broad search-first habits before applying the local docs, change records, and code-surface search order.
---

# Search First Bridge

## Overview
This bridge narrows broad search-first behavior into the Go sample's local discovery order.

## Load Order
1. `skills/search-first/SKILL.md`
2. `docs/plan/progress.md`
3. `changelogs/CHANGELOG.md`
4. `changelogs/changes/`
5. relevant `.agents/skills/` and code packages

## Common Mistake
- Searching only code and skipping change records or operating docs that already explain the surface.
