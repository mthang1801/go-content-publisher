# System Design Documentation Integration

Áp dụng workflow R1–R7 + skill `technical-doc-writer` cho system design topics.
System design docs có narrative character riêng — đây là những gì khác biệt.

---

## Style Mapping

```
System Concept docs (CAP theorem, consistent hashing, event sourcing) → Concept-First
System Design docs (design URL shortener, design Twitter feed) → Problem-Centric
Architecture Decision docs → ADR format (xem references/architecture-decisions.md)
Component Deep-dives (Kafka internals, Redis data structures) → Concept-First
Operational docs (runbooks, incident response) → Template-heavy Concept-First
```

---

## DEFINE — Tension Patterns cho System Design

### Pattern: Scale Breaking Point

Mạnh nhất — đặt người đọc vào thời điểm system phải thay đổi.

```
Template:
"[System] của bạn hoạt động tốt với [X users/QPS/data].
Sau [event — launch, viral post, partnership], traffic tăng [N]x trong [time].
[Specific metric] từ [baseline] lên [broken state].
[Business impact — revenue loss, user complaint, SLA breach].
Root cause: [1 câu — không phải code bug mà architecture assumption sai].
Đây là [concept/pattern] — và hiểu nó là điều phân biệt system survive hay crumble."
```

**Ví dụ (Database Sharding)**:
```
"Startup của bạn vừa được featured trên Product Hunt.
Trong 6 giờ, user tăng từ 50K lên 800K. Database CPU đạt 100%.
Query response time: từ 20ms lên 12 giây. Checkout bị timeout.
$180K revenue lost trong một buổi chiều.

Root cause: single PostgreSQL node không scale writes vượt ~50K QPS trên commodity hardware.
Read replicas đã cạn kiệt. Vertical scaling đã đến giới hạn.
Lần đầu bạn thực sự hiểu tại sao sharding tồn tại."
```

### Pattern: Hidden Assumption Violated

Cho distributed systems concepts — khi assumption tưởng rõ ràng lại sai.

```
Template:
"Hầu hết engineers assume [assumption phổ biến].
Điều này đúng trong 99% cases — cho đến khi [rare but real scenario].
Khi đó: [catastrophic outcome].
[Company X] học bài học này theo cách đắt nhất vào [year]: [brief incident description].
[Concept/Pattern] là lý do tại sao distributed systems cần [approach] thay vì [naive approach]."
```

**Ví dụ (Idempotency)**:
```
"Hầu hết engineers assume: gửi một request thì một action xảy ra.
Network không reliable. Client không biết request đã đến chưa.
Timeout xảy ra. Client retry. Action xảy ra hai lần.
Customer bị charge hai lần. Bank không vui.

Stripe học bài này từ sớm. Đó là lý do mọi Stripe API call yêu cầu
Idempotency-Key header — và mọi payment system serious đều implement tương tự."
```

### Pattern: Interview Decision Tree (Problem-Centric)

```
Template:
"Bạn vừa nhận câu hỏi: 'Design [system X].'
Interviewer nhìn bạn. Clock bắt đầu chạy.

Đây không phải câu hỏi về [system X]. Đây là câu hỏi về cách bạn
navigate ambiguity, make trade-offs, và communicate technical decisions
dưới time pressure.

Bài này là playbook — từng phase, từng quyết định, từng câu bạn nên nói."
```

---

## VISUAL — System Design Diagram Strategy

### Tối thiểu 3 diagram cho system design doc

```
Level 1 — System Context (Beginner accessible):
  Ai dùng system? Kết nối gì với external world?
  → ASCII boxes, không technical detail
  → Caption: user journey + external dependencies

Level 2 — Component Overview (Experienced):
  Services nào? Data flows như thế nào? Where are the boundaries?
  → Mermaid graph hoặc ASCII topology
  → Caption: critical path + bottleneck identification

Level 3 — Deep Dive on Critical Path (Expert):
  Request flow chi tiết qua critical path
  → Mermaid sequence diagram
  → Caption: latency budget + failure points
```

### Caption formula cho architecture diagrams

```
Format: [What is happening] + [Why it's designed this way] + [What breaks if you do it differently]

Ví dụ:
"Hình: Write path đi qua Kafka trước khi persist — Order Service
 không gọi trực tiếp Warehouse, Email, Analytics services.
 Thêm consumer mới: zero change đến Order Service.
 Trade-off: up to 500ms delivery delay khi Kafka bị lag."

"Hình: CDN layer serve 95% của static assets (images, JS, CSS).
 Origin chỉ nhận dynamic requests. Cache hit rate target: 92%.
 Cache miss pattern: cold start sau deployment = 3-5 min traffic spike đến origin."
```

---

## CODE Section — SQL, Config, Scripts cho System Design

System design docs thường có 3 loại "code" khác nhau:

### Type 1 — Architecture-as-Code (Terraform, K8s manifests)

