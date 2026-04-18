# PostgreSQL Diagnostics Playbook

## Quick Health Check — Chạy ngay khi có vấn đề

```sql
-- 1. Database size và activity snapshot
SELECT
    datname,
    numbackends AS connections,
    xact_commit,
    xact_rollback,
    blks_hit,
    blks_read,
    ROUND(blks_hit * 100.0 / NULLIF(blks_hit + blks_read, 0), 2) AS cache_hit_ratio,
    deadlocks,
    pg_size_pretty(pg_database_size(datname)) AS size
FROM pg_stat_database
WHERE datname = current_database();

-- Cache hit ratio: phải > 99% cho OLTP
-- Nếu < 95%: shared_buffers quá thấp hoặc working set không fit

-- 2. Current activity
SELECT
    pid,
    state,
    wait_event_type,
    wait_event,
    NOW() - query_start AS duration,
    LEFT(query, 100) AS query_preview
FROM pg_stat_activity
WHERE state != 'idle'
  AND pid <> pg_backend_pid()
ORDER BY duration DESC NULLS LAST
LIMIT 20;

-- 3. Lock contention
SELECT
    pid,
    locktype,
    relation::regclass,
    mode,
    granted,
    query
FROM pg_locks l
JOIN pg_stat_activity a USING (pid)
WHERE NOT granted
ORDER BY pid;
```

---

## Playbook A: Query Chậm — End-to-End Investigation

### Step 1: Identify slow queries

```sql
-- pg_stat_statements phải được enable
-- shared_preload_libraries = 'pg_stat_statements'

SELECT
    LEFT(query, 150) AS query,
    calls,
    ROUND(mean_exec_time::numeric, 2) AS mean_ms,
    ROUND(total_exec_time::numeric / 1000, 2) AS total_sec,
    ROUND(stddev_exec_time::numeric, 2) AS stddev_ms,
    rows,
    ROUND(rows / calls) AS rows_per_call,
    ROUND(100 * shared_blks_hit /
        NULLIF(shared_blks_hit + shared_blks_read, 0), 2) AS cache_hit_ratio
FROM pg_stat_statements
WHERE calls > 10
ORDER BY total_exec_time DESC
LIMIT 20;
```

### Step 2: Get execution plan

```sql
-- Template cho phân tích plan
EXPLAIN (
    ANALYZE,
    BUFFERS,
    FORMAT TEXT,
    SETTINGS,    -- show non-default settings ảnh hưởng plan (PG12+)
    WAL          -- show WAL generation (PG13+)
)
[YOUR QUERY HERE];
```

### Step 3: Analyze plan — red flags checklist

```
□ Seq Scan trên table lớn với high-selectivity filter → thiếu index
□ rows estimate vs actual lệch > 10x → stale statistics (ANALYZE)
□ Nested Loop với outer rows > 1000 và inner cost cao → thiếu index trên join col
□ Hash Join: Batches > 1 → work_mem quá thấp
□ Sort: external merge → work_mem quá thấp hoặc cần index
□ Rows Removed by Filter cao → index không selective đủ
□ BitmapHeapScan với high "Rows Removed by Index Recheck" → lossy index
□ Parallel workers ít hơn expected → max_parallel_workers_per_gather
```

### Step 4: Fix và verify

```sql
-- Sau khi thêm index:
EXPLAIN (ANALYZE, BUFFERS) [query];
-- Compare: actual time, shared read vs hit, rows estimate accuracy

-- Force plan để test (development only):
SET enable_seqscan = off;
EXPLAIN (ANALYZE, BUFFERS) [query];
RESET enable_seqscan;
```

---

## Playbook B: High CPU / Runaway Query

