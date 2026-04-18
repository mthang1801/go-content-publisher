# PostgreSQL Optimization Reference

## 1. Đọc EXPLAIN ANALYZE — Nhanh và Đúng

### Anatomy của một execution plan

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) 
SELECT ...;
```

```
Seq Scan on orders  (cost=0.00..18340.00 rows=850000 width=64)
                     (actual time=0.123..892.456 rows=847293 loops=1)
  Filter: (status = 'pending')
  Rows Removed by Filter: 2707
  Buffers: shared hit=8340 read=0
Planning Time: 0.234 ms
Execution Time: 934.123 ms
```

**Đọc từ trong ra ngoài, từ dưới lên trên** — node trong cùng chạy trước.

### Cost numbers — hiểu đúng

```
cost=0.00..18340.00
      ↑        ↑
  startup    total
  cost       cost

Đơn vị: arbitrary planner units (không phải ms)
seq_page_cost = 1.0 (baseline)
random_page_cost = 4.0 (SSD nên set 1.1-2.0)
cpu_tuple_cost = 0.01
```

### Dấu hiệu plan có vấn đề

```
🔴 rows estimate vs actual lệch > 10x:
   rows=1000 vs actual rows=850000
   → Stale statistics → chạy ANALYZE

🔴 Seq Scan trên bảng lớn với filter có selectivity cao:
   Filter: (user_id = 12345)  → cần index
   Filter: (status = 'active') nếu 90% là active → seq scan có thể đúng

🔴 Hash Join với large batches:
   Batches: 8 (initial: 1)
   → work_mem quá thấp → hash spill to disk

🔴 Nested Loop với large outer:
   Loops: 850000 → nhân với inner cost = catastrophic
   → Thường cần index trên join column của inner table

🟡 Recheck Cond trên Bitmap Heap Scan:
   Rows Removed by Index Recheck: 45000
   → Index too lossy (array/GIN) hoặc cần covering index

🟡 Sort với external merge:
   Sort Method: external merge  Disk: 45678kB
   → Tăng work_mem hoặc thêm index để avoid sort
```

### Key metrics để compare plans

```sql
-- Trước khi test: disable parallel để compare fair
SET max_parallel_workers_per_gather = 0;

-- Chạy nhiều lần, lấy lần 2+ (warm cache)
EXPLAIN (ANALYZE, BUFFERS) SELECT ...;

-- Metrics quan trọng:
-- actual time: wall clock time của node đó
-- Buffers shared hit: từ shared_buffers (fast)
-- Buffers shared read: từ disk/OS cache (slow)
-- rows: accuracy của planner estimate
```

---

## 2. Index Strategy — Principal Level Thinking

### Index types và khi nào dùng

```sql
-- B-tree (default) — range queries, equality, sort
CREATE INDEX idx_orders_created ON orders(created_at);
CREATE INDEX idx_orders_status_created ON orders(status, created_at DESC);

-- Partial index — subset data, nhỏ hơn, faster
CREATE INDEX idx_orders_pending ON orders(created_at) 
WHERE status = 'pending';
-- Chỉ index pending orders → 10x nhỏ hơn full index

-- Covering index — index-only scan, không touch heap
CREATE INDEX idx_orders_covering ON orders(user_id, created_at) 
INCLUDE (status, total_amount);
-- Query SELECT status, total_amount WHERE user_id=? ORDER BY created_at
-- chỉ đọc index, không cần heap access

-- GIN — full-text search, JSONB, arrays
CREATE INDEX idx_products_tags ON products USING gin(tags);
CREATE INDEX idx_events_data ON events USING gin(data jsonb_path_ops);

-- BRIN — time-series data, naturally ordered (very small, fast insert)
CREATE INDEX idx_logs_ts ON logs USING brin(timestamp) WITH (pages_per_range=128);
-- Phù hợp: table hàng tỷ rows, query theo time range, append-only

