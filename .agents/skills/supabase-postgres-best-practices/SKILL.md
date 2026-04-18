---
name: supabase-postgres-best-practices
description: Use when copied local PostgreSQL best-practice guidance is relevant but must be reinterpreted through the Go sample's Postgres and GORM rules.
---

# Postgres Best Practices Bridge

## Overview
This bridge narrows copied local PostgreSQL guidance into the Go sample's local persistence conventions.

## Load Order
1. `skills/supabase-postgres-best-practices/SKILL.md`
2. `.agents/skills/go-gorm-postgres/SKILL.md`
3. `.agents/skills/security-review/SKILL.md`

## Common Mistake
- Porting Postgres advice mechanically without checking transaction ownership or repository boundaries.