```sql
-- 1. Identify culprit process
SELECT pid, NOW() - pg_stat_activity.query_start AS duration,
       query, state, wait_event
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC;

-- 2. Get full query của pid
SELECT query FROM pg_stat_activity WHERE pid = [PID];

-- 3. Check nếu là legit query hay runaway
-- Duration > 10 phút cho OLTP query → likely runaway

-- 4. Cancel gracefully (không kill connection)
SELECT pg_cancel_backend([PID]);

-- 5. Kill nếu cancel không ăn
SELECT pg_terminate_backend([PID]);

-- 6. Root cause: tại sao query này chạy lâu?
-- - Missing index → seq scan
-- - Bad plan do stale statistics
-- - Lock wait (check wait_event = 'relation' hoặc 'tuple')
-- - Cartesian product (missing JOIN condition)
```

---

## Playbook C: Disk Full / Bloat Crisis

```sql
-- 1. Check disk usage breakdown
SELECT
    pg_size_pretty(pg_database_size(current_database())) AS db_size,
    pg_size_pretty(SUM(pg_total_relation_size(relid))) AS total_tables_size
FROM pg_catalog.pg_statio_user_tables;

-- 2. Top tables by size
SELECT
    schemaname || '.' || tablename AS table,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table,
    pg_size_pretty(
        pg_total_relation_size(schemaname||'.'||tablename) -
        pg_relation_size(schemaname||'.'||tablename)
    ) AS indexes_and_toast,
    n_dead_tup,
    n_live_tup
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;

-- 3. Check WAL size
SELECT pg_size_pretty(SUM(size)) AS wal_size
FROM pg_ls_waldir();

-- 4. Check replication slots (WAL retention)
SELECT slot_name, active,
    pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn))
FROM pg_replication_slots;

-- 5. Emergency: delete inactive slots (nếu consumer gone)
SELECT pg_drop_replication_slot('slot_name');

-- 6. Emergency vacuum để reclaim space
VACUUM (VERBOSE, ANALYZE) [table_name];
-- VACUUM FULL nếu cần trả space về OS (cần lock!)
```

---

## Playbook D: Replication Lag Crisis

```sql
-- 1. Check lag từ primary
SELECT
    application_name,
    client_addr,
    state,
    write_lag,
    flush_lag,
    replay_lag,
    sync_state
FROM pg_stat_replication;

-- 2. Từ replica: xác nhận lag
SELECT
    now() - pg_last_xact_replay_timestamp() AS lag,
    pg_last_wal_receive_lsn(),
    pg_last_wal_replay_lsn(),
    pg_is_wal_receiver_running() AS wal_receiver_running;

-- 3. Common causes:
-- a) Network bandwidth saturation → check network I/O
-- b) Replica disk I/O bottleneck → check iostat
-- c) Long-running query on replica blocking apply (hot_standby_feedback)
--    SELECT pid, query, NOW()-query_start FROM pg_stat_activity WHERE state!='idle';
-- d) Vacuum on primary generating too much WAL (index rebuilds)

-- 4. Temporary fix: throttle vacuums on primary
ALTER SYSTEM SET autovacuum_vacuum_cost_delay = 20;  -- slow down vacuum
SELECT pg_reload_conf();
-- Revert sau khi replica catch up!

-- 5. Check wal_receiver process
SELECT * FROM pg_stat_wal_receiver;
```

---

## Playbook E: Deadlock Investigation

```sql
-- 1. Check deadlock count (tăng → vấn đề)
SELECT deadlocks FROM pg_stat_database WHERE datname = current_database();

-- 2. Enable deadlock logging (postgresql.conf)
-- deadlock_timeout = 1s
-- log_lock_waits = on
-- Deadlock details sẽ xuất hiện trong pg_log

-- 3. Current locks
SELECT
    locktype,
    relation::regclass AS table,
    mode,
    granted,
    pid,
    LEFT(query, 80) AS query
FROM pg_locks
JOIN pg_stat_activity USING (pid)
WHERE locktype = 'relation'
ORDER BY relation, granted DESC;

-- 4. Lock dependency graph
SELECT
    COALESCE(blockingl.relation::regclass::text, blockingl.locktype) AS locked_item,
    blockeda.pid AS blocked_pid,
    blockeda.query AS blocked_query,
    blockedl.mode AS blocked_mode,
    blockinga.pid AS blocking_pid,
    blockinga.query AS blocking_query,
    blockingl.mode AS blocking_mode
FROM pg_locks blockedl
JOIN pg_stat_activity blockeda ON blockedl.pid = blockeda.pid
JOIN pg_locks blockingl ON (
    blockingl.granted = true
    AND blockedl.relation = blockingl.relation
    AND blockedl.locktype = blockingl.locktype
    AND blockedl.pid != blockingl.pid
)
JOIN pg_stat_activity blockinga ON blockingl.pid = blockinga.pid
WHERE blockedl.granted = false;
```

