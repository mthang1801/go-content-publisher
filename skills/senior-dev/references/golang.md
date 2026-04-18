# Go Reference — Production Grade

## 1. Concurrency — Go's Superpower và Minefield

### Goroutine lifecycle management — Không bao giờ "fire and forget"

```go
// ❌ Goroutine leak — không có way để stop hoặc track
go func() {
    for {
        data := <-ch // block forever nếu channel không bao giờ close
        process(data)
    }
}()

// ✅ Context-aware goroutine — cancellable, trackable
func (w *Worker) Start(ctx context.Context) error {
    g, gCtx := errgroup.WithContext(ctx)

    g.Go(func() error {
        return w.processLoop(gCtx)
    })

    g.Go(func() error {
        return w.healthCheck(gCtx)
    })

    return g.Wait() // blocks until all goroutines done, returns first error
}

func (w *Worker) processLoop(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // clean shutdown
        case msg, ok := <-w.input:
            if !ok {
                return nil // channel closed
            }
            if err := w.process(msg); err != nil {
                w.logger.Error("process failed", "error", err)
                // decide: continue or return error
            }
        }
    }
}
```

### Channel patterns — Beyond basic send/receive

```go
// Fan-out: 1 input → N workers
func fanOut[T any](ctx context.Context, input <-chan T, n int, fn func(T) error) error {
    g, gCtx := errgroup.WithContext(ctx)

    for range n {
        g.Go(func() error {
            for {
                select {
                case <-gCtx.Done():
                    return gCtx.Err()
                case item, ok := <-input:
                    if !ok {
                        return nil
                    }
                    if err := fn(item); err != nil {
                        return err
                    }
                }
            }
        })
    }
    return g.Wait()
}

// Fan-in: N inputs → 1 output (merge channels)
func fanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
    out := make(chan T)
    var wg sync.WaitGroup

    merge := func(ch <-chan T) {
        defer wg.Done()
        for {
            select {
            case <-ctx.Done():
                return
            case v, ok := <-ch:
                if !ok { return }
                out <- v
            }
        }
    }

    wg.Add(len(inputs))
    for _, ch := range inputs {
        go merge(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

// Semaphore: limit concurrency
type Semaphore chan struct{}

func NewSemaphore(limit int) Semaphore { return make(Semaphore, limit) }

func (s Semaphore) Acquire() { s <- struct{}{} }
func (s Semaphore) Release() { <-s }

// Usage
sem := NewSemaphore(10) // max 10 concurrent
for _, url := range urls {
    sem.Acquire()
    go func(u string) {
        defer sem.Release()
        fetch(u)
    }(url)
}
```

### sync package — Dùng đúng tool

```go
// sync.Once — initialization một lần duy nhất
type Config struct {
    once     sync.Once
    instance *configData
}

func (c *Config) Get() *configData {
    c.once.Do(func() {
        c.instance = loadConfig() // chạy 1 lần, thread-safe
    })
    return c.instance
}

// sync.Map — concurrent map (dùng khi read >> write)
var cache sync.Map

func getCached(key string) (string, bool) {
    val, ok := cache.Load(key)
    if !ok { return "", false }
    return val.(string), true
}

func setCached(key, val string) {
    cache.Store(key, val)
}

// sync.RWMutex — concurrent reads, exclusive write
type SafeMap[K comparable, V any] struct {
    mu sync.RWMutex
    m  map[K]V
}

func (s *SafeMap[K, V]) Get(key K) (V, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.m[key]
    return v, ok
}

func (s *SafeMap[K, V]) Set(key K, val V) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.m[key] = val
}

// sync.Pool — object reuse, giảm GC pressure
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

func processRequest(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf) // ✅ trả về pool sau khi dùng
    }()

    buf.Write(data)
    return buf.String()
}
```

---

## 2. Error Handling — Go's Philosophy

### Error wrapping và sentinel errors

```go
// Sentinel errors — comparable với ==
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrConflict      = errors.New("conflict")
)

// Typed errors — carry additional context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

// Wrap với context, unwrap với errors.Is / errors.As
func (r *UserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    user, err := r.db.QueryRowContext(ctx, `SELECT ...`, id).Scan(...)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("user %s: %w", id, ErrNotFound) // wrap sentinel
    }
    if err != nil {
        return nil, fmt.Errorf("query user %s: %w", id, err) // add context
    }
    return user, nil
}

// Caller: check without losing context
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, ErrNotFound // re-wrap hoặc pass through
        }
        return nil, fmt.Errorf("get user: %w", err) // add service context
    }
    return user, nil
}

// HTTP handler: translate errors
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.GetUser(r.Context(), chi.URLParam(r, "id"))
    if err != nil {
        var valErr *ValidationError
        switch {
        case errors.Is(err, ErrNotFound):
            http.Error(w, "not found", http.StatusNotFound)
        case errors.As(err, &valErr):
            http.Error(w, valErr.Error(), http.StatusBadRequest)
        default:
            h.logger.Error("get user failed", "error", err)
            http.Error(w, "internal error", http.StatusInternalServerError)
        }
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

### Must pattern — Initialization errors

```go
// Must: panic ở startup nếu config invalid (fail fast, không fail slowly)
func mustGetenv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        panic(fmt.Sprintf("required env var %s is not set", key))
    }
    return v
}

