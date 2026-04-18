# TypeScript Reference — Production Grade

## 1. Type System — Beyond the Basics

### Branded Types — Tránh primitive obsession

```typescript
// ❌ Không có type safety — dễ nhầm userId với orderId
function getOrder(userId: string, orderId: string) {}
getOrder(orderId, userId); // compile nhưng runtime wrong

// ✅ Branded types — compile-time safety
type UserId  = string & { readonly __brand: 'UserId' };
type OrderId = string & { readonly __brand: 'OrderId' };

const toUserId  = (id: string): UserId  => id as UserId;
const toOrderId = (id: string): OrderId => id as OrderId;

function getOrder(userId: UserId, orderId: OrderId) {}

// Compile error — không thể nhầm được
getOrder(toOrderId('ord_1'), toUserId('usr_1')); // ❌ Type error
getOrder(toUserId('usr_1'), toOrderId('ord_1')); // ✅
```

### Discriminated Union — Pattern exhaustiveness

```typescript
type PaymentResult =
  | { status: 'success'; transactionId: string; amount: number }
  | { status: 'failed';  errorCode: string; retryable: boolean }
  | { status: 'pending'; checkUrl: string };

function handlePayment(result: PaymentResult): string {
  switch (result.status) {
    case 'success': return `Charged ${result.amount}`;
    case 'failed':  return result.retryable ? 'Retry' : 'Contact bank';
    case 'pending': return `Check: ${result.checkUrl}`;
    // ✅ TypeScript forces exhaustive check — thêm new status → compile error đây
    default: const _exhaustive: never = result; return _exhaustive;
  }
}
```

### Template Literal Types — Type-safe string manipulation

```typescript
type HttpMethod  = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
type ApiVersion  = 'v1' | 'v2';
type ApiEndpoint = `/${ApiVersion}/${string}`;

// Type-safe event names
type EventName = `on${Capitalize<string>}`;
const handler: Record<EventName, () => void> = {
  onClick: () => {},
  onHover: () => {},
  // 'click': () => {} // ❌ must start with 'on'
};

// Infer from string pattern
type ExtractParam<T extends string> =
  T extends `${string}:${infer Param}/${string}` ? Param :
  T extends `${string}:${infer Param}` ? Param : never;

type Id = ExtractParam<'/users/:userId/orders/:orderId'>; // 'userId' | 'orderId'
```

### Mapped Types + Conditional Types — Transformation

```typescript
// Deep Readonly
type DeepReadonly<T> = {
  readonly [K in keyof T]: T[K] extends object ? DeepReadonly<T[K]> : T[K];
};

// Optional thành Required, loại trừ undefined
type RequiredNonNullable<T> = {
  [K in keyof T]-?: NonNullable<T[K]>;
};

// Pick nested fields
type PickNested<T, K extends keyof T, NK extends keyof T[K]> = {
  [P in K]: Pick<T[P], NK>;
};

// Function overload với mapped types
type ApiMethods = {
  getUser: (id: string) => Promise<User>;
  createUser: (data: CreateUserDto) => Promise<User>;
  deleteUser: (id: string) => Promise<void>;
};

// Auto-generate mock type
type MockedApi = {
  [K in keyof ApiMethods]: jest.MockedFunction<ApiMethods[K]>;
};
```

### Zod — Runtime validation + type inference (Production standard)

```typescript
import { z } from 'zod';

// Schema là single source of truth
const CreateOrderSchema = z.object({
  userId:   z.string().uuid(),
  items:    z.array(z.object({
    productId: z.string().uuid(),
    quantity:  z.number().int().min(1).max(100),
  })).min(1),
  couponCode: z.string().regex(/^[A-Z]{4}-\d{4}$/).optional(),
  shippingAddress: z.object({
    street:  z.string().min(1),
    city:    z.string().min(1),
    country: z.string().length(2), // ISO 3166-1 alpha-2
    zipCode: z.string(),
  }),
});

// TypeScript type tự động infer — không phải viết tay
type CreateOrderDto = z.infer<typeof CreateOrderSchema>;

// Parse + throw nếu invalid (production-safe)
const parseOrder = (raw: unknown): CreateOrderDto => {
  return CreateOrderSchema.parse(raw); // throws ZodError với message rõ ràng
};

// Transform khi parse
const OrderSchema = CreateOrderSchema.transform(data => ({
  ...data,
  totalItems: data.items.reduce((sum, i) => sum + i.quantity, 0),
}));
```

