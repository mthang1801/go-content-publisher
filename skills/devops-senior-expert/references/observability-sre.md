# Observability & SRE — Production Monitoring

## Table of Contents
1. [The Three Pillars](#pillars)
2. [SLO / SLI / Error Budget](#slo)
3. [Prometheus & Grafana](#prometheus)
4. [OpenTelemetry](#otel)
5. [Alert Design](#alerting)
6. [Incident Response](#incident)

---

## 1. The Three Pillars

| Pillar | Tool Stack | Answers |
|--------|-----------|---------|
| **Metrics** | Prometheus + Grafana, Datadog | "Is the system healthy? What's the rate?" |
| **Logs** | Loki + Grafana, ELK, CloudWatch | "What happened? What was the error?" |
| **Traces** | Tempo + Grafana, Jaeger, X-Ray | "Where is the latency? Which service?" |

**Rule**: Alerts from metrics, investigation via logs + traces. Don't alert on logs.

---

## 2. SLO / SLI / Error Budget

### Defining Good SLIs

```
Availability SLI = good_requests / total_requests
Latency SLI      = requests_under_threshold / total_requests
Throughput SLI   = actual_rps / target_rps (for batch systems)
```

### SLO Configuration Example

```yaml
# Pyrra / Sloth SLO definition
apiVersion: sloth.slok.dev/v1
kind: PrometheusServiceLevel
metadata:
  name: order-service-slo
spec:
  service: "order-service"
  labels:
    team: platform
    tier: critical

  slos:
    - name: "requests-availability"
      objective: 99.9    # 99.9% over rolling 30 days = 43.8 min downtime/month
      description: "Order service should serve 99.9% of requests successfully"

      sli:
        events:
          errorQuery: |
            sum(rate(http_requests_total{service="order-service",status=~"5.."}[{{.window}}]))
          totalQuery: |
            sum(rate(http_requests_total{service="order-service"}[{{.window}}]))

      alerting:
        name: OrderServiceHighErrorRate
        labels:
          severity: critical
        annotations:
          summary: "Order service burning error budget"
        pageAlert:
          labels:
            severity: page     # wake someone up
        ticketAlert:
          labels:
            severity: ticket   # create ticket, no page

    - name: "requests-latency"
      objective: 99.0
      description: "99% of requests should complete in under 500ms"
      sli:
        events:
          errorQuery: |
            sum(rate(http_request_duration_seconds_bucket{
              service="order-service",le="0.5",status!~"5.."}[{{.window}}]))
          totalQuery: |
            sum(rate(http_request_duration_seconds_count{service="order-service"}[{{.window}}]))
```

### Error Budget Burn Rate Alerts

```yaml
# Multi-window, multi-burn-rate (Google SRE Book recommendation)
# Fast burn: 2% budget in 1h = 14.4x burn rate → PAGE immediately
# Slow burn: 5% budget in 6h = 6x burn rate  → TICKET

- alert: ErrorBudgetBurnFast
  expr: |
    (
      error_budget_burn_rate:order_service:1h > 14.4
      AND
      error_budget_burn_rate:order_service:5m > 14.4
    )
  for: 2m
  labels:
    severity: page
  annotations:
    summary: "Order service burning error budget at {{ $value }}x rate"
    runbook: "https://runbooks.internal/order-service/high-error-rate"
```

---

## 3. Prometheus & Grafana

### Instrumentation — Four Golden Signals

```go
// In application code (Go example with prometheus client)
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
        []string{"method", "path", "status"},
    )

    // Saturation: queue depth, goroutine count, connection pool usage
    activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "db_connections_active",
        Help: "Active database connections",
    })
)
```

### Recording Rules — Pre-compute Expensive Queries

```yaml
groups:
  - name: order_service.rules
    interval: 30s
    rules:
      # Pre-compute 5m error rate (used in dashboards and alerts)
      - record: job:http_requests_total:error_rate5m
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) by (job)
          /
          sum(rate(http_requests_total[5m])) by (job)

      # P99 latency
      - record: job:http_request_duration_seconds:p99_5m
        expr: |
          histogram_quantile(0.99,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (job, le)
          )
```

---

## 4. OpenTelemetry — Unified Instrumentation

```yaml
# otel-collector config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  memory_limiter:
    limit_mib: 512
  resourcedetection:
    detectors: [env, eks, ec2]   # auto-detect cloud metadata

exporters:
  prometheusremotewrite:
    endpoint: "http://prometheus:9090/api/v1/write"
  otlp/tempo:
    endpoint: "http://tempo:4317"
    tls:
      insecure: true
  loki:
    endpoint: "http://loki:3100/loki/api/v1/push"

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch, resourcedetection]
      exporters: [otlp/tempo]
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [prometheusremotewrite]
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [loki]
```

---

## 5. Alert Design — Fighting Alert Fatigue

### Alert Hierarchy

```
P1 (Page now)      → Revenue impact, customer data at risk, full outage
P2 (Page in 15m)   → Degraded service, SLO burning fast
P3 (Slack/Ticket)  → Warning trend, approaching threshold
Info               → FYI, no action needed (these should be dashboards, not alerts)
```

### Alert Template — Every Alert Needs These

```yaml
- alert: OrderServiceHighLatency
  expr: job:http_request_duration_seconds:p99_5m{job="order-service"} > 1.0
  for: 5m            # must be consistently bad, not a spike
  labels:
    severity: page
    team: platform
    service: order-service
  annotations:
    summary: "Order service P99 latency {{ $value | humanizeDuration }} > 1s"
    description: |
      The 99th percentile latency for order-service has exceeded 1 second
      for the past 5 minutes. Current value: {{ $value }}s.
      This may indicate database slowness, downstream service degradation,
      or resource saturation.
    runbook: "https://runbooks.internal/order-service/high-latency"
    dashboard: "https://grafana.internal/d/order-service"
```

---

## 6. Incident Response — Full Framework

### Severity Classification

| Level | Impact | Response Time | Escalation |
|---|---|---|---|
| P1 | Full outage / data loss | Immediate | All hands + management |
| P2 | Major feature down, >10% users | 15 min | On-call + team lead |
| P3 | Minor degradation, workaround exists | 4 hours | On-call |
| P4 | Non-urgent bug | Next sprint | Ticket only |

### Incident Timeline Template

```markdown
## Incident: [INC-1234] Order service latency spike

**Severity**: P2
**Duration**: 14:23 - 16:05 UTC (1h 42m)
**Impact**: Order placement latency p99 > 5s (SLO: 500ms), ~15% of orders failed

### Timeline
14:23 - Alert fired: OrderServiceHighLatency (p99 > 1s)
14:27 - On-call [Name] acknowledged
14:35 - Identified spike in DB query times via traces
14:42 - Found new index missing after DB migration at 14:15
14:55 - Created missing index (online, non-blocking)
15:10 - Latency returning to normal
16:05 - All metrics nominal, incident closed

### Root Cause
DB migration at 14:15 added a new query pattern without a corresponding index.
As traffic increased after business hours, full table scans caused latency spike.

### Contributing Factors
- No performance testing of migrations in staging with production-scale data
- Migration ran automatically on deploy without latency validation step

### Action Items
| Item | Owner | Due |
|------|-------|-----|
| Add migration validation step to CI pipeline | @devops | 2024-02-15 |
| Set up slow query alerts on RDS | @dba | 2024-02-10 |
| Load test staging DB with production data volume | @dev | 2024-02-20 |
```