-- Hash — equality only, không sort
-- Hiếm dùng hơn B-tree, chỉ khi equality-only và very high cardinality
```

### Multi-column index — thứ tự cột

```
Rule: (equality columns first) → (range column last) → (sort column last)

Query: WHERE status = 'active' AND created_at > '2024-01-01' ORDER BY id DESC
Index: (status, created_at, id) ✅
Index: (created_at, status, id) ❌ → không dùng được cho status equality

Rule: Planner có thể dùng prefix của index
Index: (a, b, c)
→ WHERE a = 1               ✅ dùng được
→ WHERE a = 1 AND b = 2     ✅ dùng được
→ WHERE b = 2               ❌ không dùng được (không có prefix a)
→ WHERE a = 1 AND c = 3     ⚠️  dùng được nhưng chỉ filter a, không filter c
```

### Index bloat và maintenance

```sql
-- Check index bloat
SELECT
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0  -- unused indexes!
ORDER BY pg_relation_size(indexrelid) DESC;

-- Rebuild bloated index (concurrent, no lock)
REINDEX INDEX CONCURRENTLY idx_orders_status;
```

### Unused index detection

```sql
SELECT 
    schemaname || '.' || tablename AS table,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS size,
    idx_scan AS scans_since_reset
FROM pg_stat_user_indexes
WHERE idx_scan < 50  -- ít được dùng
  AND indexrelid NOT IN (
    SELECT conindid FROM pg_constraint  -- loại trừ constraint indexes
  )
ORDER BY pg_relation_size(indexrelid) DESC;
```

---

## 3. Query Optimization Patterns

### Pattern: N+1 query (ORM trap)

```sql
-- ❌ N+1: 1 query lấy orders + N queries lấy user mỗi order
SELECT * FROM orders LIMIT 100;
-- rồi: SELECT * FROM users WHERE id = 1;
-- rồi: SELECT * FROM users WHERE id = 2; ...

-- ✅ Fix: JOIN hoặc IN clause
SELECT o.*, u.name, u.email
FROM orders o
JOIN users u ON u.id = o.user_id
LIMIT 100;

-- ✅ Hoặc: batch load
SELECT * FROM users WHERE id = ANY(ARRAY[1,2,3,...100]);
```

### Pattern: COUNT(*) trên large table

```sql
-- ❌ COUNT(*) = seq scan nếu không có WHERE
SELECT COUNT(*) FROM orders;  -- 50M rows = very slow

-- ✅ Estimate cho UI (fast, slightly inaccurate)
SELECT reltuples::bigint AS estimate
FROM pg_class
WHERE relname = 'orders';

-- ✅ Partial count với index
SELECT COUNT(*) FROM orders WHERE status = 'pending';
-- Nếu có index: (status) WHERE status = 'pending' → index scan

-- ✅ Materialized count cho dashboard
CREATE MATERIALIZED VIEW order_counts AS
SELECT status, COUNT(*) FROM orders GROUP BY status;
-- Refresh theo schedule hoặc trigger
```

### Pattern: Pagination với OFFSET

```sql
-- ❌ OFFSET lớn = catastrophic (đọc rồi bỏ qua N rows)
SELECT * FROM orders ORDER BY created_at DESC OFFSET 100000 LIMIT 20;
-- Planner phải đọc 100020 rows, bỏ 100000

-- ✅ Keyset pagination (cursor-based)
SELECT * FROM orders
WHERE created_at < $last_cursor  -- cursor từ lần query trước
ORDER BY created_at DESC
LIMIT 20;
-- Dùng index scan, luôn O(log n)

-- ✅ Keyset với multiple sort columns
SELECT * FROM orders
WHERE (created_at, id) < ($last_ts, $last_id)
ORDER BY created_at DESC, id DESC
LIMIT 20;
-- Index: (created_at DESC, id DESC)
```

### Pattern: JSONB query optimization

```sql
-- ❌ Cast mất index
SELECT * FROM events WHERE (data->>'user_id')::int = 12345;