---

## 2. Async Patterns — Không leak, không block

### Result Type — Không throw qua async boundary

```typescript
// Option 1: Result monad (no exception for expected errors)
type Result<T, E = Error> =
  | { ok: true;  value: T }
  | { ok: false; error: E };

const ok  = <T>(value: T): Result<T, never> => ({ ok: true, value });
const err = <E>(error: E): Result<never, E> => ({ ok: false, error });

async function fetchUser(id: string): Promise<Result<User, 'NOT_FOUND' | 'DB_ERROR'>> {
  try {
    const user = await db.users.findUnique({ where: { id } });
    if (!user) return err('NOT_FOUND');
    return ok(user);
  } catch {
    return err('DB_ERROR');
  }
}

// Caller phải handle cả hai cases — không thể bỏ qua
const result = await fetchUser('123');
if (!result.ok) {
  // result.error is 'NOT_FOUND' | 'DB_ERROR'
  switch (result.error) {
    case 'NOT_FOUND': return res.status(404).json({ error: 'User not found' });
    case 'DB_ERROR':  return res.status(500).json({ error: 'Internal error' });
  }
}
// result.value is User — guaranteed
const user = result.value;
```

### Promise patterns — Parallel, Race, Settlement

```typescript
// ✅ Parallel (independent operations)
const [users, products, orders] = await Promise.all([
  userService.findAll(),
  productService.findAll(),
  orderService.findRecent(),
]);

// ✅ Race với timeout — không bao giờ block forever
const withTimeout = <T>(promise: Promise<T>, ms: number): Promise<T> => {
  const timeout = new Promise<never>((_, reject) =>
    setTimeout(() => reject(new Error(`Timeout after ${ms}ms`)), ms)
  );
  return Promise.race([promise, timeout]);
};

const user = await withTimeout(fetchUser(id), 5000);

// ✅ allSettled — khi muốn tất cả results, dù fail
const results = await Promise.allSettled([
  sendEmail(user1),
  sendEmail(user2),
  sendEmail(user3),
]);

const failures = results
  .filter((r): r is PromiseRejectedResult => r.status === 'rejected')
  .map(r => r.reason);

// ✅ Sequential với accumulation
const processedOrders = await orders.reduce(async (accPromise, order) => {
  const acc = await accPromise;
  const processed = await processOrder(order);
  return [...acc, processed];
}, Promise.resolve([] as ProcessedOrder[]));

// ⚠️ Cẩn thận: forEach + async KHÔNG await
// ❌ orders.forEach(async (order) => await processOrder(order));
// ✅ for...of + await
for (const order of orders) {
  await processOrder(order);
}
```

### AbortController — Cancellable async operations

```typescript
// HTTP request cancellation
class UserService {
  private abortControllers = new Map<string, AbortController>();

  async searchUsers(query: string): Promise<User[]> {
    // Cancel previous in-flight search
    this.abortControllers.get('search')?.abort();

    const controller = new AbortController();
    this.abortControllers.set('search', controller);

    try {
      const response = await fetch(`/api/users?q=${query}`, {
        signal: controller.signal,
      });
      return response.json();
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        return []; // Cancelled — not an error
      }
      throw err;
    } finally {
      this.abortControllers.delete('search');
    }
  }
}

// React useEffect cleanup pattern
useEffect(() => {
  const controller = new AbortController();

  fetchData(id, controller.signal)
    .then(setData)
    .catch(err => {
      if (err.name !== 'AbortError') setError(err);
    });

  return () => controller.abort(); // cleanup
}, [id]);
```