---

## Monitoring Queries — Production Dashboard

```sql
-- ============================================================
-- DASHBOARD: Chạy mỗi 30 giây cho real-time monitoring
-- ============================================================

-- TPS (transactions per second)
SELECT
    ROUND(xact_commit / EXTRACT(EPOCH FROM (NOW() - stats_reset))) AS tps_commit,
    ROUND(xact_rollback / EXTRACT(EPOCH FROM (NOW() - stats_reset))) AS tps_rollback,
    stats_reset
FROM pg_stat_database
WHERE datname = current_database();

-- Active queries
SELECT COUNT(*) FILTER (WHERE state = 'active') AS active,
       COUNT(*) FILTER (WHERE state = 'idle') AS idle,
       COUNT(*) FILTER (WHERE state = 'idle in transaction') AS idle_in_txn,
       COUNT(*) FILTER (WHERE wait_event_type = 'Lock') AS waiting_on_lock
FROM pg_stat_activity;

-- Table hit rate
SELECT
    schemaname,
    tablename,
    heap_blks_hit,
    heap_blks_read,
    ROUND(heap_blks_hit * 100.0 / NULLIF(heap_blks_hit + heap_blks_read, 0), 2) AS cache_hit_pct
FROM pg_statio_user_tables
WHERE heap_blks_hit + heap_blks_read > 0
ORDER BY heap_blks_read DESC
LIMIT 10;

-- Index efficiency
SELECT
    schemaname,
    tablename,
    seq_scan,
    idx_scan,
    ROUND(idx_scan * 100.0 / NULLIF(seq_scan + idx_scan, 0), 2) AS index_scan_pct,
    n_live_tup
FROM pg_stat_user_tables
WHERE n_live_tup > 10000
  AND seq_scan > 100
ORDER BY seq_scan DESC
LIMIT 20;

-- Checkpoint performance
SELECT
    checkpoints_timed,
    checkpoints_req,
    ROUND(checkpoint_write_time / 1000.0, 2) AS write_time_sec,
    ROUND(checkpoint_sync_time / 1000.0, 2) AS sync_time_sec,
    buffers_checkpoint,
    buffers_clean,
    buffers_backend
FROM pg_stat_bgwriter;

-- checkpoints_req >> checkpoints_timed → max_wal_size quá nhỏ
-- buffers_backend > 0 → checkpoint không kịp, backends phải tự write dirty pages
```

---

## Schema Inspection Queries

```sql
-- Table columns với types
SELECT
    column_name,
    data_type,
    character_maximum_length,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'orders'
ORDER BY ordinal_position;

-- All indexes trên table
SELECT
    indexname,
    indexdef,
    pg_size_pretty(pg_relation_size(indexname::regclass)) AS size
FROM pg_indexes
WHERE tablename = 'orders';

-- Foreign keys
SELECT
    tc.constraint_name,
    kcu.column_name,
    ccu.table_name AS foreign_table,
    ccu.column_name AS foreign_column,
    rc.update_rule,
    rc.delete_rule
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu USING (constraint_name)
JOIN information_schema.constraint_column_usage ccu USING (constraint_name)
JOIN information_schema.referential_constraints rc USING (constraint_name)
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_name = 'orders';

-- Sequences (auto-increment exhaustion risk)
SELECT
    sequence_name,
    last_value,
    max_value,
    ROUND(100.0 * last_value / max_value, 2) AS pct_used
FROM information_schema.sequences s
JOIN (
    SELECT sequencename, last_value, max_value
    FROM pg_sequences
) seq ON seq.sequencename = s.sequence_name;
-- Alert khi pct_used > 80%!
```
