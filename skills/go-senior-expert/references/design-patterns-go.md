# Design Patterns in Go — Idiomatic Implementations

## Functional Options — The #1 Go Pattern for Config

```go
// Problem: how to provide optional configuration without breaking API

type Server struct {
    host           string
    port           int
    timeout        time.Duration
    maxConnections int
    tlsConfig      *tls.Config
}

// Option is a function that configures a Server
type Option func(*Server)

// Provide one constructor per option — discoverable, composable
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
    return func(s *Server) { s.tlsConfig = cfg }
}

// Constructor: apply defaults, then options
func NewServer(opts ...Option) (*Server, error) {
    s := &Server{
        host:           "0.0.0.0",
        port:           8080,
        timeout:        30 * time.Second,
        maxConnections: 1000,
    }
    for _, opt := range opts {
        opt(s)
    }
    if s.port < 1 || s.port > 65535 {
        return nil, fmt.Errorf("NewServer: invalid port %d", s.port)
    }
    return s, nil
}

// Usage — clean, backward-compatible
srv, err := NewServer(
    WithPort(9090),
    WithTimeout(10*time.Second),
    WithTLS(tlsCfg),
)
```

---

## Middleware / Decorator — HTTP and gRPC

```go
// HTTP middleware
type Middleware func(http.Handler) http.Handler

func WithLogging(logger *slog.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            next.ServeHTTP(rw, r)
            logger.InfoContext(r.Context(), "request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", rw.statusCode,
                "duration", time.Since(start),
            )
        })
    }
}

func WithAuth(validator TokenValidator) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            claims, err := validator.Validate(r.Context(), token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            ctx := context.WithValue(r.Context(), claimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func WithRecovery() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    slog.ErrorContext(r.Context(), "panic recovered",
                        "error", err, "stack", debug.Stack())
                    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}

// Chain middleware — apply right-to-left (outermost first)
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Usage
handler := Chain(
    orderHandler,
    WithRecovery(),
    WithLogging(logger),
    WithAuth(tokenValidator),
)
```

---

## Strategy Pattern

```go
// Different pricing strategies
type PricingStrategy interface {
    Calculate(order *Order) Money
}

type StandardPricing struct{}
type PremiumPricing  struct{ discountPct float64 }
type BulkPricing     struct{ thresholdQty int }

func (s StandardPricing) Calculate(o *Order) Money {
    return o.SubTotal()
}

func (p PremiumPricing) Calculate(o *Order) Money {
    subtotal := o.SubTotal()
    discount := subtotal.Multiply(p.discountPct / 100)
    return subtotal.Subtract(discount)
}

type CheckoutService struct {
    pricing PricingStrategy // injected
}

func NewCheckoutService(pricing PricingStrategy) *CheckoutService {
    return &CheckoutService{pricing: pricing}
}

func (s *CheckoutService) Checkout(ctx context.Context, order *Order) (*Invoice, error) {
    total := s.pricing.Calculate(order) // strategy called here
    return &Invoice{OrderID: order.ID(), Total: total}, nil
}

// Factory: select strategy based on customer tier
func NewPricingStrategy(tier CustomerTier) PricingStrategy {
    switch tier {
    case TierPremium:
        return PremiumPricing{discountPct: 15}
    case TierBulk:
        return BulkPricing{thresholdQty: 10}
    default:
        return StandardPricing{}
    }
}
```

---

## Observer / Event Bus

```go
// Typed event bus — avoids interface{} anti-pattern
type EventHandler[T any] func(ctx context.Context, event T) error

type EventBus[T any] struct {
    mu       sync.RWMutex
    handlers []EventHandler[T]
}

func (b *EventBus[T]) Subscribe(h EventHandler[T]) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.handlers = append(b.handlers, h)
}

func (b *EventBus[T]) Publish(ctx context.Context, event T) error {
    b.mu.RLock()
    handlers := b.handlers
    b.mu.RUnlock()

    var errs []error
    for _, h := range handlers {
        if err := h(ctx, event); err != nil {
            errs = append(errs, err)
        }
    }
    return errors.Join(errs...)
}

// Usage
type OrderPlacedEvent struct {
    OrderID    string
    CustomerID string
    Total      float64
}

bus := &EventBus[OrderPlacedEvent]{}

bus.Subscribe(func(ctx context.Context, e OrderPlacedEvent) error {
    return emailSvc.SendConfirmation(ctx, e.CustomerID, e.OrderID)
})

bus.Subscribe(func(ctx context.Context, e OrderPlacedEvent) error {
    return analyticsSvc.TrackOrder(ctx, e.OrderID, e.Total)
})

// Publish
if err := bus.Publish(ctx, OrderPlacedEvent{OrderID: id, ...}); err != nil {
    log.Warn("event publish partial failure", "err", err)
}
```

---

## Circuit Breaker

```go
import "github.com/sony/gobreaker"

type ResilientPaymentClient struct {
    cb     *gobreaker.CircuitBreaker
    client PaymentClient
}

func NewResilientPaymentClient(client PaymentClient) *ResilientPaymentClient {
    settings := gobreaker.Settings{
        Name:        "payment-service",
        MaxRequests: 5,                   // requests allowed in half-open state
        Interval:    60 * time.Second,    // clear counts after this interval
        Timeout:     30 * time.Second,    // half-open after this duration
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.Requests >= 10 && counts.ConsecutiveFailures >= 5
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            slog.Warn("circuit breaker state change",
                "name", name, "from", from, "to", to)
        },
    }
    return &ResilientPaymentClient{
        cb:     gobreaker.NewCircuitBreaker(settings),
        client: client,
    }
}

func (c *ResilientPaymentClient) Charge(ctx context.Context, amount Money, token string) (string, error) {
    result, err := c.cb.Execute(func() (interface{}, error) {
        return c.client.Charge(ctx, amount, token)
    })
    if err != nil {
        if errors.Is(err, gobreaker.ErrOpenState) {
            return "", ErrPaymentServiceUnavailable
        }
        return "", fmt.Errorf("ResilientPaymentClient.Charge: %w", err)
    }
    return result.(string), nil
}
```

---

## Repository with Unit of Work

```go
// UnitOfWork ensures atomicity across multiple repositories
type UnitOfWork interface {
    Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type GormUnitOfWork struct {
    db *gorm.DB
}

func (u *GormUnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
    return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // inject tx into context so repos pick it up
        txCtx := context.WithValue(ctx, txKey, tx)
        return fn(txCtx)
    })
}

// Repository picks up tx from context
func (r *GormOrderRepository) db(ctx context.Context) *gorm.DB {
    if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
        return tx
    }
    return r.defaultDB
}
```

---

## Anti-Patterns — Call Out in Code Review

| Anti-Pattern | Problem | Go Idiomatic Fix |
|---|---|---|
| `interface{}` (any) as parameter type | No type safety | Use generics or concrete type |
| Returning `nil, nil` | Ambiguous — caller must handle both cases | Return sentinel error or typed result |
| Constructor without validation | Bad state creeps in silently | `NewXxx()` returns error |
| `init()` with side effects | Hidden dependency, untestable | Explicit initialization in main |
| Package-level var for dependency | Global state, hard to test | Constructor injection |
| Named return values (except defer trick) | Confusing, shadowing risk | Explicit return values |
| Panic instead of error | Crashes entire process | Return error; panic only for unrecoverable |
| Copying sync types | `sync.Mutex` must not be copied | Pass by pointer |
| Empty interface slice `[]interface{}` | Type safety lost | Use typed slice or generics |
