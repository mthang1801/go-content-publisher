# System Design Patterns Reference

Tri thức được distill từ production experience tại scale —
không phải textbook theory, mà là battle-tested knowledge.

---

## 1. Consistency Models — Nói đúng tên, không nhầm

```
Strong Consistency:    Read luôn thấy latest write. Cost: latency, availability.
                       Dùng: financial transactions, inventory.

Eventual Consistency:  Read có thể thấy stale data. Cost: complexity ở app layer.
                       Dùng: social feed, view count, like count.

Causal Consistency:    Write có quan hệ nhân quả được thấy đúng thứ tự.
                       Dùng: messaging (reply phải thấy sau message gốc).

Read-your-writes:      User luôn thấy writes của chính mình.
                       Dùng: profile update, post publish.

Monotonic reads:       User không bao giờ thấy data "đi lùi".
                       Dùng: timeline, feed.
```

### CAP Theorem — Practical interpretation

```
P (Partition Tolerance) là bắt buộc trong distributed system.
→ Thực tế: chọn giữa CP và AP khi có partition.

CP (Consistency + Partition Tolerance):
  Ví dụ: HBase, Zookeeper, etcd
  Behavior khi partition: refuse writes, trả error
  Dùng khi: consistency là critical (banking, config management)

AP (Availability + Partition Tolerance):
  Ví dụ: Cassandra, DynamoDB, CouchDB
  Behavior khi partition: serve potentially stale data
  Dùng khi: availability quan trọng hơn consistency (social, e-commerce catalog)

PACELC — Mô hình đầy đủ hơn CAP:
  Khi Partition: CP vs AP (như trên)
  Else (normal operation): Latency vs Consistency trade-off
  DynamoDB: AP/EL (favors latency in normal operation)
  Spanner: CP/EC (favors consistency even without partition)
```

---

## 2. Data Patterns

### Read-heavy systems

```
Strategy 1 — Read Replicas:
  Primary (W) → Replica 1 (R) → Replica 2 (R)
  Trade-off: replication lag → eventual consistency
  When: read:write > 10:1, can tolerate stale reads

Strategy 2 — Cache-Aside (Lazy Loading):
  App → Cache miss → DB → populate cache → return
  Trade-off: cold start penalty, cache invalidation complexity
  When: read pattern predictable, data not too large

Strategy 3 — Read-Through:
  App → Cache → (cache miss) → DB
  Trade-off: all reads go through cache layer
  When: cache warms automatically, simpler app logic

Strategy 4 — CQRS:
  Write model (Commands) → Event → Read model (Queries)
  Trade-off: eventual consistency, operational complexity
  When: read/write shapes are very different, complex domain
```

### Write-heavy systems

```
Strategy 1 — Write-Behind (Write-Back):
  App → Cache (immediate ack) → async persist to DB
  Trade-off: data loss risk if cache fails before write
  When: can tolerate data loss (metrics, logs, non-critical)

Strategy 2 — Event Sourcing:
  Append-only event log → derive current state
  Trade-off: learning curve, eventual consistency, storage growth
  When: audit trail required, complex domain, time-travel needed

Strategy 3 — Sharding:
  Hash sharding: consistent hash, uniform distribution
  Range sharding: good for range queries, hot shard risk
  Directory sharding: flexible, but lookup service = SPOF risk
  Trade-off: cross-shard queries, resharding pain

Strategy 4 — Write Queue + Batch:
  App → Queue → Worker → DB (batch insert)
  Trade-off: write latency (async), ordering complexity
  When: bursty writes, DB write throughput is bottleneck
```

### Caching — Tránh các anti-patterns

```
Cache Stampede (Thundering Herd):
  Problem: nhiều requests đồng thời hit DB khi cache expire
  Fix: probabilistic early expiration, mutex lock on cache miss,
       background refresh trước khi expire

Cache Penetration:
  Problem: request data không tồn tại → luôn hit DB
  Fix: cache NULL values với short TTL, Bloom filter

Cache Avalanche:
  Problem: nhiều cache keys expire cùng lúc
  Fix: jitter TTL (TTL + random(0, 300s)), staggered warm-up

Hot Key:
  Problem: 1 key được access quá nhiều → cache node bottleneck
  Fix: local in-process cache (L1), replicate hot key across nodes,
       client-side hashing
```

---

## 3. Messaging & Event-Driven Patterns

### Message Queue vs Event Streaming

```
Message Queue (RabbitMQ, SQS, BullMQ):
  - Message consumed once, deleted after ack
  - Push model (broker pushes to consumer)
  - Good for: task distribution, work queues, RPC
  - Scale: consumers scale horizontally
  - Retention: short (hours to days)

Event Streaming (Kafka, Kinesis):
  - Events retained, multiple consumers read same event
  - Pull model (consumers pull at their pace)
  - Good for: event sourcing, audit log, data pipeline, fan-out
  - Scale: partition = parallelism unit
  - Retention: long (days to forever)

Decision:
  "Multiple independent consumers need same event" → Streaming
  "One consumer processes task, mark done" → Queue
  "Replay from beginning of time" → Streaming
  "At-most-once delivery OK" → Queue (simpler)
```

