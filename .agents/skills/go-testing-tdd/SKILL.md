---
name: go-testing-tdd
description: Use when implementing or changing Go service behavior and you need test-first guidance across domain, application, transport, and infrastructure boundaries.
---

# Go Testing TDD

## Overview
Use this skill to drive changes with failing tests first and to keep verification aligned with architecture boundaries. Favor fast tests close to the logic, then add integration coverage where the seam matters.

## When To Use
- Any feature or bugfix changes behavior.
- A refactor risks breaking domain rules or handler contracts.
- The team needs a test split between unit, integration, and end-to-end levels.

## Test Pyramid
- Domain tests: invariants, state transitions, value objects.
- Application tests: use-case orchestration, transaction semantics, policy branching.
- Presentation tests: handler validation, status codes, response shapes.
- Infrastructure tests: repository mappings, SQL constraints, Redis/broker adapters.
- End-to-end tests: only for critical flows that prove wiring works.

## TDD Loop
1. Write a failing test at the narrowest useful layer.
2. Implement the minimum change to pass.
3. Refactor while keeping package boundaries intact.
4. Add regression coverage for discovered edge cases.

## Go-Specific Guidance
- Prefer table-driven tests when inputs map naturally to cases.
- Use real Postgres/Redis containers for adapter tests when mocks would hide risk.
- Mock ports, not concrete frameworks.
- Keep test data builders close to the bounded context they serve.

## Common Mistakes
- Starting from E2E tests for every change.
- Mocking GORM or Gin deeply instead of testing your own abstractions.
- Treating race conditions, cancellation, and timeout behavior as optional.
