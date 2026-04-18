# Go Concurrency Patterns — Production-Tested Recipes

## Mental Model First

Go concurrency = **CSP (Communicating Sequential Processes)**.
- Goroutine = lightweight process
- Channel = typed pipe between goroutines
- `select` = non-deterministic multiplexer

**Rules before writing concurrent code:**
1. Who **owns** this data? Only owner should write it.
2. What is the **lifetime** of this goroutine? Who waits for it?
3. How does it **stop**? Always via context cancellation or close(done).

---

## Context Propagation — The Thread Running Through Everything

```go
// Context MUST be first parameter, MUST be named ctx
func (s *Service) ProcessOrder(ctx context.Context, id string) error {
    // Derive with timeout for downstream calls
    callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel() // ALWAYS defer cancel — prevents goroutine leak

    result, err := s.client.Get(callCtx, id)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return fmt.Errorf("ProcessOrder: downstream timeout: %w", err)
        }
        return fmt.Errorf("ProcessOrder: %w", err)
    }
    _ = result
    return nil
}

// Context values — only for request-scoped metadata, NOT business logic
type contextKey string
const requestIDKey contextKey = "requestID"

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
    v, _ := ctx.Value(requestIDKey).(string)
    return v
}
```

---

## Fan-Out / Fan-In — Parallel Calls

```go
// Pattern 1: errgroup (preferred for fixed set of goroutines)
import "golang.org/x/sync/errgroup"

func (s *SummaryService) Build(ctx context.Context, orderID string) (*Summary, error) {
    g, ctx := errgroup.WithContext(ctx)

    var order   *Order
    var payment *Payment
    var shipment *Shipment

    g.Go(func() (err error) { order,    err = s.orderRepo.Find(ctx, orderID); return })
    g.Go(func() (err error) { payment,  err = s.paymentSvc.Get(ctx, orderID); return })
    g.Go(func() (err error) { shipment, err = s.shipmentSvc.Get(ctx, orderID); return })

    if err := g.Wait(); err != nil {
        return nil, fmt.Errorf("Build: %w", err) // first non-nil error
    }
    return &Summary{Order: order, Payment: payment, Shipment: shipment}, nil
}

// Pattern 2: typed result channels (for dynamic number of goroutines)
type result[T any] struct {
    val T
    err error
}

func fetchAll[T any](ctx context.Context, ids []string, fetch func(context.Context, string) (T, error)) []result[T] {
    ch := make(chan result[T], len(ids))
    for _, id := range ids {
        id := id // capture
        go func() {
            v, err := fetch(ctx, id)
            ch <- result[T]{v, err}
        }()
    }
    results := make([]result[T], 0, len(ids))
    for range ids {
        results = append(results, <-ch)
    }
    return results
}
```

---

## Worker Pool — Bounded Concurrency

```go
type Job struct {
    OrderID string
    Payload []byte
}

type WorkerPool struct {
    jobs    chan Job
    wg      sync.WaitGroup
    workers int
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
    return &WorkerPool{
        jobs:    make(chan Job, queueSize),
        workers: workers,
    }
}

func (wp *WorkerPool) Start(ctx context.Context, process func(context.Context, Job) error) {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go func() {
            defer wp.wg.Done()
            for {
                select {
                case job, ok := <-wp.jobs:
                    if !ok { return } // channel closed
                    if err := process(ctx, job); err != nil {
                        slog.ErrorContext(ctx, "worker: process job", "error", err, "orderID", job.OrderID)
                    }
                case <-ctx.Done():
                    return
                }
            }
        }()
    }
}

func (wp *WorkerPool) Submit(job Job) bool {
    select {
    case wp.jobs <- job:
        return true
    default:
        return false // queue full — caller decides what to do
    }
}

func (wp *WorkerPool) Stop() {
    close(wp.jobs)
    wp.wg.Wait()
}
```

---

## Pipeline Pattern — Channel Chain