func mustParseURL(raw string) *url.URL {
    u, err := url.Parse(raw)
    if err != nil || u.Host == "" {
        panic(fmt.Sprintf("invalid URL %q: %v", raw, err))
    }
    return u
}

// Dùng trong init/main — không dùng trong request handlers
func main() {
    cfg := Config{
        DatabaseURL: mustGetenv("DATABASE_URL"),
        RedisURL:    mustParseURL(mustGetenv("REDIS_URL")),
        Port:        mustGetenvInt("PORT"),
    }
    // Nếu config invalid → panic ở startup → rõ ràng, không subtle failure
}
```

---

## 3. Interface Design — Go's Idiomatic Way

### Small interfaces — Accept interfaces, return structs

```go
// ❌ Large interface — hard to mock, violates ISP
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindAll(ctx context.Context, opts QueryOptions) ([]*User, error)
    Count(ctx context.Context) (int64, error)
}

// ✅ Small, focused interfaces — compose them
type UserReader interface {
    FindByID(ctx context.Context, id string) (*User, error)
}

type UserWriter interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
}

type UserDeleter interface {
    Delete(ctx context.Context, id string) error
}

// Service chỉ declare gì nó cần
type UserService struct {
    reader  UserReader  // chỉ cần đọc
    writer  UserWriter  // chỉ cần write
    events  EventPublisher
}

// ✅ Interface tại điểm consume, không điểm define
// package user/service — không phải user/repository
type orderRepository interface { // lowercase = package-private interface
    FindByUserID(ctx context.Context, userID string) ([]*Order, error)
}
```

### Functional options — Flexible, backward-compatible config

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
    tls     *tls.Config
    logger  *slog.Logger
}

type Option func(*Server)

func WithHost(host string) Option {
    return func(s *Server) { s.host = host }
}

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func WithTLS(cfg *tls.Config) Option {
    return func(s *Server) { s.tls = cfg }
}

func NewServer(opts ...Option) *Server {
    // Sensible defaults
    s := &Server{
        host:    "0.0.0.0",
        port:    8080,
        timeout: 30 * time.Second,
        logger:  slog.Default(),
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage — clean, extensible
srv := NewServer(
    WithPort(9090),
    WithTimeout(60 * time.Second),
    WithTLS(tlsConfig),
)
```

---

## 4. Context — Everywhere, Every Time

```go
// Context là first parameter, luôn luôn
func (s *Service) ProcessOrder(ctx context.Context, orderID string) error {

    // Timeout cho toàn bộ operation
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel() // ✅ Always defer cancel

    // Value: request-scoped data (không phải config, không phải DB)
    traceID := TraceIDFromContext(ctx) // extracted từ middleware

    // Propagate xuống — mọi DB call, HTTP call, goroutine đều nhận ctx
    order, err := s.orderRepo.FindByID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("find order %s: %w", orderID, err)
    }

    // Check cancellation trong long loops
    for _, item := range order.Items {
        select {
        case <-ctx.Done():
            return fmt.Errorf("processing cancelled: %w", ctx.Err())
        default:
        }
        if err := s.processItem(ctx, item); err != nil {
            return err
        }
    }
    return nil
}

// Context key — typed để tránh collision
type contextKey struct{ name string }

var (
    traceIDKey = &contextKey{"traceID"}
    userIDKey  = &contextKey{"userID"}
)

func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
    if id, ok := ctx.Value(traceIDKey).(string); ok {
        return id
    }
    return ""
}
```

---

## 5. Testing — Table-Driven và Testable Design

### Table-driven tests — Go standard

```go
func TestCalculateDiscount(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name          string
        orderAmount   float64
        userTier      UserTier
        couponCode    string
        wantDiscount  float64
        wantErr       error
    }{
        {
            name:         "gold tier no coupon",
            orderAmount:  100.0,
            userTier:     TierGold,
            couponCode:   "",
            wantDiscount: 10.0, // 10% for gold
        },
        {
            name:         "silver tier with valid coupon",
            orderAmount:  100.0,
            userTier:     TierSilver,
            couponCode:   "SAVE20",
            wantDiscount: 25.0, // 5% silver + 20% coupon
        },
        {
            name:       "invalid coupon",
            orderAmount: 100.0,
            userTier:   TierBronze,
            couponCode: "INVALID",
            wantErr:    ErrInvalidCoupon,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // parallel subtests

            discount, err := CalculateDiscount(tc.orderAmount, tc.userTier, tc.couponCode)

            if tc.wantErr != nil {
                assert.ErrorIs(t, err, tc.wantErr)
                return
            }
            require.NoError(t, err)
            assert.InDelta(t, tc.wantDiscount, discount, 0.01)
        })
    }
}
```

