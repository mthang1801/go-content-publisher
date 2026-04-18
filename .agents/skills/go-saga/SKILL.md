---
name: go-saga
description: Use when a Go system needs long-running business workflows, compensating actions, or coordination across multiple services or bounded contexts.
---

# Go Saga

## Overview
Use this skill for cross-service workflows that cannot be handled in one transaction. It focuses on explicit state, compensations, and observability.

## When To Use
- A workflow spans multiple services or data stores.
- Failure recovery needs compensating actions rather than rollback.
- The team is deciding between orchestration and choreography.

## Guidance
- Start with explicit orchestration when the workflow is business-critical or hard to reason about.
- Define forward steps, compensation steps, and terminal failure behavior.
- Persist saga state and correlation IDs.
- Make every step idempotent.
- Distinguish business compensation from technical retries.

## Not For
- Simple local transactions.
- Cases where one aggregate can enforce consistency alone.

## Common Mistakes
- Hiding saga state in transient logs only.
- Mixing compensation logic with HTTP handlers.
- Treating retries as compensation.
