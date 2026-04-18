# Cross-Language Reference — TypeScript ↔ Go

## 1. Comparison Matrix — Chọn đúng tool

| Dimension | TypeScript | Go |
|-----------|-----------|-----|
| **Concurrency model** | Event loop (single-threaded) + async/await | Goroutines (M:N scheduling, true parallelism) |
| **CPU-bound work** | ❌ Blocked by event loop | ✅ True parallel với GOMAXPROCS |
| **I/O-bound work** | ✅ Excellent (non-blocking by design) | ✅ Excellent (goroutines lightweight) |
| **Memory model** | GC, less predictable pauses | GC, more predictable, lower footprint |
| **Startup time** | 100-500ms (Node startup) | < 10ms (compiled binary) |
| **Binary size** | Runtime required | Self-contained binary |
| **Type system** | Structural, expressive generics | Nominal, simpler generics (Go 1.18+) |
| **Error handling** | Exceptions (try/catch) + Result types | Explicit return values |
| **Ecosystem** | npm (massive) | Go modules (curated, smaller) |
| **Learning curve** | Medium (JS background helps) | Low (simple language spec) |
| **Refactoring safety** | High (TypeScript strict) | High (compiler enforces) |

---

## 2. Khi nào dùng Go vs TypeScript

### Go cho:

```
✅ High-throughput services (> 10K RPS per instance)
   → Goroutines cheaper than async callbacks at scale

✅ CPU-intensive processing
   → Image resizing, video processing, crypto, compression
   → TS/Node: CPU work blocks event loop → latency spike

✅ Predictable latency requirements (P99 < 5ms)
   → Go GC pauses: microseconds
   → Node GC pauses: can be 10-50ms under load

✅ Memory-constrained environments
   → Go binary: ~10-30MB RSS for simple service
   → Node: 50-150MB RSS base (V8 + Node runtime)

✅ CLI tools và system utilities
   → Single binary, no runtime dependency
   → Cross-compile easily (GOOS=linux GOARCH=amd64)

✅ Long-running background workers
   → Worker pools with goroutines: elegant, efficient
   → No event loop to starve

✅ Infrastructure-adjacent code
   → Kubernetes operators, CLI tools, proxies
   → Same ecosystem as Docker, K8s, Terraform
```

### TypeScript cho:

```
✅ REST/GraphQL API với complex business logic
   → NestJS DI, decorators, pipes, guards
   → Rich ecosystem: Prisma, Drizzle, TypeORM

✅ Shared types với frontend
   → tRPC: end-to-end type safety TS → TS
   → Zod schemas: validation on both ends

✅ Rapid feature development
   → npm ecosystem speeds up prototyping
   → Decorator-based patterns reduce boilerplate

✅ Complex domain modeling
   → Class-based OOP + generics + conditional types
   → DDD với NestJS modules maps cleanly

✅ Team từ JavaScript background
   → Learning curve: low for JS devs
   → Gradual adoption (any → typed migration)

✅ Serverless functions
   → Cold start: Node acceptable cho most cases
   → Lambda, Cloud Functions ecosystem
```

---

## 3. Polyglot Architecture — Cả hai cùng nhau

### Pattern: Go API Gateway + TypeScript Business Services

```
Client
  │
  ▼
┌─────────────────┐
│  Go API Gateway │  ← Rate limiting, auth, routing, SSL termination
│  (Gin/Fiber)    │    High-throughput, low-latency edge
└────────┬────────┘
         │ gRPC / HTTP
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼────────────┐
│ Go    │ │  TypeScript   │
│Worker │ │  Services     │
│(ETL,  │ │  (NestJS:     │
│ queue │ │   Orders,     │
│ proc) │ │   Users,      │
└───────┘ │   Products)   │
          └───────────────┘
```

### Pattern: gRPC bridge giữa Go và TypeScript

```protobuf
// user.proto — shared contract
syntax = "proto3";

service UserService {
  rpc GetUser (GetUserRequest) returns (User);
  rpc CreateUser (CreateUserRequest) returns (User);
  rpc StreamUserEvents (StreamRequest) returns (stream UserEvent);
}

message User {
  string id         = 1;
  string email      = 2;
  string name       = 3;
  int64  created_at = 4; // Unix timestamp
}
```

