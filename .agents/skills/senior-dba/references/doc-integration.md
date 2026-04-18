# DB Documentation Integration

Hướng dẫn áp dụng workflow R1–R7 + skill `technical-doc-writer`
khi viết tài liệu về database topics. DB documentation có đặc thù riêng
so với application code documentation.

---

## DB Doc Style Mapping

```
DB Concept docs (index, vacuum, MVCC, replication) → Concept-First
DB Problem docs (slow queries, bloat, deadlock, connection exhaustion) → Problem-Centric
DB Operation docs (backup, migration, monitoring setup) → Concept-First với heavy TEMPLATE section
```

---

## DEFINE — Tension Patterns cho DB Topics

### Pattern: Production Incident Opening

Mạnh nhất cho DB topics — vì hầu hết DBA knowledge đến từ incidents.

```
Template:
"[Thời điểm]. Alert đến: [metric cụ thể — số liệu, không mô tả].
[Metric bình thường A]. [Metric bình thường B]. Chỉ có [X] là khác.
[Hậu quả kinh doanh — khách không mua được / data sai / downtime].
Root cause: [1 câu]. Bài này giải thích tại sao và cách phòng ngừa."

Ví dụ (VACUUM):
"2:47 AM. Alert: PostgreSQL trả về lỗi 'database is not accepting commands'.
CPU bình thường. Memory bình thường. Disk 98% full — nhưng chỉ có 50GB data.
Production down 23 phút. Root cause: transaction ID wraparound.
48GB đó là dead tuples mà autovacuum không kịp thu hồi."
```

### Pattern: Silent Degradation Opening

Cho các vấn đề tích lũy theo thời gian (bloat, index drift, stale stats):

```
Template:
"[Query/operation] này chạy [X ms] khi launch. 6 tháng sau: [10X ms].
Không có code change. Không có schema change. Chỉ có data grow.
[Số liệu cụ thể — table size, index size, dead tuples].
Đây là [concept] — và nó ảnh hưởng mọi PostgreSQL database theo thời gian."

Ví dụ (Index Bloat):
"Query này chạy 8ms khi go-live. 8 tháng sau: 340ms. Không có gì thay đổi
ngoài data volume — từ 2M rows lên 18M rows, nhưng response time tăng 40x.
Table size tăng 9x là bình thường. Index size tăng 45x thì không.
Đây là index bloat — và không có gì tự fix nó."
```

### Pattern: Invisible Constraint Opening

Cho architectural limits mà dev thường không biết:

```
Template:
"PostgreSQL có một giới hạn mà hầu hết người dùng không biết cho đến khi
nó hit them in production: [giới hạn cụ thể].
Khi đó xảy ra: [hậu quả nghiêm trọng].
Điều này không xuất hiện trong docs cơ bản. Đây là cách hiểu và tránh nó."

Ví dụ (Transaction ID Wraparound):
"PostgreSQL có transaction ID (XID) counter — 32-bit integer.
Khi nó wrap around sau ~2 tỷ transactions, database sẽ từ chối mọi write.
Không phải gradually degrade — là hard stop.
Amazon RDS đã downtime vì điều này. Bài này giải thích cơ chế và prevention."
```

---

## VISUAL — DB-Specific Diagram Patterns

### Execution Plan Diagram

```
EXPLAIN output dưới dạng tree với annotations:

Hash Join  (cost=2847..45892 rows=18340)  ✅ Fast: in-memory hash
├── Seq Scan: users (18,340 rows)          ⚠️  Full scan — nhỏ nên OK
└── Hash
    └── Bitmap Heap Scan: orders           ✅ Index được dùng
        ├── Recheck Cond: user_id = ANY(…)
        └── BitmapOr
            ├── Bitmap Index Scan: idx_orders_user_status
            └── Bitmap Index Scan: idx_orders_user_created

*Hình: Planner chọn Hash Join vì users table nhỏ (18K rows) fit RAM —
 nếu users table lớn hơn work_mem, plan sẽ degrade sang Nested Loop.*
```

### Index B-tree Visualization

```
                    [Root: 500]
                   /            \
          [250]                   [750]
         /     \                 /     \
     [100]    [400]          [600]    [900]
    /    \   /    \         /    \   /    \
 [50] [150][300][450]   [550][650][800][950]
  ↑         ↑                  ↑
leaf pages — actual data pointers

Dead tuples tích lũy ở leaf level:
[50][DEAD][150][DEAD][300]... → bloat mà VACUUM FULL mới fix được

*Hình: Index bloat xảy ra ở leaf pages — free space không được tái dụng
 ngay cả sau VACUUM thường, chỉ VACUUM FULL hoặc REINDEX mới compact lại.*
```

