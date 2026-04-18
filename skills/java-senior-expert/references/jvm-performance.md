# JVM Performance — Tuning & Incident Response

## GC Selection Guide

| Workload | Recommended GC | JVM Flag |
|---|---|---|
| Low-latency APIs (< 10ms p99) | ZGC | `-XX:+UseZGC` |
| Balanced throughput/latency | G1GC | `-XX:+UseG1GC` (default Java 9+) |
| Batch / throughput-heavy | Parallel GC | `-XX:+UseParallelGC` |
| Legacy Java 8 | CMS (deprecated) | → migrate to G1 |

---

## Essential JVM Flags for Production

```bash
# Heap sizing — never let -Xmx exceed 75% of container memory
-Xms2g -Xmx2g                          # same value = no resizing overhead

# GC
-XX:+UseG1GC
-XX:MaxGCPauseMillis=200               # target pause goal (not guaranteed)
-XX:G1HeapRegionSize=16m               # adjust for large heaps

# Observability — always on in production
-XX:+HeapDumpOnOutOfMemoryError
-XX:HeapDumpPath=/var/log/app/heapdump.hprof
-XX:+PrintGCDetails
-Xlog:gc*:file=/var/log/app/gc.log:time,uptime:filecount=5,filesize=20m

# JFR (Java Flight Recorder) — near-zero overhead profiling
-XX:StartFlightRecording=duration=60s,filename=/var/log/app/profile.jfr

# Container awareness (Java 11+)
-XX:+UseContainerSupport
-XX:MaxRAMPercentage=75.0              # use 75% of container RAM
```

---

## Diagnosing Memory Leaks

### Step 1: Collect evidence

```bash
# Heap histogram (live objects) — non-intrusive
jcmd <pid> GC.heap_info
jcmd <pid> GC.class_histogram | head -30

# Full heap dump — triggers full GC, use with caution in prod
jcmd <pid> GC.heap_dump /tmp/heap.hprof

# Or via kill signal (if HeapDumpOnOutOfMemoryError is set, this creates one on demand)
jmap -dump:format=b,file=/tmp/heap.hprof <pid>
```

### Step 2: Analyze with Eclipse MAT

Look for:
- **Leak Suspects Report** → MAT identifies objects retained by unexpected references
- **Dominator Tree** → what's consuming the most heap
- **Retained Heap** → how much would be freed if this object were GC'd

### Common Leak Patterns

| Symptom | Likely Cause |
|---|---|
| `byte[]` growing unboundedly | Response body accumulation, log buffer |
| `char[]` growing | String interning, cached strings |
| `HashMap$Entry` growing | Unbounded cache (no eviction) |
| `Thread` objects not decreasing | ExecutorService not shut down |
| `ClassLoader` retained | Hot-reload without cleanup |

---

## Diagnosing CPU Spikes

```bash
# 1. Find hot threads
jcmd <pid> Thread.print > threads.txt
# or
kill -3 <pid>  # thread dump to stdout/log

# 2. Get CPU usage per thread
top -H -p <pid>  # shows threads with %CPU
# Note the tid (decimal), convert to hex

# 3. Match to thread dump
printf '%x\n' <tid>  # e.g. 12345 → 0x3039
# Search threads.txt for nid=0x3039
```

### Async-profiler (best tool for CPU profiling)

```bash
# Profile for 30 seconds, generate flamegraph
./asprof -d 30 -f /tmp/flamegraph.html <pid>
```

---

## Connection Pool Tuning (HikariCP)

```yaml
spring:
  datasource:
    hikari:
      maximum-pool-size: 20        # start here; formula: (core_count * 2) + effective_spindle_count
      minimum-idle: 5
      connection-timeout: 3000     # ms — fail fast if no connection available
      idle-timeout: 600000         # 10 min
      max-lifetime: 1800000        # 30 min — rotate before DB kills idle connections
      leak-detection-threshold: 5000  # warn if connection held > 5s
```

**Warning signs of pool exhaustion:**
- Latency spike correlating with `HikariPool-1 - Connection is not available`
- Thread dump shows many threads blocked on `HikariPool.getConnection()`
- Fix: increase pool size, find slow queries, reduce transaction scope

---

## GC Pause Analysis

```
# In GC log, look for:
[GC pause (G1 Evacuation Pause) (young) 512M->256M(1024M), 0.3456789 secs]
                                                              ^^^^^^^^^^^^^ pauses app threads

# Red flags:
# - Pause > MaxGCPauseMillis frequently → heap pressure, tune region size
# - Full GC → heap exhaustion or fragmentation → increase heap or fix leak
# - Humongous allocations → objects > 50% of region size → increase region size
```

---

## Performance Checklist for Code Review

- [ ] No unbounded collections (Lists, Maps without size limit)
- [ ] Pagination on all DB queries returning multiple rows
- [ ] No N+1 queries (use JOIN FETCH or batch fetching)
- [ ] String concatenation in loops uses StringBuilder
- [ ] `instanceof` + cast chains replaced with polymorphism
- [ ] Regex patterns compiled once as `static final Pattern`
- [ ] DateTimeFormatter / NumberFormat stored as static (they're thread-safe)
- [ ] No `System.out.println` in production code
- [ ] Logging uses parameterized form: `log.debug("Order {}", id)` not string concat
- [ ] External calls have timeouts configured
- [ ] Large object creation inside tight loops (allocate once, reuse)