-- ✅ Dùng JSONB operator
SELECT * FROM events WHERE data @> '{"user_id": 12345}';
-- Index: CREATE INDEX ON events USING gin(data jsonb_path_ops);

-- ❌ ANY với unnest — không dùng được index
SELECT * FROM events WHERE data->>'type' = ANY(ARRAY['click', 'view']);

-- ✅ IN clause
SELECT * FROM events WHERE data->>'type' IN ('click', 'view');
-- Hoặc: functional index
CREATE INDEX ON events ((data->>'type'));
```

### Pattern: Window function optimization

```sql
-- ❌ Correlated subquery trong window
SELECT *, 
  (SELECT SUM(amount) FROM orders o2 
   WHERE o2.user_id = o1.user_id AND o2.created_at <= o1.created_at) 
FROM orders o1;

-- ✅ Window function (single pass)
SELECT *,
  SUM(amount) OVER (
    PARTITION BY user_id 
    ORDER BY created_at 
    ROWS UNBOUNDED PRECEDING
  ) AS running_total
FROM orders;
```

---

## 4. Statistics và Planner Configuration

### Stale statistics — dấu hiệu và fix

```sql
-- Check last analyze time
SELECT schemaname, tablename, last_analyze, last_autoanalyze, 
       n_live_tup, n_dead_tup, n_mod_since_analyze
FROM pg_stat_user_tables
WHERE n_mod_since_analyze > 1000  -- nhiều changes chưa được analyze
ORDER BY n_mod_since_analyze DESC;

-- Manual analyze (targeted)
ANALYZE VERBOSE orders;

-- Increase statistics target cho columns có bad estimates
ALTER TABLE orders ALTER COLUMN status SET STATISTICS 500;
-- Default: 100. Tăng lên 200-500 cho high-cardinality columns với skewed distribution
ANALYZE orders;
```

### Planner hints (khi cần override)

```sql
-- Disable specific join type (debugging only, không dùng production lâu dài)
SET enable_hashjoin = off;
SET enable_nestloop = off;
SET enable_seqscan = off;  -- force index scan

-- Session-level: chỉ ảnh hưởng connection hiện tại
-- Không SET globally trừ khi rất chắc chắn
```

### work_mem tuning

```sql
-- work_mem ảnh hưởng: Sort, Hash Join, Hash Aggregate
-- Default: 4MB — thường quá thấp

-- Check sort spill to disk
EXPLAIN (ANALYZE, BUFFERS) SELECT ... ORDER BY ...;
-- Nếu thấy: Sort Method: external merge Disk: XXXkB → tăng work_mem

-- Set per-session cho heavy queries
SET work_mem = '256MB';
SELECT ...;
RESET work_mem;

-- CẢNH BÁO: work_mem × max_connections có thể = OOM
-- Công thức safe: work_mem = RAM / (max_connections × parallel_workers)
-- Ví dụ: 32GB RAM, 100 connections, 4 workers = 32GB/400 = ~80MB max
```

---

## 5. CTE và Subquery Optimization

```sql
-- PostgreSQL 12+: CTE không còn optimization barrier mặc định
-- Trước 12: CTE luôn materialized (chạy riêng, không thể optimize cùng outer)

-- ✅ PostgreSQL 12+: CTE inline (optimizer có thể push down predicates)
WITH recent_orders AS (
  SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '7 days'
)
SELECT * FROM recent_orders WHERE user_id = 123;
-- Planner có thể combine cả 2 conditions

-- Nếu cần force materialization (CTE như temp table):
WITH recent_orders AS MATERIALIZED (
  SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '7 days'
)
SELECT * FROM recent_orders WHERE user_id = 123;

-- Lateral join — thay thế correlated subquery
SELECT u.*, recent.*
FROM users u
CROSS JOIN LATERAL (
  SELECT * FROM orders o
  WHERE o.user_id = u.id
  ORDER BY created_at DESC
  LIMIT 5
) recent;
-- Efficient: chạy inner query 1 lần per user, dùng index
```