### MVCC / Dead Tuple Visualization

```
Timeline:
T1: INSERT row (xmin=100, xmax=null)    → VISIBLE (live tuple)
T2: UPDATE row (new: xmin=200, xmax=null, old: xmin=100, xmax=200)
T3: DELETE row (xmin=200, xmax=300)

Snapshot của transaction xid=250:
  Row xmin=100 xmax=200 → xmax < 250 → NOT VISIBLE (dead)
  Row xmin=200 xmax=300 → xmax > 250 → VISIBLE

VACUUM chạy sau T3:
  Row xmin=100 xmax=200 → dead, safe to reclaim ✅
  Row xmin=200 xmax=300 → xmax=300 > oldest_xmin? → depends

*Hình: Dead tuples chỉ được reclaim khi không có transaction nào cần
 snapshot cũ hơn — đây là lý do long-running transactions block VACUUM.*
```

### Replication Architecture

```
                    ┌─────────────────────┐
                    │   Primary (RW)       │
                    │   host: db-primary   │
                    │   port: 5432         │
                    └──────────┬──────────┘
                               │ WAL stream
                  ┌────────────┴────────────┐
                  │                         │
        ┌─────────▼─────────┐    ┌─────────▼─────────┐
        │  Replica 1 (RO)   │    │  Replica 2 (RO)   │
        │  Sync replication │    │  Async replication │
        │  host: db-rep1    │    │  host: db-rep2     │
        │  lag: < 1ms       │    │  lag: < 5s         │
        └───────────────────┘    └───────────────────┘
                  │                         │
                  └──────────┬──────────────┘
                             │
                   ┌─────────▼─────────┐
                   │    PgBouncer      │
                   │  Read load balance│
                   └───────────────────┘

*Hình: Sync replica đảm bảo zero data loss cho writes (Primary đợi ack)
 nhưng tăng write latency. Async replica có thể lag nhưng không block writes.*
```

---

## CODE — DB-Specific Example Patterns

### SQL Example Nhịp 4 Bước

```markdown
### Ví dụ 2: Intermediate — Cursor-based Pagination thay thế OFFSET

[INTRODUCE]:
Application có API lấy orders với pagination. OFFSET đang dùng:

```sql
SELECT * FROM orders ORDER BY created_at DESC OFFSET 10000 LIMIT 20;
```

Page 1 response: 12ms. Page 500 response: 4.2s. Page 2000: timeout.
Vấn đề không phải code — đó là cách OFFSET hoạt động trong PostgreSQL:
planner phải đọc và discard 10,000 rows trước khi trả 20 rows bạn cần.
Với 100M-row table, OFFSET 5,000,000 = đọc và bỏ 5M rows.

Keyset pagination giải quyết bài này: dùng WHERE clause thay vì OFFSET,
tận dụng index để jump trực tiếp đến vị trí cần.

Prerequisite: index trên `(created_at DESC, id DESC)`

[CODE]:
```sql
-- orders.sql — Pagination: Keyset cursor-based
-- Schema: orders(id BIGSERIAL, user_id, created_at TIMESTAMPTZ, ...)
-- Index: CREATE INDEX idx_orders_cursor ON orders(created_at DESC, id DESC);

-- First page: không cần cursor
SELECT id, user_id, created_at, total_amount
FROM orders
ORDER BY created_at DESC, id DESC  -- ⚠️ id để break tie khi cùng timestamp
LIMIT 20;
-- → Trả về last_cursor = (created_at=..., id=...) của row cuối

-- Next pages: dùng cursor từ response trước
SELECT id, user_id, created_at, total_amount
FROM orders
WHERE (created_at, id) < ($1, $2)  -- ✅ row value comparison, dùng index
ORDER BY created_at DESC, id DESC
LIMIT 20;
-- → O(log n) dù page 10000, response time stable
```

[TẠI SAO?]:
> **Tại sao `(created_at, id) < ($1, $2)` work với index?**
> PostgreSQL hiểu row value comparison và map nó thành index scan:
> "tìm entry đầu tiên nhỏ hơn tuple (ts, id), đọc tiếp 20 entries".
> Với B-tree index, đây là binary search O(log n) — không scan từ đầu.
> 
> **Tại sao cần `id` trong cursor?**
> Nếu 2 orders có cùng `created_at` (timestamp collision), chỉ dùng
> timestamp sẽ bỏ sót hoặc duplicate rows giữa pages.
> `id` là tiebreaker — kết hợp `(created_at, id)` là unique cursor.

[KẾT LUẬN]:
> Keyset pagination: response time O(log n) bất kể page number.
> 
> **Caveat**: Không support random access ("nhảy thẳng đến page 500") —
> chỉ next/prev. Nếu UI cần số trang, xem hybrid approach: COUNT estimate
> từ pg_class cho total + keyset cho actual data.
>
> **Dùng khi**: Infinite scroll, "Load more", API cursor pagination.
> **Không dùng khi**: UI cần pagination numbered với jump-to-page.
```

