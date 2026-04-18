# Code Documentation Integration

Áp dụng workflow R1–R7 + skill `technical-doc-writer` cho code/programming topics.

---

## Style Mapping

```
Language concept (goroutine, generics, async/await) → Concept-First
Algorithm/Pattern (concurrent pipeline, saga, retry) → Concept-First
Problem-solving (N+1, memory leak, goroutine leak) → Problem-Centric
Language comparison (Go vs TS concurrency) → Concept-First với heavy comparison tables
DSA trong TS/Go context → Problem-Centric
```

---

## DEFINE — Tension Patterns cho Code Topics

### Pattern: Bug-in-Production Opening (strongest)

```
Template:
"[Service] của bạn deploy xong bình thường. Sau [time],
alert: [metric bất thường — không phải vague mà cụ thể].
Không có code thay đổi. Không có deploy mới.
[pprof/heap/memory profiler] reveal: [root cause].
Đây là [concept/pattern] — và nó xảy ra trong mọi [language] service
một khi traffic vượt [threshold]."

Ví dụ (sync.Pool / object reuse):
"Order service của bạn handle được 5,000 RPS tốt trong staging.
Production launch: 15,000 RPS. GC pauses bắt đầu xuất hiện mỗi 2 giây.
CPU không cao. Latency spike: P99 = 450ms thay vì 20ms.
pprof CPU profile reveal: 40% CPU time spent in garbage collector.
Culprit: mỗi request allocate một []byte buffer 4KB, dùng xong bỏ.
15,000 requests/second = 15,000 allocations/second = GC overwhelmed.
sync.Pool giải quyết bài này bằng cách tái sử dụng buffers thay vì allocate mới."
```

### Pattern: "Trông đúng, thực ra sai" Opening

```
Template:
"Code này trông hoàn toàn đúng:

[code snippet — 5-10 dòng, có bug ẩn]

Senior dev review? Pass. Unit test? Green. Code review? Approved.
Production với [high concurrency / large dataset / edge case]: [failure].
Đây là [concept] — behavior mà mọi developer assume nhưng [language runtime] không guarantee."

Ví dụ (goroutine loop variable capture):
"Code này compile, chạy, thậm chí test green trong 99% trường hợp:

for _, item := range items {
    go func() {
        process(item) // bug here — pre Go 1.22
    }()
}

Review comment: 'looks good'. Integration test: passes.
Production với 100+ items và high CPU load: output corrupt, items processed multiple times.
Đây là loop variable capture — một trong những gotchas phổ biến nhất trong Go."
```

### Pattern: Performance Cliff Opening

```
Template:
"Query này return trong 15ms với 10,000 rows trong database:
[code snippet]

6 tháng sau, data grow lên 5 triệu rows. Same query: 8 giây.
Không có gì thay đổi — chỉ có data volume.
Đây là [N+1 / full table scan / missing index / etc.] — behavior
mà không hiển nhiên cho đến khi scale up."
```

---

## VISUAL — Code Topic Diagrams

### Memory / Object Lifecycle

```
Go GC Pressure — Without sync.Pool:

Request 1 ──► Allocate Buffer (4KB) ──► Use ──► GC collects
Request 2 ──► Allocate Buffer (4KB) ──► Use ──► GC collects
...
Request N ──► Allocate Buffer (4KB) ──► Use ──► GC collects
                                                      │
                                              GC pause: 50-200ms
                                              All requests stall

Go GC Pressure — With sync.Pool:

Request 1 ──► Pool.Get() ──► Use ──► Pool.Put()
Request 2 ──► Pool.Get() (reuse!) ──► Use ──► Pool.Put()
...
Request N ──► Pool.Get() (reuse!) ──► Use ──► Pool.Put()
                                              │
                                        GC sees few live objects
                                        GC pause: < 1ms

*Hình: sync.Pool break allocate-use-GC cycle — GC chỉ collect objects
 không được reuse, dramatically giảm GC pressure tại scale.*
```

### Concurrency Flow

```
TypeScript Event Loop (single-threaded):

   Task Queue: [req1] [req2] [req3]
         │
    ┌────▼──────────────────────┐
    │       Event Loop          │
    │  ┌──────────────────────┐ │
    │  │ Execute req1          │ │
    │  │ await db.query() ─── ┼─┼──► I/O Thread Pool (libuv)
    │  │   [yield to loop]    │ │            │ (non-blocking)
    │  └──────────────────────┘ │            │
    │  ┌──────────────────────┐ │            │
    │  │ Execute req2          │ │◄────── DB result (req1)
    │  └──────────────────────┘ │
    └──────────────────────────┘

Go Goroutines (M:N, true parallel):

    goroutine 1 (req1) ──► OS Thread 1 ──► CPU Core 1
    goroutine 2 (req2) ──► OS Thread 2 ──► CPU Core 2
    goroutine 3 (req3) ──► OS Thread 1 (scheduled)
    goroutine 4 (req4) ──► OS Thread 2 (scheduled)
    ... 10,000 goroutines on 4 OS threads

*Hình: TS event loop: 1 CPU core, I/O async. Go goroutines: N CPU cores,
 cả CPU và I/O đều parallel. CPU-bound work: Go wins significantly.*
```

