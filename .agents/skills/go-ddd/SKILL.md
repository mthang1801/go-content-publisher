---
name: go-ddd
description: Use when modeling bounded contexts, aggregates, value objects, domain services, or domain events in Go services that follow clean architecture.
---

# Go DDD

## Overview
Use this skill to model the business first, then attach transport and persistence concerns later. It keeps ubiquitous language visible in package names, types, and use cases.

## When To Use
- Starting a new bounded context or subdomain.
- Defining aggregates, entities, and value objects.
- Splitting logic between application orchestration and domain behavior.
- Reviewing whether a repository or service boundary matches the business model.

## Tactical Guidance
- One aggregate root owns consistency inside its boundary.
- Value objects should be immutable by convention and validate themselves on creation.
- Domain services exist only when behavior does not belong naturally on one entity or value object.
- Repositories load and persist aggregates or projections, not arbitrary persistence details.
- Domain events describe facts that already happened; they are not commands.

## Go-Specific Shape
- Use constructors such as `NewOrder(...)` and methods that enforce invariants.
- Keep unexported fields where possible; expose behavior, not setters.
- Prefer small packages per bounded context over one giant `domain` bucket.
- Use explicit typed IDs and value objects when they prevent cross-context confusion.

## Boundary Test
- If a rule depends on HTTP, DB schema, or queue behavior, it is not domain logic.
- If a rule expresses a business invariant, it belongs in the domain even when infrastructure also validates it.

## Common Mistakes
- Treating CRUD tables as aggregates.
- Creating repositories per table instead of per aggregate or capability.
- Putting workflow orchestration, retries, and external calls inside aggregates.