```go
// Each stage: receives from upstream, sends to downstream, stops on ctx.Done()
func generator(ctx context.Context, orders []Order) <-chan Order {
    out := make(chan Order)
    go func() {
        defer close(out)
        for _, o := range orders {
            select {
            case out <- o:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func validate(ctx context.Context, in <-chan Order) <-chan Order {
    out := make(chan Order)
    go func() {
        defer close(out)
        for o := range in {
            if err := o.Validate(); err != nil {
                slog.Warn("skipping invalid order", "id", o.ID, "err", err)
                continue
            }
            select {
            case out <- o:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func enrich(ctx context.Context, in <-chan Order, svc ProductService) <-chan EnrichedOrder {
    out := make(chan EnrichedOrder)
    go func() {
        defer close(out)
        for o := range in {
            enriched, err := svc.Enrich(ctx, o)
            if err != nil { continue }
            select {
            case out <- enriched:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Usage
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

pipeline := enrich(ctx, validate(ctx, generator(ctx, orders)), productSvc)
for result := range pipeline {
    // process result
}
```

---

## Mutex Patterns — When Channels Aren't the Right Fit

```go
// Use mutex when: protecting shared state, not coordinating goroutines

type Cache[K comparable, V any] struct {
    mu    sync.RWMutex
    items map[K]V
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
    c.mu.RLock()         // multiple readers OK
    defer c.mu.RUnlock()
    v, ok := c.items[key]
    return v, ok
}

func (c *Cache[K, V]) Set(key K, val V) {
    c.mu.Lock()          // exclusive writer
    defer c.mu.Unlock()
    c.items[key] = val
}

// sync.Once — lazy initialization, thread-safe
type Config struct {
    once sync.Once
    data *ConfigData
}

func (c *Config) Get() *ConfigData {
    c.once.Do(func() {
        c.data = loadFromFile()
    })
    return c.data
}

// sync.Map — for highly concurrent read-mostly maps (profile before choosing over RWMutex)
var cache sync.Map
cache.Store("key", value)
if v, ok := cache.Load("key"); ok {
    _ = v.(*MyType) // type assert
}
```

---

## Common Goroutine Leaks (Production Bugs)

### Leak 1: Goroutine blocked on channel send, nobody reading

```go
// BAD
func process(req Request) {
    resultCh := make(chan Result) // unbuffered
    go func() {
        resultCh <- compute(req) // blocks forever if caller returns early
    }()
    select {
    case r := <-resultCh:
        return r
    case <-time.After(1 * time.Second):
        return // goroutine leaks! still blocked on send
    }
}

// GOOD — buffered channel or use context
resultCh := make(chan Result, 1) // buffer 1 — goroutine never blocks
```

### Leak 2: Goroutine with no stop signal

```go
// BAD
go func() {
    for {
        data := fetch() // runs forever
        process(data)
    }
}()

// GOOD — always accept ctx.Done()
go func() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            process(fetch())
        case <-ctx.Done():
            return
        }
    }
}()
```

### Leak 3: HTTP response body not closed

```go
// BAD — connection stays open, eventually pool exhausted
resp, err := http.Get(url)
if err != nil { return err }
json.NewDecoder(resp.Body).Decode(&v)

// GOOD
resp, err := http.Get(url)
if err != nil { return err }
defer resp.Body.Close()      // ALWAYS
json.NewDecoder(resp.Body).Decode(&v)
```

### Leak 4: WaitGroup misuse

```go
// BAD — panic if Add called after Wait
var wg sync.WaitGroup
for _, item := range items {
    go func(i Item) {
        wg.Add(1)    // race: Add must happen before goroutine starts
        defer wg.Done()
    }(item)
}
wg.Wait()

// GOOD
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)        // Add BEFORE goroutine launch
    go func(i Item) {
        defer wg.Done()
        process(i)
    }(item)
}
wg.Wait()
```

---

## Race Detection

Always run tests with race detector in CI:

```bash
go test -race ./...
go run -race main.go
```

Common races:
- Loop variable capture: `go func() { use(item) }()` → capture `item := item` first
- Shared slice append: multiple goroutines appending to same slice
- Map concurrent read/write: use `sync.RWMutex` or `sync.Map`
