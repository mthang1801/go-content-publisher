# Distributed Systems & Operations in Go

Derived from:
- `documents/assets/go/microservices/README.md`
- `documents/assets/go/messaging/README.md`
- `documents/assets/go/observability/README.md`
- `documents/assets/go/cloud-infra/README.md`
- `documents/assets/go/deployment/README.md`

## When To Read

Load this reference when the problem crosses process boundaries: remote calls, brokers, tracing, probes, rollout behavior, container/runtime setup, or release automation.

## Microservices

- Remote calls are not local method calls. Every downstream call needs timeout, cancellation, retry, and failure-isolation thinking.
- Prefer gRPC/protobuf for strict internal contracts and streaming; use REST or BFF layers intentionally at public edges.
- If consistency spans services, think outbox and saga before inventing distributed ACID.

## Messaging

- Pick the broker for its delivery semantics, not brand familiarity: Kafka for replay/partitioned streams, RabbitMQ/NATS for queue/request patterns.
- Consumer correctness means idempotency, bounded retries, dead-letter handling, and explicit offset/ack strategy.
- Queue lag and poison messages are product incidents, not just infrastructure noise.

## Observability

- Structured `slog` logs, RED metrics, distributed traces, and internal-only `pprof` endpoints are the default production baseline.
- Every telemetry signal should answer an investigation question: what failed, where, and how many users are affected?
- If you cannot explain the next debugging step from logs/metrics/traces, the instrumentation is incomplete.

## Cloud Runtime & Deployment

- Readiness, liveness, and startup probes encode different truths; do not collapse them into one endpoint.
- Graceful shutdown means request draining, worker stop, context cancellation, and broker/queue coordination.
- Scaling needs the right signal: HTTP saturation, queue lag, or worker concurrency limits.
- Release pipelines should use multi-stage builds, minimal runtime images, CI quality gates, provenance/build metadata, and a fast rollback path.

## Symptom Router

| Symptom | Open lane |
| --- | --- |
| cascading timeout, downstream collapse, missing cross-service contract | `microservices/` |
| duplicate side effect, poison message, queue lag, replay questions | `messaging/` |
| "we have dashboards but still can't explain the failure" | `observability/` |
| pods restart cleanly but traffic still drops on deploy | `cloud-infra/` |
| build/release image/rollout pipeline confusion | `deployment/` |
