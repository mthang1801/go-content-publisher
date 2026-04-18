---
name: security-review
description: Use when reviewing Go service changes for authentication, authorization, data exposure, input handling, secrets, transport security, or unsafe operational defaults.
---

# Security Review

## Overview
Use this skill to review changes for common service-level security risks. It is especially important at boundaries: HTTP input, persistence, cache, messaging, config, and operational scripts.

## When To Use
- New endpoints, middleware, or auth-sensitive flows are added.
- Persistence or cache code handles customer or order data.
- Operational scripts, config, or integrations change.

## Review Areas
- Authentication and authorization checks are explicit and testable.
- Input validation and output redaction happen at the boundary.
- Secrets come from config providers, not hardcoded values or logs.
- SQL queries, GORM usage, and raw expressions are parameterized safely.
- Redis, Kafka, RabbitMQ, and HTTP clients have sane timeouts and TLS expectations.

## Common Mistakes
- Logging raw tokens, full payloads, or internal stack traces to clients.
- Assuming internal services are trusted by default.
- Adding convenience admin paths without auditability or access control.
