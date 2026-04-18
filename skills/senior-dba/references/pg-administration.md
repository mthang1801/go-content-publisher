# PostgreSQL Administration Reference

## 1. VACUUM và Bloat Management

### Tại sao VACUUM tồn tại — MVCC fundamentals

PostgreSQL dùng MVCC (Multi-Version Concurrency Control): khi UPDATE/DELETE,
row cũ không bị xóa ngay — nó được đánh dấu "dead" và giữ lại để serve
transactions đang chạy với snapshot cũ.

```
Kết quả: table tích lũy dead tuples → bloat
VACUUM: thu hồi dead tuples, cập nhật visibility map
VACUUM FULL: rewrite table (lock!), trả space về OS
autovacuum: background process tự động chạy VACUUM
```

### Autovacuum tuning

```sql
-- Xem config hiện tại
SELECT name, setting, unit 
FROM pg_settings 
WHERE name LIKE '%autovacuum%'
ORDER BY name;

-- Khi nào autovacuum trigger:
-- n_dead_tup > autovacuum_vacuum_threshold + autovacuum_vacuum_scale_factor × reltuples
-- Default: 50 + 0.02 × table_size = 2% dead tuples

-- Vấn đề: table 100M rows → trigger ở 2M dead tuples (quá trễ!)
-- Fix: per-table override
ALTER TABLE orders SET (
  autovacuum_vacuum_scale_factor = 0.01,  -- 1% thay vì 2%
  autovacuum_vacuum_threshold = 1000,
  autovacuum_analyze_scale_factor = 0.005
);

-- Với very large tables (> 1B rows):
ALTER TABLE large_events SET (
  autovacuum_vacuum_scale_factor = 0.001,  -- 0.1%
  autovacuum_vacuum_cost_delay = 2,        -- ms giữa các vacuum "nap" (default 2ms)
  autovacuum_vacuum_cost_limit = 400       -- work per nap (default 200)
);
```

### Bloat detection và fix

```sql
-- Quick bloat check
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    ROUND(
        100 * pg_relation_size(schemaname||'.'||tablename) /
        NULLIF(pg_total_relation_size(schemaname||'.'||tablename), 0)
    ) AS table_pct,
    n_dead_tup,
    ROUND(100 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0)) AS dead_pct,
    last_autovacuum,
    last_autoanalyze
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Fix bloat cho table không có downtime:
-- Option 1: pg_repack (best — no long lock)
-- pg_repack -t orders -j4

-- Option 2: VACUUM FULL (full lock — production có downtime)
VACUUM FULL ANALYZE orders;

-- Option 3: Cluster (sort + rewrite, brief lock per chunk)
CLUSTER orders USING idx_orders_created_at;
ANALYZE orders;
```

### Transaction ID Wraparound — P0 risk

```sql
-- Monitor XID age — MUST alert khi > 1.5 billion
SELECT
    datname,
    age(datfrozenxid) AS xid_age,
    pg_size_pretty(pg_database_size(datname)) AS db_size
FROM pg_database
ORDER BY age(datfrozenxid) DESC;

-- Alert threshold: 1,500,000,000 (1.5B)
-- Emergency threshold: 2,000,000,000 (2B) — DB becomes READ ONLY

-- Fix: agressive autovacuum
ALTER TABLE bloated_table SET (autovacuum_freeze_max_age = 100000000);
-- Hoặc: manual VACUUM FREEZE (nếu cấp bách)
VACUUM FREEZE VERBOSE orders;
```

---

## 2. Connection Management

### Connection exhaustion — diagnosing

```sql
-- Current connection state
SELECT
    state,
    COUNT(*) AS count,
    MAX(NOW() - state_change) AS max_duration,
    MAX(NOW() - query_start) AS max_query_time
FROM pg_stat_activity
WHERE pid <> pg_backend_pid()
GROUP BY state
ORDER BY count DESC;

-- Idle connections (wasted)
SELECT pid, usename, application_name, client_addr, state_change
FROM pg_stat_activity
WHERE state = 'idle'
  AND NOW() - state_change > INTERVAL '5 minutes'
ORDER BY state_change;

-- Long running queries
SELECT pid, NOW() - query_start AS duration, state, query
FROM pg_stat_activity
WHERE state != 'idle'
  AND NOW() - query_start > INTERVAL '1 minute'
ORDER BY duration DESC;

-- Kill specific connection
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE pid <> pg_backend_pid()
  AND state = 'idle'
  AND NOW() - state_change > INTERVAL '10 minutes';
```

