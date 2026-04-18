# Architecture Decision Records & Review Framework

---

## 1. ADR Format — Battle-Tested Template

ADR là "git commit message cho architecture decisions" — immutable history,
không sửa decision cũ mà tạo ADR mới supersede nó.

```markdown
# ADR-{NUMBER}: {Tiêu đề ngắn gọn, action-oriented}

**Date**: YYYY-MM-DD  
**Status**: Proposed | Accepted | Deprecated | Superseded by ADR-{N}  
**Deciders**: {Tên hoặc team}  
**Reviewed by**: {Tên senior/principal}

---

## Context

{2-4 đoạn. Mô tả bối cảnh dẫn đến quyết định này.
Tập trung vào: vấn đề gì, tại sao cần quyết định ngay bây giờ.
Không mention solution ở đây — chỉ problem space.}

## Decision Drivers

- {Force 1: constraint, quality attribute, hoặc concern quan trọng nhất}
- {Force 2}
- {Force 3}

## Options Considered

### Option A: {Tên ngắn gọn}

{Mô tả kỹ thuật — đủ để engineer hiểu mà không cần research}

**Pros:**
- {Lợi điểm cụ thể, có thể đo được}

**Cons:**
- {Nhược điểm cụ thể, không né tránh}

**Risks:**
- {Risk ẩn — thứ có thể không hiển nhiên ngay}

### Option B: {Tên ngắn gọn}

{...}

### Option C: Status Quo (do nothing)

{Luôn xem xét option "không làm gì". Tại sao current state không đủ?}

## Decision

**Chọn: Option {X}**

{Lý do chọn — không chỉ "pros > cons" mà phải explain trade-off nào được
accept và tại sao, trong context cụ thể của team/product/timeline này.}

## Consequences

**Positive:**
- {Benefit cụ thể từ decision này}

**Negative (Accepted Trade-offs):**
- {Nhược điểm được chấp nhận, và tại sao}

**Risks & Mitigations:**
- Risk: {X} → Mitigation: {Y}

## Validation

{Làm sao biết decision này đúng? Metric nào sẽ confirm hoặc deny?
Timeline để review lại nếu metric không đạt?}
```

---

## 2. RFC Format — Đề xuất thay đổi lớn

RFC (Request for Comments) dùng cho changes ảnh hưởng nhiều team
hoặc thay đổi architecture cơ bản.

```markdown
# RFC-{NUMBER}: {Tiêu đề}

**Author**: {Tên}  
**Date**: YYYY-MM-DD  
**Status**: Draft | Under Review | Accepted | Rejected  
**Target teams**: {Các team bị ảnh hưởng}  
**Review deadline**: YYYY-MM-DD  

---

## Summary

{1 đoạn: vấn đề + giải pháp đề xuất. Đủ để manager hiểu trong 30 giây.}

## Problem Statement

### Current State

{Mô tả hiện tại — metrics cụ thể nếu có.
Ví dụ: "P99 latency hiện tại là 850ms, SLO là 500ms, breach 3 tuần/tháng"}

### Why Now

{Tại sao cần solve bây giờ. Cost of inaction là gì?}

### Non-Goals

{Explicit: bài này KHÔNG solve những vấn đề nào. Giúp scope discussion.}

## Proposed Solution

### High-Level Design

{Diagram hoặc ASCII art. Architecture mới trông như thế nào.}

### Detailed Design

{Technical details cho engineers implement.
Cụ thể đủ để estimate effort, không cần speculative.}

### Migration Strategy

{Từ current state đến proposed state.
Big bang hay incremental? Rollback plan?}

## Impact Analysis

### Performance

{Expected improvement với data backing — load test result hoặc estimation}

### Reliability

{Failure modes mới? Cái gì cải thiện, cái gì trade-off?}

### Security

{New attack surface? Security review needed?}

### Operational

{Monitoring changes, runbook changes, on-call impact}

### Cost

{Infrastructure cost delta. Engineer time estimate.}

## Alternatives Considered

{2-3 alternatives với lý do không chọn. Shows due diligence.}

## Open Questions

{Những quyết định chưa được đưa ra. Cần input từ ai?}

## Success Metrics

{Làm sao measure thành công. Timeline review.}

## Rollout Plan

{Phase 1: ... | Phase 2: ... | Rollback trigger: ...}
```

---

## 3. Architecture Review Checklist

Dùng khi review design của người khác — structured, không skip bước.

### Functional Correctness

```
□ Có đáp ứng đủ functional requirements không?
□ Edge cases được handle chưa? (empty state, error state, race condition)
□ Data flow có consistent không (input → transform → output)?
□ API contracts rõ ràng? (request/response schema, error codes)
```

### Scalability

```
□ Bottleneck ở đâu khi 10x current load?
□ Stateless components có thể horizontal scale không?
□ Database layer: read replicas? sharding strategy?
□ Caching được dùng đúng layer chưa? (CDN, API layer, application, DB)
□ Async processing cho non-critical paths?
```