---

## 3. NestJS Patterns — Production Architecture

### Module organization — Feature-first, không layer-first

```typescript
// ❌ Layer-first (anti-pattern cho large codebase)
// controllers/user.controller.ts
// services/user.service.ts
// repositories/user.repository.ts

// ✅ Feature-first (domain co-location)
// users/
//   user.module.ts
//   user.controller.ts
//   user.service.ts
//   user.repository.ts
//   dto/create-user.dto.ts
//   entities/user.entity.ts
//   __tests__/user.service.spec.ts

@Module({
  imports: [TypeOrmModule.forFeature([UserEntity])],
  controllers: [UserController],
  providers: [UserService, UserRepository],
  exports: [UserService], // chỉ export interface, không expose repository
})
export class UserModule {}
```

### Custom Decorators — Tái sử dụng cross-cutting concerns

```typescript
// Param decorator: extract validated user từ JWT
export const CurrentUser = createParamDecorator(
  (data: unknown, ctx: ExecutionContext): UserPayload => {
    const request = ctx.switchToHttp().getRequest();
    return request.user; // populated by JwtAuthGuard
  },
);

// Method decorator: combine Auth + Roles
export const AuthGuard = (...roles: Role[]) =>
  applyDecorators(
    UseGuards(JwtAuthGuard, RolesGuard),
    Roles(...roles),
    ApiSecurity('bearer'),
  );

// Usage — clean controller
@Controller('orders')
export class OrderController {
  @Post()
  @AuthGuard(Role.USER)
  async create(
    @CurrentUser() user: UserPayload,
    @Body() dto: CreateOrderDto,
  ): Promise<OrderResponse> {
    return this.orderService.create(user.id, dto);
  }
}
```

### Exception filters — Consistent error responses

```typescript
@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  constructor(private readonly logger: Logger) {}

  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx  = host.switchToHttp();
    const req  = ctx.getRequest<Request>();
    const res  = ctx.getResponse<Response>();

    const { status, body } = this.mapException(exception);

    this.logger.error({
      message: body.message,
      path: req.url,
      method: req.method,
      status,
      stack: exception instanceof Error ? exception.stack : undefined,
      traceId: req.headers['x-trace-id'],
    });

    res.status(status).json({
      ...body,
      timestamp: new Date().toISOString(),
      path: req.url,
      traceId: req.headers['x-trace-id'],
    });
  }

  private mapException(err: unknown): { status: number; body: ErrorBody } {
    if (err instanceof HttpException) {
      return { status: err.getStatus(), body: { message: err.message, code: 'HTTP_ERROR' } };
    }
    if (err instanceof ZodError) {
      return { status: 400, body: { message: 'Validation failed', code: 'VALIDATION_ERROR', details: err.errors } };
    }
    if (err instanceof PrismaClientKnownRequestError) {
      if (err.code === 'P2002') {
        return { status: 409, body: { message: 'Duplicate entry', code: 'CONFLICT' } };
      }
      if (err.code === 'P2025') {
        return { status: 404, body: { message: 'Record not found', code: 'NOT_FOUND' } };
      }
    }
    // Unknown error — don't leak internals
    return { status: 500, body: { message: 'Internal server error', code: 'INTERNAL_ERROR' } };
  }
}
```

### Interceptors — Request/response transformation