```markdown
### Ví dụ 2: Intermediate — Multi-AZ RDS với Read Replica

[INTRODUCE]:
Bạn vừa hit read bottleneck — single DB node đang nhận 8K QPS read,
query response time tăng từ 10ms lên 200ms ở P99. CPU 85%.

Thêm read replica giải quyết read scaling. Nhưng cần trả lời:
- Replica ở AZ nào? (cross-AZ vs same-AZ latency)
- App routing: round-robin hay sticky?
- Failure handling khi replica lag > threshold?

Infrastructure as Code đảm bảo reproducible setup:

[CODE]:
```hcl
# rds.tf — Multi-AZ PostgreSQL với Read Replica
# Pattern: Primary (write) + Replica (read) + Automatic failover

resource "aws_db_instance" "primary" {
  identifier           = "orders-primary"
  engine               = "postgres"
  engine_version       = "15.4"
  instance_class       = "db.r6g.xlarge"  # 4 vCPU, 32 GB RAM
  multi_az             = true              # ✅ Sync replica trong AZ khác (auto failover ~60s)
  
  allocated_storage     = 500
  storage_type          = "gp3"
  storage_throughput    = 500  # MB/s
  iops                  = 12000

  # ✅ Automated backups: 7-day retention, daily snapshots
  backup_retention_period = 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "mon:04:00-mon:05:00"

  # ⚠️ Performance Insights: phải enable để debug slow queries
  performance_insights_enabled = true
  monitoring_interval          = 60  # Enhanced monitoring mỗi 60s
}

resource "aws_db_instance" "replica" {
  identifier          = "orders-replica"
  replicate_source_db = aws_db_instance.primary.id
  instance_class      = "db.r6g.large"  # Replica nhỏ hơn primary nếu read pattern simpler
  
  # ✅ Replica ở AZ khác để không compete I/O với primary
  availability_zone = "us-east-1b"  # Primary ở us-east-1a

  # ⚠️ Không bật multi_az cho replica — nếu muốn cross-AZ read, tạo thêm replica
  multi_az = false
}
```

[TẠI SAO?]:
> **Tại sao Replica nhỏ hơn Primary?**
> Read queries thường simpler — aggregations, lookups, joins đơn giản.
> Primary phải handle writes (WAL generation), vacuum, checkpoint.
> Nếu read pattern phức tạp (analytics), size replica bằng primary.
>
> **Tại sao different AZ cho replica?**
> Same AZ: replica replication lag = 0-1ms (network trong AZ rất nhanh)
> Different AZ: replica replication lag = 1-5ms (cross-AZ ~0.5ms RTT)
> Trade-off: different AZ isolates từ AZ-level failure, không block primary writes khi AZ down.

[KẾT LUẬN]:
> Multi-AZ Primary: automatic failover trong ~60s. RPO = 0 (synchronous replication).
> Read Replica: async, lag thường < 5ms. App phải tolerate stale reads.
>
> **Khi nào add thêm replica**: P99 read latency tăng, replica CPU > 70%,
> hoặc cần geographic distribution.
> **Không dùng replica cho**: write operations, transactions cần read-your-writes.
```

### Type 2 — Diagnostic Queries và Runbooks

```markdown
### Ví dụ 3: Advanced — Replica Lag Monitoring và Auto-routing

[INTRODUCE]:
Replica lag tăng đột ngột → queries đọc stale data → UX issues.
Cần: detect lag > threshold, route reads về primary, alert on-call.

[CODE]:
```sql
-- replica_health.sql — Monitoring và routing decisions
-- Chạy trên replica mỗi 30 giây bởi monitoring service

-- Metric 1: Replication delay
SELECT
    EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))::INT 
    AS replication_lag_seconds,
    pg_last_wal_receive_lsn() AS receive_lsn,
    pg_last_wal_replay_lsn() AS replay_lsn,
    pg_is_in_recovery() AS is_replica;

-- Routing decision logic (application layer pseudo-code):
-- IF lag_seconds > 30:
--   route_all_reads_to_primary()
--   alert_oncall("Replica lag exceeded 30s threshold")
-- ELSE IF lag_seconds > 5:
--   route_critical_reads_to_primary()  -- e.g., post-write reads
-- ELSE:
--   route_reads_to_replica()           -- normal operation
```

> **Tại sao threshold 30s và 5s?**
> 30s: SLA violation risk — user sees data older than 30s.
> 5s: read-after-write consistency concern — user updates profile,
> refreshes page, sees old data (confusing).
```

### Type 3 — Decision Tables và Trade-off Analysis

