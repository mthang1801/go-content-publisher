---
name: go-stream-pipeline
description: Use when designing event, queue, or stream processing pipelines in Go, especially for outbox consumers, Kafka, RabbitMQ, or background workers.
---

# Go Stream Pipeline

## Overview
Use this skill when messages or events become part of the delivery model. It keeps pipelines observable, retry-safe, and aligned with business ownership.

## When To Use
- Adding background consumers or workers.
- Planning Kafka or RabbitMQ adapters after phase 1.
- Designing retry, dead-letter, batching, or idempotency behavior.

## Core Pattern
- Producer side: domain event -> outbox -> publisher adapter.
- Consumer side: message decode -> contract validation -> idempotency check -> application handler -> ack/nack decision.

## Rules
- Message handlers should call application services, not reimplement business logic.
- Keep transport envelopes separate from domain events.
- Include correlation IDs, causation IDs, version, and event time.
- Define retry policy and poison-message handling up front.
- Record processing outcomes with enough metadata to debug duplicates or stalls.

## Common Mistakes
- Coupling consumer logic directly to broker clients.
- No idempotency or deduplication plan.
- Infinite retries with no dead-letter strategy.
