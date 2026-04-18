# Technical Writing Guide — DevOps Blogs, Runbooks & Docs

## Table of Contents
1. [Blog Post Structure](#blog)
2. [Runbook Template](#runbook)
3. [Postmortem Template](#postmortem)
4. [Architecture Decision Record (ADR)](#adr)
5. [Writing Voice & Style](#voice)

---

## 1. Blog Post Structure

### Title Formula
```
[Strong Verb] + [Specific Outcome] + [Optional: qualifier]

GOOD: "Zero-Downtime Kubernetes Deployments: A Battle-Tested Approach"
GOOD: "How We Reduced Cloud Costs 40% Without Touching Application Code"
GOOD: "Debugging Goroutine Leaks in Production: What Worked and What Didn't"

BAD: "Introduction to Kubernetes" (too generic)
BAD: "Kubernetes Best Practices" (no specific outcome)
BAD: "Amazing New Features in Terraform 1.7" (vendor-speak)
```

### Full Blog Template

```markdown
# [Title: Verb + Specific Outcome]

> **TL;DR**: [3 sentences. What the problem is, what you did, what the result was.
> Skimmers should get full value from this paragraph.]

---

## The Problem

[Start with pain. Real scenario, real numbers if possible.
"We were processing 50k orders/day when suddenly..." not "Kubernetes is important because..."]

[Why existing solutions failed or didn't apply.]

---

## Why the Obvious Answer Doesn't Work

[Show the naive approach — and where it breaks.
This builds credibility: you've tried it. You're not just theorizing.]

```bash
# This seems right...
kubectl apply -f deployment.yaml

# But under load, you'll see this:
# Error: too many open connections (max: 100, got: 143)
```

[Explanation of WHY it breaks.]

---

## The Solution

[Step-by-step. Every step has a command or config snippet.
Assume the reader is smart but unfamiliar with your specific setup.]

### Step 1: [Concrete action title]

[One paragraph explaining what this step does and why, THEN the code.]

```yaml
# Comment explaining the non-obvious parts
apiVersion: apps/v1
kind: Deployment
# ...
```

> **WARNING**: [Callout for anything that can destroy data or cause downtime]

> **TIP**: [Callout for optimization or time-saving shortcut]

### Step 2: [Next action]
...

---

## Verifying It Works

[How to confirm the solution is working. Real commands with expected output.]

```bash
$ kubectl get pods -n production -l app=order-service
NAME                            READY   STATUS    RESTARTS   AGE
order-service-7d6f9b8c4-x9z2k   1/1     Running   0          2m
order-service-7d6f9b8c4-m3p1l   1/1     Running   0          2m
order-service-7d6f9b8c4-k7w9n   1/1     Running   0          2m
```

---

## War Stories: What Bit Us in Production

[This section is what separates expert blogs from tutorial blogs.
Real failure you encountered, how you found it, how you fixed it.
Anonymize if needed.]

**The incident**: [brief description]
**The symptom**: [what we saw]  
**The gotcha**: [what we missed initially]
**The fix**: [what actually solved it]

---

## What Could Go Wrong

[Honest, pre-emptive troubleshooting. This saves readers hours.]

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| Pods stuck in Terminating | PVC not released | `kubectl delete pvc --grace-period=0` |
| HPA not scaling | metrics-server not running | `kubectl top pods` to verify |
| ... | ... | ... |

---

## Further Reading

- [Official Kubernetes Docs — Rolling Updates](https://kubernetes.io/docs/...) — The spec, dry but accurate
- [Google SRE Book — Chapter on Release Engineering](https://sre.google/sre-book/) — Mental model for reliability
- [Argo Rollouts Docs](https://argoproj.github.io/rollouts/) — If you need canary/blue-green

---

*Found an error? [Open an issue](https://github.com/...) or ping us on Slack #platform-eng.*
```

---

## 2. Runbook Template

```markdown
# Runbook: [Service Name] — [Scenario Name]
**Last Updated**: YYYY-MM-DD | **Owner**: @team-name | **Severity**: P1/P2/P3

---

## Overview

**Service**: order-service  
**Alert**: `OrderServiceHighErrorRate`  
**SLO Impact**: Burning error budget at >10x rate  
**Expected Resolution Time**: 15-30 minutes

---

## Symptoms

- [ ] Error rate > 5% (check: [Grafana link])
- [ ] P99 latency > 2s
- [ ] User-facing: "Order placement failed" errors

---

## Immediate Actions (First 5 Minutes)

1. **Check if it's a deployment issue**
   ```bash
   kubectl rollout history deployment/order-service -n production
   # If recent deploy: consider rollback first
   ```

2. **Check pod health**
   ```bash
   kubectl get pods -n production -l app=order-service
   kubectl describe pod <crashing-pod> -n production | tail -30
   ```

3. **Check error logs**
   ```bash
   kubectl logs -n production -l app=order-service --since=10m | grep ERROR | tail -50
   ```

4. **If obvious cause found → jump to resolution. If not → continue diagnosis.**

---

## Diagnosis

### Hypothesis 1: Database Connection Pool Exhausted

**Signal**: Logs contain "connection pool exhausted" or DB latency spike on [Grafana link]

```bash
# Check DB connections
kubectl exec -it <pod> -n production -- \
  curl -s localhost:8080/metrics | grep db_connections
```

**Resolution**: [link to DB section below]

### Hypothesis 2: Memory Pressure / OOMKilled

**Signal**: `RESTARTS > 0` in pod list, exit code 137

```bash
kubectl get pod <pod> -n production -o json | \
  jq '.status.containerStatuses[0].lastState.terminated'
```

**Resolution**: [link to OOM section below]

### Hypothesis 3: Downstream Service Degraded

**Signal**: Traces show latency in payment-service calls [Jaeger link]

**Resolution**: Enable circuit breaker: `kubectl set env deployment/order-service PAYMENT_CIRCUIT_BREAKER=true -n production`

---

## Resolution Steps

### Rollback (fastest mitigation)
```bash
kubectl rollout undo deployment/order-service -n production
kubectl rollout status deployment/order-service -n production
```

### Scale Up (buy time while diagnosing)
```bash
kubectl scale deployment/order-service --replicas=10 -n production
```

### Emergency: Redirect Traffic to Backup Region
[Link to DR runbook]

---

## Verification

```bash
# Error rate should drop below 1%
watch -n 5 'kubectl exec -it <pod> -n production -- \
  curl -s localhost:8080/metrics | grep http_requests_total'
```

---

## Escalation

| Time Without Resolution | Escalate To |
|------------------------|-------------|
| 15 min | Team Lead on-call |
| 30 min | Engineering Manager |
| 1 hour | VP Engineering |

**Slack**: #incident-[INC-number]  
**Bridge**: [Zoom link]
```

---

## 3. Postmortem Template

```markdown
# Postmortem: [INC-1234] [Brief Description]

**Date**: YYYY-MM-DD  
**Duration**: Xh Ym  
**Severity**: P1/P2  
**Author**: @name  
**Reviewers**: @team-lead, @oncall-engineer  
**Status**: Draft / In Review / Approved

---

## Impact

| Metric | Value |
|--------|-------|
| Duration | 1h 42m |
| Users Affected | ~15,000 (15% of active users) |
| Revenue Impact | ~$45,000 (estimated) |
| SLO Budget Burned | 32% of monthly error budget |

---

## Summary

[2-3 sentence plain English summary. No jargon. A non-engineer should understand this.]

A missing database index caused full table scans on our orders table. As traffic increased
during peak hours, query latency spiked from 10ms to 8000ms, causing cascading timeouts
across the order placement flow.

---

## Timeline

All times UTC.

| Time  | Event |
|-------|-------|
| 14:00 | DB migration deployed to production (automatic, on deploy) |
| 14:23 | Alert fired: OrderServiceHighLatency p99 > 1s |
| 14:27 | On-call @name acknowledged alert |
| 14:35 | Identified spike in DB query times via distributed traces |
| 14:42 | Root cause found: missing index on `orders.customer_id` column |
| 14:55 | `CREATE INDEX CONCURRENTLY` executed (online, non-blocking) |
| 15:10 | Index build complete, queries returning to 10ms |
| 15:15 | Error rate dropped below 0.1% |
| 16:05 | Monitoring confirmed stable, incident closed |

---

## Root Cause

[Single, specific root cause. Not "human error". Not "insufficient testing".]

The DB migration for feature X added a new query pattern filtering by `customer_id`
without creating an index on that column. At low traffic in staging, the full table scan
completed in ~50ms (undetected). At production scale (8M rows, peak traffic), it took 6-8s.

---

## Contributing Factors

[Things that made this possible or worse. These drive action items.]

1. **Staging DB has 50k rows; production has 8M rows** — Performance difference invisible in staging
2. **Migration applies automatically on deploy** — No human review of query patterns
3. **No slow query alerting configured** — We detected via user impact, not proactively
4. **Missing query plan review in PR process** — No tooling to catch missing indexes

---

## What Went Well

[Genuine positives. This matters for morale and learning.]

- On-call response was fast (4 minutes to acknowledge)
- Distributed tracing (Jaeger) pinpointed DB latency in minutes, not hours
- `CREATE INDEX CONCURRENTLY` prevented additional downtime during fix
- Rollback runbook was available and accurate

---

## Action Items

| Item | Owner | Priority | Due Date |
|------|-------|----------|---------|
| Add EXPLAIN ANALYZE check to migration CI step | @devops | P1 | 2024-02-15 |
| Configure RDS slow query log + Cloudwatch alert (>500ms) | @dba | P1 | 2024-02-10 |
| Seed staging DB with production-scale anonymized data | @infra | P2 | 2024-02-28 |
| Add index review checklist to PR template | @dev-lead | P2 | 2024-02-20 |

---

## Lessons Learned

[What would you tell your past self?]

Query performance at production scale cannot be predicted from staging with 600x fewer rows.
Every migration that introduces a new query pattern needs an explicit index review,
regardless of how fast it runs in testing.

---

*This is a blameless postmortem. Focus is on systems and processes, not individuals.*
```

---

## 4. Architecture Decision Record (ADR)

```markdown
# ADR-0042: Use ArgoCD for GitOps Deployments

**Date**: 2024-01-15  
**Status**: Accepted  
**Deciders**: @platform-team, @engineering-leads  

---

## Context

We need a deployment mechanism for our 12 microservices across 3 environments (dev, staging, production).
Currently: engineers run `kubectl apply` manually. Issues: no audit trail, configuration drift,
no rollback mechanism, inconsistent environments.

---

## Decision

Adopt **ArgoCD** as our GitOps controller. All cluster state will be declared in a dedicated
`gitops-infra` repository. ArgoCD will continuously reconcile cluster state to match the repo.

---

## Considered Alternatives

| Option | Pros | Cons | Rejected Because |
|--------|------|------|-----------------|
| **Flux v2** | Simpler, lighter | Weaker UI, less ecosystem | Team prefers ArgoCD's UI for visibility |
| **Helm + CI push** | Simple pipeline integration | No drift detection, push model | Doesn't solve drift problem |
| **Spinnaker** | Feature-rich | Complex ops, heavyweight | Over-engineered for our scale |
| **Keep kubectl** | Zero migration effort | All current problems remain | Status quo is unacceptable |

---

## Consequences

**Positive**:
- Full audit trail of every deployment (git history)
- Automatic drift reconciliation
- Self-service rollback via `argocd app rollback`
- Multi-cluster support as we grow

**Negative**:
- Learning curve for teams (~1 week ramp-up)
- Additional component to operate in cluster
- Pull model means slight propagation delay (30s–2m) vs push

**Risks**:
- ArgoCD itself becomes a critical dependency; must be HA-deployed
- GitOps repo becomes a high-value target; branch protection required

---

## Implementation Notes

- Phase 1 (Week 1-2): Deploy ArgoCD, migrate staging environments
- Phase 2 (Week 3-4): Migrate production, enable sync waves
- Phase 3 (Month 2): Enable ApplicationSets for multi-cluster
```

---

## 5. Writing Voice & Style

### Do
- **Lead with the problem**, not the solution
- **Show, don't tell**: "the deploy failed with exit code 137" not "there was an error"
- **Use real numbers**: "reduced p99 from 800ms to 45ms" not "significantly improved"
- **One idea per paragraph**; short paragraphs (3-5 lines max)
- **Active voice**: "The migration dropped the index" not "The index was dropped"
- **Callout boxes** for WARNING and TIP — readers scan for these

### Don't
- Marketing language: "seamless", "effortless", "game-changing", "leverage", "utilize"
- Passive voice when active is clearer
- Burying the lede — tell them what happened first
- Assuming context — every runbook must be self-contained
- Blaming in postmortems — systems and processes, never people

### Vietnamese Technical Writing
When writing in Vietnamese, maintain technical terms in English where they are standard
in the Vietnamese DevOps community: Kubernetes, container, deployment, pipeline, microservice, etc.
Translate explanations, not the tooling vocabulary.

Example:
```
GOOD: "Chúng ta sẽ cấu hình HorizontalPodAutoscaler để tự động scale deployment
      khi CPU utilization vượt quá 70%."

BAD:  "Chúng ta sẽ thiết lập Bộ Tự Động Mở Rộng Pod Ngang để..."
```