### Exactly-Once Delivery — The Hard Problem

```
At-most-once: fire and forget. Simple, but data loss possible.
At-least-once: retry on failure. Duplicates possible → need idempotency.
Exactly-once: hardest. Requires distributed transaction or outbox pattern.

Practical approach — At-least-once + Idempotency:
  1. Assign idempotency key to each operation
  2. Store processed keys (deduplication table, Redis SET)
  3. Check before processing: if key exists → skip, return cached result
  4. Cleanup old keys (TTL > message retention period)

Outbox Pattern (Transactional Messaging):
  BEGIN TRANSACTION
    UPDATE orders SET status = 'confirmed'
    INSERT INTO outbox (event_type, payload) VALUES ('OrderConfirmed', {...})
  COMMIT
  → Separate process reads outbox, publishes to Kafka, marks sent
  Guarantees: event published iff DB transaction committed
```

### Saga Pattern — Distributed Transactions

```
Problem: Update multiple services atomically without 2PC

Choreography Saga:
  Service A event → Service B listens → Service B event → Service C listens...
  Pro: loose coupling, no orchestrator SPOF
  Con: hard to track flow, compensating transactions scattered

Orchestration Saga:
  Orchestrator → call Service A → call Service B → call Service C
  Pro: explicit flow, easy to monitor
  Con: orchestrator = central coordinator, more coupling

Compensation (Rollback):
  Each step must have compensating action:
  BookFlight → (fail) → CancelFlight
  ChargeCreditCard → (fail) → RefundCreditCard

State Machine for Saga:
  PENDING → FLIGHT_BOOKED → HOTEL_RESERVED → CONFIRMED
                          ↓ (fail)
                    HOTEL_FAILED → FLIGHT_CANCELLED → FAILED
```

---

## 4. API Design Patterns

### REST vs GraphQL vs gRPC

```
REST:
  + Universal, cacheable, simple
  - Over/under fetching, multiple round trips
  When: public API, simple CRUD, caching important

GraphQL:
  + Flexible fetching, single endpoint, strong typing
  - Complex caching, N+1 problem, learning curve
  When: complex data requirements, multiple clients with different needs,
        BFF (Backend for Frontend) pattern

gRPC:
  + Strongly typed, streaming, efficient binary (Protobuf), 10x faster than REST
  - Not browser-friendly (needs proxy), harder to debug
  When: internal service-to-service, streaming data, performance critical

WebSocket:
  When: real-time bidirectional (chat, collaborative editing, live dashboard)
  Not when: simple request-response (REST sufficient)

SSE (Server-Sent Events):
  When: server push only (notifications, live feed), simpler than WebSocket
```

### Rate Limiting Algorithms

```
Fixed Window:
  Count requests per fixed time window (0-60s, 60-120s...)
  Problem: burst at window boundary (2x limit in 1 second)

Sliding Window Log:
  Track timestamp of each request, count in last N seconds
  Problem: memory usage = O(requests)

Sliding Window Counter:
  Hybrid: weighted count from current + previous window
  Balance: accuracy + memory efficiency

Token Bucket:
  Bucket refills at rate R tokens/sec, max capacity B
  Allows burst up to B, sustained rate R
  Used by: AWS API Gateway, most CDNs

Leaky Bucket:
  Queue processes at constant rate (output is smooth)
  Excessive requests dropped or queued
  Used when: downstream can't handle bursts

Implementation:
  Single server: Redis INCR + EXPIRE, or Lua script for atomic ops
  Distributed: Redis Cluster, Sliding window with Lua
```

---

## 5. Database Design at Scale

### Sharding Strategy Decision

```
Hash Sharding:
  shard = hash(user_id) % N
  + Even distribution
  - Range queries span all shards
  - Resharding = rehash everything (use consistent hashing)

Range Sharding:
  shard by date range or ID range
  + Range queries efficient
  - Hot shards (recent data more active)
  When: time-series, archive old data frequently

Consistent Hashing:
  Virtual nodes on ring, each server owns arc
  Adding server: only 1/N keys move
  Used by: Cassandra, DynamoDB, Amazon's Dynamo paper

Geo Sharding:
  Shard by geography → data sovereignty, latency
  Complication: global users, cross-geo queries
```

### Database Selection Guide

```
When PostgreSQL:
  Complex queries, joins, ACID, moderate scale (<10TB per node),
  JSONB when flexible schema needed

When Cassandra / DynamoDB:
  Write-heavy, time-series, wide-column queries by partition key,
  Multi-region active-active, massive scale

When Redis:
  Caching, session store, leaderboard (sorted sets),
  rate limiting, pub/sub, distributed locks

When Elasticsearch:
  Full-text search, log analytics, faceted search,
  NOT for primary storage (eventual consistency, no transactions)

When ClickHouse / BigQuery:
  Analytics, OLAP, aggregations over billions of rows,
  NOT for OLTP (batch-optimized)

When MongoDB:
  Flexible schema, document model, moderate scale,
  When team more comfortable vs relational
```