```markdown
### Ví dụ 1: Basic — Chọn Message Queue vs Event Streaming

[INTRODUCE]:
Order Service cần notify: Email Service, Analytics, Warehouse.
3 options: direct HTTP calls, message queue, event streaming.

[CODE (Table format)]:

| Criteria           | Direct HTTP       | Message Queue (BullMQ) | Event Streaming (Kafka) |
|--------------------|-------------------|------------------------|------------------------|
| Coupling           | Tight             | Loose                  | Very loose             |
| Delivery guarantee | Best-effort       | At-least-once          | At-least-once          |
| Multiple consumers | N copies needed   | N queues needed        | 1 topic, N consumers   |
| Replay history     | No                | No (deleted after ack) | Yes (configurable TTL) |
| Ops complexity     | Low               | Medium                 | High                   |
| Latency            | Sync (ms)         | Near-real-time         | Near-real-time         |
| When to use        | Simple, 1 consumer | Task distribution     | Fan-out, audit, replay |

Decision cho Order Service: **Kafka**
Lý do: 3+ consumers, audit trail requirement, future consumers unknown.
Trade-off accepted: Kafka operational overhead vs future flexibility.
```

---

## PITFALLS — System Design Fatal Patterns

### Template: Architectural Anti-pattern

```markdown
### 🔴 Pitfall #N — [Tên mô tả hậu quả, không phải kỹ thuật]

Architecture này appear trong 40% system design interviews — và trong nhiều startups:

[ASCII diagram của anti-pattern]

Trông ổn khi traffic thấp. Sau đây là failure cascade khi traffic tăng:
[Step-by-step failure cascade với numbers]

Chúng tôi thấy pattern này fail tại [scale] với [consequence].

**Root cause**: [1-2 câu nguyên nhân kỹ thuật]

**Fix**: [Correct architecture với diagram]

**Detection**: [Metric/symptom báo hiệu vấn đề trước khi fail hoàn toàn]
```

**Ví dụ thực tế:**

```markdown
### 🔴 Pitfall #1 — Synchronous Chain of Death

Architecture phổ biến — đặc biệt khi convert monolith sang microservices:

Client → API Gateway → Order Svc → [Payment Svc → User Svc → Inventory Svc]

Trông clean. Trong development, hoạt động tốt.
Trong production, với real latency:

Payment Svc P99: 200ms
User Svc P99: 150ms
Inventory Svc P99: 100ms
Network overhead × 3: 30ms

Order API P99 ≥ 200 + 150 + 100 + 30 = 480ms (best case)

Inventory Svc bắt đầu slow (100ms → 3s during flash sale):
Order API P99 = 200 + 150 + 3000 = 3350ms → user abandon

Inventory Svc timeout spike:
Order thread pool exhausted → Order Svc unavailable
API Gateway returns 503 → User cannot checkout
→ **Cascading failure từ 1 service ảnh hưởng toàn bộ checkout**

Root cause: synchronous coupling + no circuit breaker = blast radius = entire system.

Fix:
1. Circuit breaker trên mỗi downstream call
2. Timeout budget: Order SLA 500ms → Payment max 300ms, User max 100ms, Inventory max 50ms
3. Async non-critical: Inventory reservation → queue (eventual consistency)
4. Cache User data (changes rarely) → eliminate synchronous User Svc call

Detection: P99 latency tăng gradual, thread pool metrics, downstream timeout rate
```

---

## RECOMMEND — System Design Learning Path

```markdown
## 6. RECOMMEND

Bạn vừa hiểu [concept] — [1 câu về vị trí của nó trong distributed systems landscape].

Điểm thú vị: [concept] không đứng một mình. Nó thường xuất hiện cùng với
[related concept] và [another related concept] trong production architectures.
Hiểu cách 3 thứ này interact với nhau là điều phân biệt senior với junior
khi design system at scale.

| Mở rộng | Khi nào | Lý do | Link |
|---------|---------|-------|------|
| [Related concept 1] | [Condition] | [Why it complements] | [File] |
| [Related concept 2] | [Condition] | [Why it's next level] | [File] |
| [Production resource] | [Condition] | [Real-world application] | [Link] |
```

**Ví dụ sau bài về Event Sourcing:**

```markdown
## 6. RECOMMEND

Bạn vừa hiểu Event Sourcing — cách model state như immutable sequence of events,
không phải snapshot hiện tại. Đây là foundation của audit-first systems.

Nhưng Event Sourcing không phải silver bullet — nó shine khi kết hợp với
CQRS (Command Query Responsibility Segregation): writes theo event model,
reads được denormalize thành optimized read models. Và khi system grows
multi-service, Saga Pattern giải quyết distributed transactions trên event foundation.
3 concepts này — Event Sourcing, CQRS, Saga — là the holy trinity của
complex domain modeling ở scale.

| Mở rộng | Khi nào | Lý do | Link |
|---------|---------|-------|------|
| CQRS Pattern | Read/write shapes diverge | Event Sourcing alone still has read problem | [doc link] |
| Saga Pattern | Multi-service transactions | Events + compensation = distributed ACID | [doc link] |
| Axon Framework (Java) | Implement in production | Battle-tested ES+CQRS framework | axoniq.io |
| Martin Fowler's CQRS | Deep understand | Canonical reference | martinfowler.com |
```