### Reliability & Resilience

```
□ Single points of failure được identify chưa?
□ Circuit breakers trên external dependencies?
□ Retry strategy với exponential backoff và jitter?
□ Graceful degradation khi dependency down?
□ Health check endpoints? Readiness vs Liveness?
□ Data backup và recovery strategy?
□ RTO và RPO đã được define chưa?
```

### Security

```
□ Authentication: OAuth2, JWT, API Key — ai authenticate ai?
□ Authorization: RBAC, ABAC — ai được phép làm gì?
□ Data in transit: TLS 1.2+ everywhere?
□ Data at rest: encryption cho PII, credentials?
□ Input validation: injection prevention (SQL, XSS, SSRF)?
□ Rate limiting để prevent abuse?
□ Secrets management: không hard-code credentials?
□ Least privilege principle cho service accounts?
```

### Observability

```
□ Structured logging (JSON, với trace_id, user_id)?
□ Metrics: RED (Rate, Errors, Duration) cho mỗi service?
□ Distributed tracing (OpenTelemetry)?
□ Alerting: SLO-based alerts, không raw metric alerts?
□ Dashboards: per-service, end-to-end, business metrics?
□ Runbooks cho common failure scenarios?
```

### Operational Excellence

```
□ Deployment strategy: blue-green, canary, rolling?
□ Feature flags cho gradual rollout?
□ Database migrations: backward compatible? Zero-downtime?
□ Configuration: env-based, không hard-coded?
□ Cost: estimated infra cost? Cost alerting?
□ Documentation: README, API docs, runbooks?
□ On-call: who owns this? What's the escalation path?
```

---

## 4. Trade-off Language — Nói chuyện như Principal

Tránh: "This is better" → Dùng: "This optimizes for X at the cost of Y"

### Trade-off Matrix Template

```
|              | Option A         | Option B         | Option C         |
|--------------|------------------|------------------|------------------|
| Consistency  | Strong           | Eventual         | Causal           |
| Latency      | P99: 50ms        | P99: 5ms         | P99: 20ms        |
| Throughput   | 10K WPS          | 100K WPS         | 50K WPS          |
| Complexity   | Low              | High             | Medium           |
| Cost         | $500/mo          | $2000/mo         | $800/mo          |
| Ops burden   | Low              | High             | Medium           |
| Scale limit  | 100K users       | 100M users       | 10M users        |
```

### Ngôn ngữ trade-off chuẩn

```
"Option A optimizes for operational simplicity at the cost of write throughput.
 Given our team of 5 engineers and current scale of 50K users,
 the simpler system is the right choice for the next 18 months."

"Option B would give us 10x write throughput, but introduces
 Kafka operational complexity and eventual consistency semantics.
 We should revisit this when we hit 500K users or if write latency
 becomes a customer complaint."

"The real risk isn't choosing A over B — it's not having a clear
 migration path to B when the time comes. Let's design A
 with explicit seams where B would plug in."
```

---

## 5. Estimation Templates

### Capacity Planning Template

```markdown
## Capacity Planning: {System Name}

### Assumptions
- DAU: {X} million
- Average sessions/day: {N}
- Average requests/session: {M}
- Peak multiplier: {3-10x} over average

### Traffic Estimation
- Average QPS: DAU × requests/session / 86400 = {X} QPS
- Peak QPS: {X × multiplier} QPS
- Read:Write ratio: {10:1 / 100:1 / ...}

### Storage Estimation
- Data per user: {size calculation}
- Total data (year 1): {users × data/user × 365}
- Growth rate: {X% month over month}
- Storage at year 3: {calculation}

### Infrastructure Estimation (rough)
- App servers: Peak QPS / {500 QPS/server} = {N} servers
- Cache: {working set size} GB
- DB: {storage + 20% overhead + IOPS calculation}
- Bandwidth: Peak QPS × {avg response size}

### Cost Estimate (Cloud)
- Compute: ${N} servers × ${price}/mo = ${total}/mo
- Storage: ${GB} × ${price/GB}/mo = ${total}/mo
- Network egress: ${GB} × ${price/GB} = ${total}/mo
- Total: ~${sum}/month
```

---

## 6. Tech Radar — Adopt/Trial/Assess/Hold

Dùng để categorize technology recommendations:

```
ADOPT: Proven, battle-tested, recommend for new projects
  - PostgreSQL (relational), Redis (cache/queue), Kafka (streaming)
  - Kubernetes (orchestration), Terraform (IaC)
  - React/Next.js (frontend), Go/Java/Node.js (backend)

TRIAL: Promising, use cautiously with learning budget
  - ClickHouse (analytics), CockroachDB (distributed SQL)
  - Temporal (workflow orchestration), Vector DBs for RAG

ASSESS: Investigate, not yet ready for production
  - New LLM integrations, experimental cloud services

HOLD: Avoid for new projects, migrate existing
  - Monolithic session management, XML-based APIs
  - Synchronous coupling across service boundaries
  - In-process messaging (use queue instead)
```
