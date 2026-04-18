# Go Performance & Profiling — Production Guide

## Table of Contents
1. [pprof Profiling](#pprof)
2. [Benchmarking](#benchmarks)
3. [Escape Analysis](#escape)
4. [Memory Optimization](#memory)
5. [GC Tuning](#gc)
6. [Common Performance Bugs](#bugs)

---

## 1. pprof Profiling

### Always-On HTTP pprof Endpoint (internal only)

```go
import _ "net/http/pprof" // registers /debug/pprof handlers

// In main.go — on a SEPARATE internal port, never expose externally
go func() {
    slog.Info("pprof listening", "addr", ":6060")
    if err := http.ListenAndServe(":6060", nil); err != nil {
        slog.Error("pprof server failed", "err", err)
    }
}()
```

### Collecting Profiles

```bash
# CPU — what's consuming CPU cycles?
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# Heap — what's consuming memory?
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap

# Goroutine — how many, what are they doing?
curl http://localhost:6060/debug/pprof/goroutine?debug=2 > goroutines.txt

# Trace — timeline of goroutine activity (for latency investigation)
curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out
go tool trace trace.out

# Mutex contention
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### In-Code Profiling

```go
import "runtime/pprof"

// CPU profile a specific code path
func profileCPU(filename string, fn func()) error {
    f, err := os.Create(filename)
    if err != nil { return err }
    defer f.Close()

    if err := pprof.StartCPUProfile(f); err != nil { return err }
    defer pprof.StopCPUProfile()

    fn()
    return nil
}

// Heap snapshot
func snapshotHeap(filename string) error {
    f, err := os.Create(filename)
    if err != nil { return err }
    defer f.Close()
    runtime.GC() // get up-to-date picture
    return pprof.WriteHeapProfile(f)
}
```

---

## 2. Benchmarking — Measure Before You Optimize

```go
// benchmark_test.go
func BenchmarkOrderProcessing(b *testing.B) {
    svc := setupService()
    order := buildTestOrder()
    
    b.ResetTimer()          // exclude setup time
    b.ReportAllocs()        // show allocations per op

    for i := 0; i < b.N; i++ {
        if err := svc.Process(context.Background(), order); err != nil {
            b.Fatal(err)
        }
    }
}

// Run: go test -bench=BenchmarkOrderProcessing -benchmem -count=5 ./...
// Output:
// BenchmarkOrderProcessing-8   50000   24123 ns/op   4096 B/op   32 allocs/op

// Sub-benchmarks for comparison
func BenchmarkSerializer(b *testing.B) {
    order := buildTestOrder()
    b.Run("json", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, _ = json.Marshal(order)
        }
    })
    b.Run("protobuf", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, _ = proto.Marshal(toProto(order))
        }
    })
}
```

---

## 3. Escape Analysis — Heap vs Stack

```bash
# See what escapes to heap
go build -gcflags="-m -m" ./... 2>&1 | grep "escapes to heap"
```

**Stack is fast** (no GC pressure). **Heap is slower** (GC must collect it).

```go
// What causes heap allocation?

// 1. Interface assignment — concrete value boxed
var r io.Reader = &bytes.Buffer{} // Buffer escapes to heap

// 2. Returning pointer to local var (sometimes — compiler may optimize)
func newOrder() *Order {
    o := Order{} // may escape
    return &o
}

// 3. Slice/map that grows beyond initial size
s := make([]byte, 0, 1024) // stack if small enough

// 4. Closure capturing a variable
x := 42
go func() { fmt.Println(x) }() // x escapes to heap

// Optimization: pass large structs by pointer, small by value
func process(o Order) { ... }    // Order copied on stack (fast if small)
func process(o *Order) { ... }   // pointer — no copy, but must be on heap
```

---

## 4. Memory Optimization

### sync.Pool — Reuse Expensive Objects

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return &bytes.Buffer{}
    },
}

func encode(data interface{}) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer bufferPool.Put(buf)  // return to pool

    if err := json.NewEncoder(buf).Encode(data); err != nil {
        return nil, err
    }
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes()) // copy before returning buffer to pool
    return result, nil
}
```

### Pre-allocate Slices

```go
// BAD — repeated allocations as slice grows
var results []Order
for rows.Next() {
    results = append(results, scanOrder(rows))
}

// GOOD — pre-allocate if count is known
count := getCount(ctx)
results := make([]Order, 0, count)
for rows.Next() {
    results = append(results, scanOrder(rows))
}
```

### String Building

```go
// BAD — O(n²) allocations
s := ""
for _, part := range parts {
    s += part  // new string each iteration
}

// GOOD
var sb strings.Builder
sb.Grow(estimatedSize) // optional: pre-allocate
for _, part := range parts {
    sb.WriteString(part)
}
s := sb.String()
```

### Avoid Unnecessary Conversions

```go
// BAD — allocates new string
key := string(bytes)
if key == "expected" { ... }

// GOOD — no allocation
if string(bytes) == "expected" { ... } // compiler optimizes this!

// Or better with bytes.Equal
if bytes.Equal(b, []byte("expected")) { ... }
```

---

## 5. GC Tuning

```go
// GOGC — controls when GC runs (default: 100 = run when heap doubles)
// Lower = more frequent GC, less memory usage
// Higher = less frequent GC, more memory usage

// For latency-sensitive services: GOGC=off + manual tuneup
// For memory-constrained: GOGC=50

// GOMEMLIMIT (Go 1.19+) — soft memory limit
// GOMEMLIMIT=2GiB go run main.go
// Prevents OOM; GC runs more aggressively as limit approaches

import "runtime/debug"

// In high-throughput batch processing: pause GC during critical path
debug.SetGCPercent(-1)   // disable
processCriticalPath()
debug.SetGCPercent(100)  // restore

// Monitor GC in production
var stats runtime.MemStats
runtime.ReadMemStats(&stats)
slog.Info("gc stats",
    "heap_alloc_mb", stats.HeapAlloc/1024/1024,
    "num_gc", stats.NumGC,
    "pause_total_ms", stats.PauseTotalNs/1e6,
)
```

---

## 6. Common Performance Bugs

### Bug 1: JSON Unmarshal into interface{} (allocates map)

```go
// BAD — allocates map[string]interface{} on heap
var result interface{}
json.Unmarshal(data, &result)

// GOOD — unmarshal directly to typed struct
var result OrderResponse
json.Unmarshal(data, &result)
```

### Bug 2: Regex Compiled on Every Call

```go
// BAD — compiled 10k times/second
func isValidEmail(s string) bool {
    matched, _ := regexp.MatchString(`^[\w.]+@[\w.]+\.\w+$`, s)
    return matched
}

// GOOD — compile once
var emailRegex = regexp.MustCompile(`^[\w.]+@[\w.]+\.\w+$`)
func isValidEmail(s string) bool {
    return emailRegex.MatchString(s)
}
```

### Bug 3: Defer in Hot Loop

```go
// BAD — defer doesn't run until function returns; defers accumulate
func processAll(items []Item) {
    for _, item := range items {
        mu.Lock()
        defer mu.Unlock() // BAD: unlock deferred to end of processAll!
        process(item)
    }
}

// GOOD — inline unlock or use closure
func processAll(items []Item) {
    for _, item := range items {
        func() {
            mu.Lock()
            defer mu.Unlock() // unlocks at end of closure
            process(item)
        }()
    }
}
```

### Bug 4: Large Struct Copy in Range

```go
type LargeOrder struct {
    Items [1000]OrderItem // large array on stack
    // ... many fields
}

var orders []LargeOrder

// BAD — copies entire LargeOrder on each iteration
for _, order := range orders {
    process(order)
}

// GOOD — use index to avoid copy
for i := range orders {
    process(&orders[i])
}
```

### Bug 5: HTTP Client Not Reused

```go
// BAD — each call creates new transport, no connection reuse
func callAPI(url string) ([]byte, error) {
    resp, err := http.Get(url) // uses default client, but creates per-call overhead
    ...
}

// GOOD — shared client with proper timeouts
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

---

## Performance Checklist for Code Review

- [ ] No regex compiled inside functions — use `var` or `sync.Once`
- [ ] Pre-allocated slices/maps where size is known
- [ ] `strings.Builder` for string concatenation in loops
- [ ] `sync.Pool` for frequently allocated/discarded objects
- [ ] HTTP client reused (not created per request)
- [ ] DB connection pool sized appropriately
- [ ] `context.WithTimeout` on every external call
- [ ] No `time.Sleep` in hot paths — use `time.Timer` or ticker
- [ ] `defer` not inside tight loops
- [ ] Large structs passed by pointer, not value
- [ ] `bytes.Equal` instead of `string(b) == str` in loops
- [ ] Benchmarks exist for critical path functions