---

## 6. Reliability Patterns

### Circuit Breaker

```
States: CLOSED → OPEN → HALF-OPEN → CLOSED

CLOSED: Normal operation. Count failures.
  Failure threshold exceeded → OPEN

OPEN: Fail fast. Don't call downstream.
  After timeout (30s) → HALF-OPEN

HALF-OPEN: Allow limited test traffic.
  Success → CLOSED
  Failure → OPEN again

Implementation: Hystrix, Resilience4j, or custom with Redis counter

Why it matters: Without circuit breaker, slow downstream = thread exhaustion
= cascading failure = your service down too
```

### Bulkhead Pattern

```
Problem: One slow downstream exhausts shared thread pool
Fix: Separate thread pools per downstream dependency

Service A
├── Thread pool for Service B (10 threads)
├── Thread pool for Service C (10 threads)
└── Thread pool for DB (20 threads)

B going down → only B's pool exhausted
C and DB still work normally
```

### Retry Strategy

```
❌ Naive retry: immediate retry on failure → thundering herd

✅ Exponential backoff with jitter:
  attempt 1: wait 1s + random(0, 1s)
  attempt 2: wait 2s + random(0, 2s)
  attempt 3: wait 4s + random(0, 4s)
  max wait: 30s
  max attempts: 3-5

Retry only on:
  - Network timeout
  - 5xx server errors (502, 503, 504)
  - Idempotent operations

Never retry:
  - 4xx client errors (400, 401, 403, 404)
  - Non-idempotent operations (unless with idempotency key)
```

### SLO / SLA / SLI

```
SLI (Service Level Indicator): actual measurement
  "P99 latency = 230ms over last 5 minutes"

SLO (Service Level Objective): target you set for yourself
  "P99 latency < 300ms for 99.9% of requests"
  "Availability > 99.95%"

SLA (Service Level Agreement): contract with customer
  "We guarantee 99.9% uptime, compensation if breached"

Error Budget:
  99.9% SLO = 0.1% error budget = 8.7 hours/year
  When error budget depleted: freeze features, focus on reliability
  This is Google's SRE model — reliability as shared responsibility

Alert on SLO, not on raw metrics:
  ❌ "Alert if error rate > 1%"
  ✅ "Alert if error budget burn rate > 14x (consume 1hr budget in 5min)"
```

---

## 7. Classic System Design Answers

### URL Shortener (TinyURL)

```
Core: encode long URL → short code, decode short code → redirect

Write path:
  Long URL → Hash (MD5/SHA256) → take first 7 chars → check collision → store
  Storage: {short_code: long_url, created_at, user_id, expiry}

Read path (critical - 100:1 read:write):
  short_code → Cache lookup → (miss) DB lookup → 301/302 redirect
  Cache: Redis with LRU, TTL = link expiry

Scale:
  100M URLs × 500 bytes = 50GB total → fits in Redis
  10K QPS read → Cache hit rate > 99% → < 100 QPS to DB
  Multiple regions: write to primary, read from local cache/replica

Unique code generation:
  Option A: Hash + collision detection (simple, predictable)
  Option B: Global counter + Base62 encode (sequential, no collision)
  Option C: Distributed ID (Snowflake) + Base62 (best for scale)
```

### Notification System

```
Challenges: multi-channel (push, email, SMS), rate limiting per user,
            priority (OTP > marketing), delivery guarantee, retry

Architecture:
  Event → Notification Service → Priority Queue (per channel)
         → Channel Workers → External Providers (FCM, SES, Twilio)
         → Delivery Tracker → Retry/DLQ

Key decisions:
  Push notification: FCM (Android), APNs (iOS) — handle their rate limits
  Email: SES/SendGrid — DMARC/SPF, bounce handling
  SMS: Twilio — expensive, use only for critical (OTP, alerts)

Rate limiting: per user per channel per day (marketing: 1/day, alerts: unlimited)
Deduplication: Redis SET with event_id, TTL = retry window
Failure: exponential backoff, DLQ after 5 failures, alert on-call
```

### Design a Feed System (Twitter/Instagram)

```
Fan-out on Write (Push model):
  When user posts → push to all followers' feed tables
  Read: instant (pre-computed)
  Problem: celebrity with 10M followers → 10M writes per post

Fan-out on Read (Pull model):
  Read time: merge feeds from all followings
  Problem: read is slow (query N followings)

Hybrid (Twitter's approach):
  Normal users: fan-out on write
  Celebrities (>1M followers): fan-out on read
  Read time: merge pre-computed feed + celebrity posts

Storage:
  Redis sorted set: {user_id → [(timestamp, post_id)...]} with max 1000 items
  MySQL/Cassandra: post content, user data
  Object store (S3): media

Pagination: cursor-based (timestamp of last seen post)
```