### PgBouncer configuration — Production standard

```ini
; pgbouncer.ini

[databases]
mydb = host=localhost port=5432 dbname=mydb

[pgbouncer]
pool_mode = transaction          ; session | transaction | statement
; transaction mode: phù hợp nhất cho web apps (KHÔNG dùng SET, LISTEN/NOTIFY)
; session mode: dùng khi cần server-side prepared statements, LISTEN

max_client_conn = 10000          ; max connections từ app
default_pool_size = 25           ; connections thực sự đến PostgreSQL
min_pool_size = 5
reserve_pool_size = 5            ; emergency pool khi pool_size đầy
max_db_connections = 100         ; hard limit per database

server_idle_timeout = 600        ; giải phóng idle server connections
client_idle_timeout = 0          ; giữ client connections (app manages)

; Query tuning
server_round_robin = 1           ; load balance giữa replicas
query_wait_timeout = 120         ; max wait khi pool full
```

### Deadlock detection và prevention

```sql
-- Monitor deadlocks
SELECT
    deadlocks,
    conflicts,
    checkpoints_timed,
    checkpoints_req
FROM pg_stat_bgwriter;

-- Recent lock waits (pg 14+)
SELECT
    pid,
    query_start,
    wait_event_type,
    wait_event,
    state,
    query
FROM pg_stat_activity
WHERE wait_event_type = 'Lock'
ORDER BY query_start;

-- Deadlock prevention principles:
-- 1. Lock tables/rows trong thứ tự nhất quán (alphabetical hoặc by PK)
-- 2. Giữ transactions ngắn nhất có thể
-- 3. Dùng SELECT ... FOR UPDATE NOWAIT hoặc SKIP LOCKED thay vì blocking
-- 4. Tránh lock escalation (UPDATE nhiều rows → explicit lock order)
```

---

## 3. Replication

### Replication lag monitoring

```sql
-- Trên Primary: check replication status
SELECT
    application_name,
    client_addr,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    write_lag,
    flush_lag,
    replay_lag,
    sync_state
FROM pg_stat_replication
ORDER BY replay_lag DESC NULLS LAST;

-- Trên Replica: check lag
SELECT
    now() - pg_last_xact_replay_timestamp() AS replication_delay,
    pg_is_in_recovery() AS is_replica,
    pg_last_wal_receive_lsn() AS receive_lsn,
    pg_last_wal_replay_lsn() AS replay_lsn;

-- Alert khi delay > 30s (typical threshold)
```

### Replication slots — dangerous nếu không monitor

```sql
-- Xem slots (unused slots = WAL accumulation = disk risk!)
SELECT
    slot_name,
    plugin,
    slot_type,
    active,
    pg_size_pretty(
        pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)
    ) AS retained_wal_size
FROM pg_replication_slots
ORDER BY retained_wal_size DESC;

-- NGUY HIỂM: slot inactive nhưng giữ WAL
-- Drop nếu consumer không reconnect
SELECT pg_drop_replication_slot('slot_name');

-- Config an toàn: giới hạn WAL giữ lại
-- max_slot_wal_keep_size = 10GB  -- PostgreSQL 13+
```

### Failover checklist (manual)

```bash
# Trên replica — promote
pg_ctl promote -D /var/lib/postgresql/data

# Hoặc dùng trigger file (old method)
touch /tmp/postgresql.trigger

# Verify promotion
psql -c "SELECT pg_is_in_recovery();"  -- phải return false

# Update app connection string trỏ về replica mới là primary
# Update các replica khác để follow new primary
# pg_rewind để reconfigure old primary khi khôi phục
```

---

## 4. Backup và Recovery

### pg_basebackup — Physical backup

```bash
# Full backup
pg_basebackup \
  -h localhost \
  -D /backup/base \
  -P -Xs -R \
  --checkpoint=fast

# Options:
# -Xs: stream WAL trong quá trình backup
# -R: tạo recovery.conf / standby.signal cho replica setup
# -P: progress
# --checkpoint=fast: forced checkpoint (faster start)

# Verify backup
pg_basebackup --check-integrity /backup/base
```

### Backup strategy — Production minimum

```
Daily: pg_basebackup full backup → offsite storage
Continuous: WAL archiving → pg_wal_archive hoặc pgBackRest/WAL-G

Retention:
  - Daily backups: 7-14 ngày
  - Weekly: 4 tuần
  - Monthly: 12 tháng (compliance)

RPO target:
  - Tier 1 (critical): < 1 minute (streaming + WAL archiving)
  - Tier 2 (important): < 1 hour (WAL archiving mỗi 5-15 phút)
  - Tier 3 (standard): < 24 hours (daily backup)
```