```typescript
// Logging interceptor với timing
@Injectable()
export class LoggingInterceptor implements NestInterceptor {
  constructor(private readonly logger: Logger) {}

  intercept(context: ExecutionContext, next: CallHandler): Observable<unknown> {
    const req   = context.switchToHttp().getRequest<Request>();
    const start = Date.now();

    return next.handle().pipe(
      tap({
        next: () => this.logger.log({
          type: 'REQUEST',
          method: req.method,
          path: req.url,
          duration: Date.now() - start,
          status: 'success',
        }),
        error: (err: unknown) => this.logger.error({
          type: 'REQUEST',
          method: req.method,
          path: req.url,
          duration: Date.now() - start,
          status: 'error',
          error: err instanceof Error ? err.message : String(err),
        }),
      }),
    );
  }
}

// Response wrapping interceptor
@Injectable()
export class TransformInterceptor<T> implements NestInterceptor<T, ApiResponse<T>> {
  intercept(_: ExecutionContext, next: CallHandler): Observable<ApiResponse<T>> {
    return next.handle().pipe(
      map(data => ({ success: true, data, timestamp: new Date().toISOString() }))
    );
  }
}
```

---

## 4. Testing — Không phải coverage, là confidence

### Unit tests — Fast, isolated, deterministic

```typescript
// Pattern: Arrange → Act → Assert
describe('OrderService.create', () => {
  let service: OrderService;
  let orderRepo: jest.Mocked<OrderRepository>;
  let eventBus: jest.Mocked<EventBus>;

  beforeEach(() => {
    orderRepo = {
      create:  jest.fn(),
      findById: jest.fn(),
      // Only mock what this test uses
    } as jest.Mocked<OrderRepository>;

    eventBus = { publish: jest.fn() } as jest.Mocked<EventBus>;
    service = new OrderService(orderRepo, eventBus);
  });

  afterEach(() => jest.clearAllMocks());

  it('should create order and publish event', async () => {
    // Arrange
    const userId = toUserId('usr_123');
    const dto: CreateOrderDto = { items: [{ productId: 'p1', quantity: 2 }] };
    const savedOrder = { id: 'ord_1', userId, ...dto, status: 'pending' };

    orderRepo.create.mockResolvedValue(savedOrder);
    eventBus.publish.mockResolvedValue(undefined);

    // Act
    const result = await service.create(userId, dto);

    // Assert
    expect(result).toEqual(savedOrder);
    expect(orderRepo.create).toHaveBeenCalledWith({ userId, ...dto });
    expect(eventBus.publish).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'OrderCreated', orderId: 'ord_1' })
    );
  });

  it('should NOT publish event if save fails', async () => {
    orderRepo.create.mockRejectedValue(new Error('DB error'));

    await expect(service.create(toUserId('usr_1'), dto)).rejects.toThrow('DB error');
    expect(eventBus.publish).not.toHaveBeenCalled();
  });
});
```

### Integration tests — NestJS TestingModule

```typescript
describe('OrderController (integration)', () => {
  let app: INestApplication;
  let prisma: PrismaService;

  beforeAll(async () => {
    const module = await Test.createTestingModule({
      imports: [AppModule],
    })
    .overrideProvider(PaymentService) // mock external service
    .useValue({ charge: jest.fn().mockResolvedValue({ txnId: 'txn_1' }) })
    .compile();

    app    = module.createNestApplication();
    prisma = module.get(PrismaService);

    app.useGlobalPipes(new ValidationPipe({ transform: true }));
    await app.init();
  });

  afterAll(async () => {
    await prisma.order.deleteMany(); // cleanup
    await app.close();
  });

  it('POST /orders → 201 with valid body', async () => {
    const token = generateTestJwt({ userId: 'usr_1', roles: ['user'] });

    const res = await request(app.getHttpServer())
      .post('/orders')
      .set('Authorization', `Bearer ${token}`)
      .send({ items: [{ productId: 'p1', quantity: 1 }] });

    expect(res.status).toBe(201);
    expect(res.body.data).toMatchObject({
      id: expect.any(String),
      status: 'pending',
    });
  });
});
```

---

## 5. Performance Patterns

### Batch loading — N+1 prevention với DataLoader