```go
// Go server implementation
type userServer struct {
    pb.UnimplementedUserServiceServer
    repo UserRepository
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    user, err := s.repo.FindByID(ctx, req.Id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
        }
        return nil, status.Errorf(codes.Internal, "internal error: %v", err)
    }
    return &pb.User{
        Id:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: user.CreatedAt.Unix(),
    }, nil
}
```

```typescript
// TypeScript client (NestJS service calling Go gRPC server)
import { credentials } from '@grpc/grpc-js';
import { UserServiceClient } from './generated/user_grpc_pb';

@Injectable()
export class UserGrpcClient {
  private readonly client: UserServiceClient;

  constructor() {
    this.client = new UserServiceClient(
      process.env.USER_SERVICE_URL!,
      credentials.createInsecure(),
    );
  }

  async getUser(id: string): Promise<User> {
    return new Promise((resolve, reject) => {
      const req = new GetUserRequest();
      req.setId(id);

      this.client.getUser(req, (err, response) => {
        if (err) {
          if (err.code === status.NOT_FOUND) {
            reject(new NotFoundException(`User ${id} not found`));
          } else {
            reject(new InternalServerErrorException(err.message));
          }
          return;
        }
        resolve({
          id:        response.getId(),
          email:     response.getEmail(),
          name:      response.getName(),
          createdAt: new Date(response.getCreatedAt() * 1000),
        });
      });
    });
  }
}
```

---

## 4. Equivalent Patterns — Side by Side

### Dependency Injection

```typescript
// TypeScript — NestJS DI với decorators
@Injectable()
class OrderService {
  constructor(
    @InjectRepository(OrderEntity) private readonly repo: Repository<OrderEntity>,
    private readonly eventBus: EventBus,
    private readonly logger: Logger,
  ) {}
}
```

```go
// Go — Constructor injection (manual hoặc Wire)
type OrderService struct {
    repo     OrderRepository
    eventBus EventPublisher
    logger   *slog.Logger
}

func NewOrderService(
    repo OrderRepository,
    eventBus EventPublisher,
    logger *slog.Logger,
) *OrderService {
    return &OrderService{repo: repo, eventBus: eventBus, logger: logger}
}

// Wire (Google's DI generator)
// go:build wireinject
func InitializeApp(cfg *Config) (*App, error) {
    wire.Build(
        NewOrderService,
        NewOrderRepository,
        NewEventBus,
        NewApp,
    )
    return nil, nil
}
```

### Middleware / Interceptor

```typescript
// TypeScript — NestJS Guard
@Injectable()
export class JwtAuthGuard implements CanActivate {
  constructor(private readonly jwtService: JwtService) {}

  canActivate(context: ExecutionContext): boolean {
    const req   = context.switchToHttp().getRequest<Request>();
    const token = req.headers.authorization?.split(' ')[1];
    if (!token) throw new UnauthorizedException();

    try {
      req.user = this.jwtService.verify(token);
      return true;
    } catch {
      throw new UnauthorizedException('Invalid token');
    }
  }
}
```

```go
// Go — HTTP Middleware
func JWTMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            claims, err := validateJWT(token, jwtSecret)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            // Inject into context
            ctx := WithUserClaims(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Async/Concurrent Processing

```typescript
// TypeScript — Async queue với BullMQ
@Processor('orders')
export class OrderProcessor {
  @Process('send-confirmation')
  async processConfirmation(job: Job<OrderConfirmationJob>): Promise<void> {
    const { orderId, userId } = job.data;

    await job.updateProgress(10);
    const order = await this.orderService.findById(orderId);

    await job.updateProgress(50);
    await this.emailService.sendOrderConfirmation(userId, order);

    await job.updateProgress(100);
  }
}

// Add to queue
await this.orderQueue.add('send-confirmation', { orderId, userId }, {
  attempts: 3,
  backoff: { type: 'exponential', delay: 1000 },
});
```

```go
// Go — Worker pool với channels
type OrderProcessor struct {
    jobs    chan OrderJob
    results chan ProcessResult
    wg      sync.WaitGroup
    logger  *slog.Logger
}

func NewOrderProcessor(workers int) *OrderProcessor {
    p := &OrderProcessor{
        jobs:    make(chan OrderJob, 100), // buffered
        results: make(chan ProcessResult, 100),
    }

    for range workers {
        p.wg.Add(1)
        go p.worker()
    }
    return p
}

func (p *OrderProcessor) worker() {
    defer p.wg.Done()
    for job := range p.jobs {
        result := p.processWithRetry(job, 3)
        p.results <- result
    }
}

