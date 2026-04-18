---
name: go-debugging
description: Use when a Go service has panics, deadlocks, goroutine leaks, race conditions, timeout failures, or unexpected behavior across HTTP, database, cache, or worker flows.
---

# Go Debugging

## Overview
Use this skill for runtime diagnosis. It emphasizes reproducing the failure, narrowing the layer, and collecting evidence before changing code.

## When To Use
- Requests hang, time out, or return inconsistent results.
- Panics or race conditions appear in tests or production logs.
- Background workers leak goroutines or process duplicate work.

## Debug Flow
1. Reproduce with the smallest stable command or test.
2. Classify the failure: panic, race, deadlock, timeout, data corruption, bad mapping.
3. Check the owning layer first: presentation, application, domain, infrastructure, or `pkg/`.
4. Add temporary logs, traces, or focused tests to prove the hypothesis.
5. Fix the root cause, then remove or reduce temporary instrumentation.

## Go-Specific Tools
- `go test -run ...`
- `go test -race ./...`
- `pprof` for CPU, heap, goroutine, blocking analysis
- structured request logs with request IDs and correlation IDs
- containerized dependency checks for Postgres and Redis

## Common Mistakes
- Fixing symptoms before reproducing.
- Assuming concurrency bugs are database bugs.
- Leaving broad debug logs or sleeps in the code after the fix.