### Error Flow

```
TypeScript — Exception propagation (implicit):

    handler()
        └── service()          [throws UserNotFoundException]
                └── repo()     [throws DatabaseError]
                        └── db.query() → Error

    Without proper catch: UnhandledPromiseRejection → crash
    With catch: must remember to catch at every async boundary

Go — Explicit error propagation:

    handler()          err != nil → return 500
        └── service()  err != nil → return fmt.Errorf("service: %w", err)
                └── repo()   err != nil → return fmt.Errorf("repo: %w", err)
                        └── db.Query() → (rows, err)

    Error context accumulates: "service: repo: connection timeout"
    Impossible to accidentally ignore error (compiler warns on unused return)

*Hình: Go errors phải được explicitly propagated — compiler không cho phép
 bỏ qua return value. TypeScript exceptions có thể bị swallowed.*
```

---

## CODE Section — Multi-language đúng nhịp 4 bước

### Pre-context pattern cho code topics

```markdown
### Ví dụ 2: Intermediate — Context Propagation cho Request Cancellation

[INTRODUCE]:
User click search, kết quả mất 3 giây. Trong khi chờ, user click cancel.
Frontend abort request. Backend không biết — vẫn tiếp tục query DB, gọi external APIs,
consume resources cho một request không ai còn cần.

Tại scale 10K concurrent users, 5% cancel rate = 500 "zombie" requests running at all times.
500 wasted DB connections, 500 wasted external API calls per second.

Context propagation giải quyết bài này: khi client disconnect,
context bị cancel, tất cả downstream operations automatically stop.
```

### "Tại sao?" patterns cho code

```markdown
> **Tại sao Go context phải là first parameter, không phải field của struct?**
>
> Context là request-scoped — mỗi request có lifetime khác nhau, cancellation khác nhau.
> Nếu là struct field: service instance bị tie vào một context, không thể reuse.
> First parameter: function signature explicit rằng "operation này có thể bị cancel".
> Convention này enforce correct usage — không thể "forget" context.

> **Tại sao AbortController thay vì global cancel flag trong TypeScript?**
>
> Global flag: tất cả operations cùng service bị ảnh hưởng.
> AbortController: per-request, composable, tích hợp với fetch API.
> AbortSignal có thể passed xuống, combined (anySignal), và Web Platform standard.
```

---

## PITFALLS — Code-Specific Fatal Patterns

### Template cho concurrent programming pitfall

```markdown
### 🔴 Pitfall #N — [Mô tả symptom, không phải technical cause]

Code này chạy tốt trong development và staging:

```go
// hoặc TypeScript
```

Trong production với [concurrent users / large dataset / sustained load]:
[Failure mode với numbers]

Đây không phải race condition theo định nghĩa OS — đây là [specific behavior]:
[Technical explanation tại sao failure xảy ra]

**Detection**:
```bash
# Go: race detector
go test -race ./...
# hoặc runtime flag
go run -race main.go
```

**Fix**: [Code cụ thể với explanation]
**Prevention**: [Design pattern hoặc linting rule]
```

---

## RECOMMEND — Code Topic Navigation

### Learning path patterns

```
Sau bài về Go Goroutines:
  → Channels (communication giữa goroutines)
  → sync package (coordination primitives)
  → Context (cancellation propagation)
  → errgroup (error handling với goroutines)
  → pprof (profiling goroutine leaks)

Sau bài về TypeScript Generics:
  → Conditional Types (type-level programming)
  → Template Literal Types (string manipulation)
  → Mapped Types (transformation)
  → Zod (runtime validation matching static types)
  → tRPC (end-to-end type safety)

Sau bài về Error Handling (either language):
  → Testing error cases (table-driven tests)
  → Observability (logging errors với context)
  → Circuit breaker (handle downstream errors)
  → Result types (railway-oriented programming)
```

### RECOMMEND narrative template cho code

```markdown
## 6. RECOMMEND

Bạn vừa hiểu [concept] — [1 câu về tại sao nó matter trong production].

Nhưng [concept] không đứng một mình. Trong real-world systems,
nó thường được kết hợp với [related pattern] để solve [higher-level problem].
Và để observe behavior của nó trong production, [observability tool] là
thứ bạn cần biết tiếp theo.

| Mở rộng | Khi nào | Lý do | File/Link |
|---------|---------|-------|-----------|
```