### Diagnostic Query Example Pattern

```markdown
### Ví dụ: Identify Missing Indexes — Production Diagnostic

[INTRODUCE]:
Table lớn nhưng không rõ index nào đang miss. Seq scan trên table
lớn = immediate performance hit. Query này identify candidates.

[CODE]:
```sql
-- pg_diagnostics.sql — Missing Index Detection
-- Chạy trên database đã có production traffic (pg_stat_user_tables cần warm)

SELECT
    schemaname || '.' || tablename AS table,
    seq_scan,                          -- số lần full table scan
    idx_scan,                          -- số lần index scan
    ROUND(
        seq_scan * 100.0 / NULLIF(seq_scan + idx_scan, 0), 2
    ) AS seq_scan_pct,                 -- % queries dùng seq scan
    n_live_tup AS estimated_rows,
    pg_size_pretty(pg_relation_size(
        (schemaname||'.'||tablename)::regclass
    )) AS table_size
FROM pg_stat_user_tables
WHERE n_live_tup > 100000             -- ✅ chỉ xét tables đủ lớn
  AND seq_scan > idx_scan             -- ❌ seq scan nhiều hơn index scan
  AND seq_scan > 1000                 -- ignore tables ít được access
ORDER BY seq_scan * n_live_tup DESC;  -- prioritize: scan nhiều × rows nhiều
```
> **Output**: Sorted by impact = số lần scan × rows per scan.
> Table đầu tiên = highest priority để add index.

[TẠI SAO?]:
> Tại sao `seq_scan * n_live_tup`?
> Seq scan 1000 lần trên table 100 rows = 100K work units.
> Seq scan 10 lần trên table 10M rows = 100M work units.
> Nhân 2 metric cho proxy của total I/O cost — không phải perfect
> nhưng useful để prioritize.

[KẾT LUẬN]:
> Đây là starting point, không phải verdict.
> Seq scan không phải luôn luôn sai — small table, low selectivity,
> bulk operations đều có thể prefer seq scan.
> Với mỗi table trong output: chạy EXPLAIN trên top queries để verify.
```

---

## PITFALLS — DB-Specific Fatal Pattern

### Template cho DB Fatal Pitfall

```markdown
### 🔴 Pitfall #N — [Mô tả hậu quả, không phải kỹ thuật]

[Setup: tại sao cái này nguy hiểm hơn những cái khác]

Config trông ổn này trên replicas:

```sql
-- hoặc config, hoặc query pattern
```

[Giải thích cơ chế tại sao nó fail]:
Vấn đề không xuất hiện ngay — nó tích lũy. Sau [thời gian/điều kiện]:
[failure mode cụ thể]

[Bối cảnh production] — điều này thường chỉ xuất hiện khi:
- [condition 1]
- [condition 2]

**Detection**:
```sql
-- Query để detect vấn đề này
```

**Fix**: [Cách đúng, với lý do kỹ thuật]
**Prevention**: [Monitoring query hoặc config để không bị lại]
```

---

## RECOMMEND — DB Topic Navigation

DB topics có hierarchy rõ ràng — leverage điều này trong RECOMMEND:

```
Sau bài về Query Optimization:
  → Index Strategy (tại sao indexes giải quyết vấn đề trên)
  → EXPLAIN Deep Dive (đọc plan để verify fix)
  → pg_stat_statements (monitor queries theo thời gian)

Sau bài về VACUUM / Bloat:
  → Autovacuum Tuning (tự động hóa maintenance)
  → Table Partitioning (tránh bloat bằng cách drop old partitions)
  → pg_repack (fix bloat không cần downtime)

Sau bài về Replication:
  → HA Setup (Patroni, repmgr)
  → Connection Pooling (PgBouncer với replica routing)
  → Backup Strategy (WAL archiving + base backup)
```

RECOMMEND narrative template cho DB:

```markdown
## 6. RECOMMEND

Bạn vừa hiểu [concept] — [1 câu về giá trị của nó trong production].

Nhưng [concept] chỉ là một phần của story. [Limitation tự nhiên dẫn đến next topic]:
[Next concept] giải quyết phần đó. Và để monitor hiệu quả của cả hai,
[monitoring topic] cung cấp visibility bạn cần.

| Mở rộng | Khi nào | Lý do | File/Link |
|---------|---------|-------|-----------|
```