### Point-in-Time Recovery (PITR)

```bash
# 1. Restore base backup
rsync -av /backup/base/ /var/lib/postgresql/data/

# 2. Config recovery
cat >> /var/lib/postgresql/data/postgresql.conf << EOF
restore_command = 'cp /backup/wal/%f %p'
recovery_target_time = '2024-01-15 14:30:00 UTC'
recovery_target_action = 'promote'
EOF

# 3. Tạo signal file
touch /var/lib/postgresql/data/recovery.signal

# 4. Start PostgreSQL — sẽ replay WAL đến target time
pg_ctl start -D /var/lib/postgresql/data

# 5. Verify recovery
psql -c "SELECT NOW();"  -- check current time
psql -c "SELECT pg_is_in_recovery();"  -- false = recovery complete
```

---

## 5. PostgreSQL Configuration — Production Tuning

### postgresql.conf — Critical settings

```ini
# Memory
shared_buffers = 25% of RAM       # 8GB cho 32GB server
effective_cache_size = 75% of RAM  # 24GB — hint cho planner, không allocate
work_mem = 64MB                    # per sort/hash operation
maintenance_work_mem = 1GB         # VACUUM, CREATE INDEX

# WAL
wal_level = replica                # minimum cho replication
max_wal_size = 4GB                 # default 1GB — tăng để reduce checkpoints
min_wal_size = 1GB
checkpoint_completion_target = 0.9 # spread checkpoint I/O
wal_compression = on               # zstd ở PG15+, lz4 ở PG13+

# Connections
max_connections = 200              # keep low, dùng PgBouncer
superuser_reserved_connections = 5

# Parallelism
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
max_parallel_maintenance_workers = 4

# Query planner
random_page_cost = 1.1  # SSD: 1.1-2.0, HDD: 4.0 (default)
effective_io_concurrency = 200  # SSD: 100-200, HDD: 2-4
jit = on  # Có lợi cho analytical queries, hại cho OLTP nhỏ

# Logging (minimal for monitoring)
log_min_duration_statement = 1000  # log queries > 1s
log_checkpoints = on
log_connections = off  # noise khi dùng pgbouncer
log_lock_waits = on
log_temp_files = 0  # log tất cả temp files
deadlock_timeout = 1s

# Autovacuum
autovacuum_max_workers = 5         # default 3, tăng nếu nhiều tables
autovacuum_vacuum_cost_delay = 2ms # default 2ms
autovacuum_vacuum_cost_limit = 400 # default 200, tăng để vacuum nhanh hơn
```

---

## 6. Partitioning

### Khi nào partition — Principal decision

```
Nên partition khi:
  ✅ Table > 100GB và tiếp tục grow
  ✅ Query thường xuyên filter theo partition key (time-series, tenant_id)
  ✅ Cần drop old data nhanh (DETACH PARTITION = O(1))
  ✅ Index size không vừa RAM (partition index nhỏ hơn, fit cache better)

Không nên partition khi:
  ❌ Table < 50GB — overhead không worth it
  ❌ Query không filter theo partition key (full scan vẫn xảy ra)
  ❌ Nhiều cross-partition JOINs
  ❌ Team chưa có experience quản lý partitioned tables
```

### Range partition — Time-series pattern

```sql
-- Create partitioned table
CREATE TABLE events (
    id BIGSERIAL,
    user_id BIGINT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    data JSONB
) PARTITION BY RANGE (created_at);

-- Create partitions
CREATE TABLE events_2024_01 PARTITION OF events
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE events_2024_02 PARTITION OF events
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Index per partition (automatic khi tạo index trên parent)
CREATE INDEX ON events (user_id, created_at);

-- Detach old partition (O(1), no scan)
ALTER TABLE events DETACH PARTITION events_2024_01 CONCURRENTLY;
-- Sau đó DROP hoặc archive
DROP TABLE events_2024_01;

-- Automation: pg_partman extension
```

### Partition pruning verification

```sql
-- Check partition pruning đang hoạt động
EXPLAIN SELECT * FROM events 
WHERE created_at BETWEEN '2024-01-01' AND '2024-01-31';

-- Phải thấy: "Partitions selected: 1" hoặc "Append" với chỉ 1 child
-- Nếu scan tất cả partitions → partition key không match hoặc cast issue
SET enable_partition_pruning = on;  -- verify nó đang on
```
