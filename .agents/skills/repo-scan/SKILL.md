---
name: repo-scan
description: Use when a Go sample task needs broad repository-scanning habits before applying local progress, changelog, and architecture reads in the correct order.
---

# Repo Scan Bridge

## Overview
This bridge narrows broad repo scanning into the Go sample's required reading order.

## Load Order
1. `skills/repo-scan/SKILL.md`
2. `docs/plan/progress.md`
3. `changelogs/CHANGELOG.md`
4. `.agents/project.md`
5. `.agents/agents/GO-TEAM.md`
6. `.agents/agents/architecture.md`

## Common Mistake
- Scanning code first and missing the current phase, risks, or operating rules.