```typescript
import DataLoader from 'dataloader';

// Batch function: single DB query cho nhiều IDs
const userBatchFn = async (ids: readonly string[]): Promise<User[]> => {
  const users = await prisma.user.findMany({
    where: { id: { in: [...ids] } },
  });
  // MUST return in same order as ids
  const userMap = new Map(users.map(u => [u.id, u]));
  return ids.map(id => userMap.get(id) ?? new Error(`User ${id} not found`));
};

// DataLoader instance per request (không share across requests!)
const createLoaders = () => ({
  users: new DataLoader<string, User>(userBatchFn, {
    cacheKeyFn: key => key,
    maxBatchSize: 100,
  }),
});

// GraphQL resolver — N+1 tự động resolved
@ResolveField('owner', () => UserType)
async resolveOwner(
  @Parent() order: Order,
  @Context() { loaders }: AppContext,
): Promise<User> {
  return loaders.users.load(order.userId); // batched!
}
```

### Memory management — Không leak streams và event listeners

```typescript
// ✅ Readable stream với backpressure
async function exportLargeDataset(res: Response): Promise<void> {
  const stream = new PassThrough();

  res.setHeader('Content-Type', 'application/json');
  stream.pipe(res);

  try {
    stream.write('[');
    let first = true;

    // Process in chunks — không load toàn bộ vào memory
    const cursor = prisma.order.findMany({ cursor: undefined });
    for await (const batch of paginatePrisma(cursor, 1000)) {
      for (const item of batch) {
        stream.write(`${first ? '' : ','}${JSON.stringify(item)}`);
        first = false;
      }
      // ✅ Yield để không block event loop
      await new Promise(resolve => setImmediate(resolve));
    }

    stream.write(']');
    stream.end();
  } catch (err) {
    stream.destroy(err instanceof Error ? err : new Error(String(err)));
  }
}

// ✅ EventEmitter cleanup — prevent memory leak
class PricingService extends EventEmitter {
  private readonly cleanup: Array<() => void> = [];

  subscribe(topic: string, handler: (data: PriceUpdate) => void): () => void {
    this.on(topic, handler);
    const unsubscribe = () => this.off(topic, handler);
    this.cleanup.push(unsubscribe);
    return unsubscribe;
  }

  destroy(): void {
    this.cleanup.forEach(fn => fn());
    this.cleanup.length = 0;
    this.removeAllListeners();
  }
}
```

---

## 6. TypeScript Anti-Patterns — Senior phải biết avoid

```typescript
// ❌ Type assertion để "fix" type error — lying to compiler
const user = data as User; // data có thể là bất cứ gì
user.email.toLowerCase(); // runtime crash nếu data.email undefined

// ✅ Type guard + validation
function isUser(data: unknown): data is User {
  return typeof data === 'object' && data !== null &&
    'id' in data && 'email' in data && typeof (data as User).email === 'string';
}
if (isUser(data)) {
  user.email.toLowerCase(); // safe
}

// ❌ any phá vỡ type system
function processData(data: any) {
  return data.value.nested.field; // no protection
}

// ✅ unknown + narrowing
function processData(data: unknown) {
  if (!isValidPayload(data)) throw new Error('Invalid payload');
  return data.value; // typed after guard
}

// ❌ Non-null assertion khi không chắc
const name = user!.profile!.name!; // triple! = triple danger

// ✅ Optional chaining + nullish coalescing
const name = user?.profile?.name ?? 'Anonymous';

// ❌ Enum (có nhiều gotchas)
enum Status { Active, Inactive } // Status.Active = 0, có thể so sánh với số bất kỳ

// ✅ Const assertion / string union
const Status = { Active: 'active', Inactive: 'inactive' } as const;
type Status = typeof Status[keyof typeof Status]; // 'active' | 'inactive'

// ❌ Promise trong constructor
class Service {
  private data: Data;
  constructor() {
    this.data = await fetchData(); // ❌ constructor không async được
  }
}

// ✅ Static async factory
class Service {
  private constructor(private readonly data: Data) {}
  static async create(): Promise<Service> {
    const data = await fetchData();
    return new Service(data);
  }
}
```