### Interface mocking — testify/mock

```go
// Auto-generate mock: mockery --name=UserRepository --output=mocks
// hoặc viết tay:

type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// Test với mock
func TestUserService_GetUser(t *testing.T) {
    t.Parallel()

    repo := new(MockUserRepo)
    svc  := NewUserService(repo)

    // Setup expectation
    expectedUser := &User{ID: "usr_1", Email: "test@example.com"}
    repo.On("FindByID", mock.Anything, "usr_1").
        Return(expectedUser, nil).
        Once()

    // Execute
    user, err := svc.GetUser(context.Background(), "usr_1")

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    repo.AssertExpectations(t)
}
```

### Benchmarks + pprof

```go
// Benchmark — identify hot paths
func BenchmarkJSONMarshal(b *testing.B) {
    order := generateTestOrder()

    b.ReportAllocs() // show memory allocations
    b.ResetTimer()

    for range b.N {
        _, err := json.Marshal(order)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// Run: go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
// Analyze: go tool pprof cpu.prof → top, list FunctionName, web

// Race detector — always run in CI
// go test -race ./...
```

---

## 6. Go Anti-Patterns — Senior phải know

```go
// ❌ Goroutine leak — không check context trong loop
go func() {
    for {
        time.Sleep(time.Second)
        doWork() // runs forever even after service shutdown
    }
}()

// ✅ Context-aware loop
go func() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            doWork()
        }
    }
}()

// ❌ Capture loop variable (Go < 1.22)
for _, v := range items {
    go func() {
        process(v) // v là shared variable — race condition
    }()
}

// ✅ Pass as argument (hoặc Go 1.22+ tự handle)
for _, v := range items {
    v := v // shadow
    go func() { process(v) }()
}
// Go 1.22+: loop variable per iteration, không cần shadow

// ❌ Naked return trong complex function — confusing
func calculate(x, y int) (result int, err error) {
    // ... 50 lines ...
    return // what is returned?
}

// ✅ Explicit return
func calculate(x, y int) (int, error) {
    // ... 50 lines ...
    return result, nil
}

// ❌ Interface pollution — return interface từ constructor
func NewService() ServiceInterface { // caller không thể extend
    return &serviceImpl{}
}

// ✅ Return concrete type — caller decides what interface to use
func NewService() *Service {
    return &Service{}
}

// ❌ Panic trong non-startup code
func GetUser(id string) *User {
    user, err := db.Find(id)
    if err != nil {
        panic(err) // kills entire server
    }
    return user
}

// ✅ Return error
func GetUser(ctx context.Context, id string) (*User, error) {
    user, err := db.Find(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return user, nil
}

// ❌ Global state — impossible to test
var db *sql.DB // package-level global

// ✅ Dependency injection
type Repository struct {
    db *sql.DB // injected
}
```

---

## 7. Go + DDD — Clean Architecture

```
internal/
├── domain/           ← Pure business logic, no imports từ infrastructure
│   ├── user/
│   │   ├── user.go         (entity)
│   │   ├── user_repo.go    (repository interface)
│   │   └── user_service.go (domain service)
│   └── order/
├── application/      ← Use cases, orchestrates domain + infrastructure
│   ├── create_order.go
│   └── cancel_order.go
├── infrastructure/   ← DB, HTTP, cache implementations
│   ├── postgres/
│   │   └── user_repo.go    (implements domain.UserRepository)
│   ├── redis/
│   └── http/
└── cmd/              ← Entry points, dependency wiring
    └── api/
        └── main.go
```

```go
// domain/order/order.go — pure business logic
package order

import (
    "errors"
    "time"
)

type Order struct {
    id        OrderID
    userID    UserID
    items     []Item
    status    Status
    createdAt time.Time
}

// Factory method — enforces invariants
func NewOrder(userID UserID, items []Item) (*Order, error) {
    if len(items) == 0 {
        return nil, errors.New("order must have at least one item")
    }
    return &Order{
        id:        NewOrderID(),
        userID:    userID,
        items:     items,
        status:    StatusPending,
        createdAt: time.Now(),
    }, nil
}

// Domain behavior — not CRUD
func (o *Order) Cancel(reason string) error {
    if o.status == StatusShipped {
        return ErrCannotCancelShippedOrder
    }
    o.status = StatusCancelled
    return nil
}
```