func (p *OrderProcessor) processWithRetry(job OrderJob, maxRetries int) ProcessResult {
    var lastErr error
    for attempt := range maxRetries {
        if attempt > 0 {
            // Exponential backoff
            time.Sleep(time.Duration(1<<attempt) * time.Second)
        }
        if err := p.process(job); err != nil {
            lastErr = err
            continue
        }
        return ProcessResult{JobID: job.ID, Success: true}
    }
    return ProcessResult{JobID: job.ID, Success: false, Error: lastErr}
}
```

### Validation

```typescript
// TypeScript — Zod (runtime + compile-time)
const CreateOrderSchema = z.object({
  userId:   z.string().uuid(),
  items:    z.array(z.object({
    productId: z.string().uuid(),
    quantity:  z.number().int().positive(),
  })).nonempty(),
});
type CreateOrderDto = z.infer<typeof CreateOrderSchema>;

// NestJS pipe integration
@Injectable()
export class ZodValidationPipe implements PipeTransform {
  constructor(private schema: ZodSchema) {}

  transform(value: unknown) {
    const result = this.schema.safeParse(value);
    if (!result.success) {
      throw new BadRequestException({
        message: 'Validation failed',
        errors: result.error.errors,
      });
    }
    return result.data;
  }
}
```

```go
// Go — struct tags + custom validator
type CreateOrderRequest struct {
    UserID string      `json:"user_id" validate:"required,uuid"`
    Items  []OrderItem `json:"items"   validate:"required,min=1,dive"`
}

type OrderItem struct {
    ProductID string `json:"product_id" validate:"required,uuid"`
    Quantity  int    `json:"quantity"   validate:"required,min=1,max=100"`
}

// Custom validation trong domain
func (r *CreateOrderRequest) Validate() error {
    var errs []string

    if r.UserID == "" {
        errs = append(errs, "user_id is required")
    }
    if len(r.Items) == 0 {
        errs = append(errs, "items cannot be empty")
    }
    for i, item := range r.Items {
        if item.Quantity <= 0 {
            errs = append(errs, fmt.Sprintf("items[%d].quantity must be positive", i))
        }
    }
    if len(errs) > 0 {
        return &ValidationError{Fields: errs}
    }
    return nil
}
```

---

## 5. Shared Infrastructure Patterns

### Distributed Tracing — OpenTelemetry (both languages)

```typescript
// TypeScript — OpenTelemetry với NestJS
import { trace, context } from '@opentelemetry/api';

@Injectable()
export class OrderService {
  async createOrder(dto: CreateOrderDto): Promise<Order> {
    const tracer = trace.getTracer('order-service');
    const span   = tracer.startSpan('createOrder');

    return context.with(trace.setSpan(context.active(), span), async () => {
      try {
        span.setAttributes({ 'order.userId': dto.userId });
        const order = await this.repo.create(dto);
        span.setAttributes({ 'order.id': order.id });
        return order;
      } catch (err) {
        span.recordException(err as Error);
        span.setStatus({ code: SpanStatusCode.ERROR });
        throw err;
      } finally {
        span.end();
      }
    });
  }
}
```

```go
// Go — OpenTelemetry
import "go.opentelemetry.io/otel"

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    ctx, span := otel.Tracer("order-service").Start(ctx, "CreateOrder")
    defer span.End()

    span.SetAttributes(
        attribute.String("order.userId", req.UserID),
    )

    order, err := s.repo.Create(ctx, req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    span.SetAttributes(attribute.String("order.id", order.ID))
    return order, nil
}
```

### Structured Logging — Consistent format

```typescript
// TypeScript — Winston/Pino với structured output
const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  formatters: {
    level: (label) => ({ level: label }),
  },
  base: { service: 'order-service', version: pkg.version },
});

// Usage
logger.info({ orderId, userId, amount }, 'Order created');
logger.error({ err, orderId }, 'Failed to process order');
```

```go
// Go — slog (stdlib Go 1.21+)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
    ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
        if a.Key == slog.TimeKey {
            a.Value = slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339))
        }
        return a
    },
})).With("service", "order-service", "version", version)

// Usage
logger.InfoContext(ctx, "order created",
    "orderId", order.ID,
    "userId", order.UserID,
    "amount", order.Total,
)
logger.ErrorContext(ctx, "failed to process order",
    "error", err,
    "orderId", orderID,
)
```
