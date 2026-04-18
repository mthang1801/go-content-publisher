---
name: backend-patterns
description: >
    NestJS full-stack architecture guide — DDD, CQRS, Event Sourcing, Saga, Microservices, WebSockets, OpenAPI.
    Covers Overview → Fundamentals → Techniques → Security → WebSockets → Microservices → OpenAPI.
    Triết lý, kiến trúc, rules, bảng tra cứu — không có code (xem EXAMPLE.md).
    Dự án: Antigravity — NestJS + DDD + Event Sourcing Monorepo.
---

# NestJS Backend Patterns — Antigravity Architecture Guide

> **Phân tầng nhanh**: Triết lý + Rules + Bảng tra cứu ở đây. Code thực tế → xem `EXAMPLE.md`.

---

## 🚨 FORCE RULE — Import Alias (MANDATORY, KHÔNG ĐƯỢC VI PHẠM)

> **Lý do được thêm**: Audit 2026-03-13 phát hiện 63 violations trên 22 files — toàn bộ `src/presentation/` và một số `src/application/` đang dùng relative imports `../../../application/...` thay vì alias đúng. Đã fix hàng loạt bằng sed. Rule này bắt buộc từ nay.

### Các alias đã define trong `tsconfig.json`

| Alias | Maps to | Dùng cho |
| --- | --- | --- |
| `@application/*` | `src/application/*` | Use Cases, Services, Commands, Queries trong Application layer |
| `@domain/*` | `src/domain/*` | Domain Entities, Aggregates, Value Objects, Ports |
| `@infrastructure/*` | `src/infrastructure/*` | Adapters, ORM Entities, DB clients, External services |
| `@presentation/*` | `src/presentation/*` | Controllers, DTOs, Modules của Presentation layer |
| `@modules-shared/*` | `src/shared/*` | Shared utilities dùng chung giữa các module trong src/ |
| `@ddd/*` | `libs/src/ddd/*` | DDD primitives (BaseEntity, BaseCommand, IRepository...) |
| `@core/*` | `libs/src/core/*` | Core infrastructure (TypeORM base, Redis, Kafka...) |
| `@shared/*` | `libs/src/shared/*` | Shared libs utilities (logger, decorators...) |
| `@common/*` | `libs/src/common/*` | Common exceptions, interceptors, guards |

### Quy tắc bắt buộc

**✅ ĐÚNG — Luôn dùng alias khi import cross-boundary (khác thư mục src/ top-level):**

```typescript
// Presentation → Application
import { CreateOrderUseCase } from '@application/order/use-cases/create-order.use-case';

// Presentation → Infrastructure
import { OrderRepository } from '@infrastructure/order/repositories/order.repository';

// Application → Domain
import { OrderRepositoryPort } from '@domain/order/ports/order-repository.port';

// Bất kỳ layer nào → libs
import { BaseCommand } from '@ddd/application/commands/command.base';
import { UsecaseBadRequestException } from '@common/exceptions';
```

**❌ SAI — Không được dùng relative path `../` khi đi xuyên boundary:**

```typescript
// WRONG: presentation → application qua relative path
import { CreateOrderUseCase } from '../../../application/order/use-cases/create-order.use-case';

// WRONG: 3+ levels up === cross-boundary, bắt buộc dùng alias
import { OrderRepositoryPort } from '../../../../domain/order/ports/order-repository.port';
```

**✅ OK — Relative import trong CÙNG module (1–2 cấp):**

```typescript
// OK: cùng trong /presentation/portal/order/
import { OrderCommandController } from './controllers/order-command.controller';
import { CreateOrderDto } from '../dtos/create-order.dto';
```

### Nguyên tắc phân biệt

> **Rule of thumb**: Nếu đường dẫn relative có **2+ cấp `../`** và vượt qua ranh giới `src/<layer>/`, **BẮT BUỘC dùng alias**.

Ranh giới layer là: `application/`, `domain/`, `infrastructure/`, `presentation/`, `shared/` dưới `src/`.

### Kiểm tra vi phạm (chạy trước khi commit)

```bash
# Detect cross-boundary relative imports (sẽ báo lỗi nếu còn vi phạm)
grep -rn "from '\.\." src --include="*.ts" | grep -v ".spec." \
  | grep -E "from '(\.\./){2,}" \
  | grep -E "(application|domain|infrastructure|presentation|shared)"
# Output rỗng = OK. Có output = VIOLATION cần fix.
```

---

## Triết lý cốt lõi (Core Philosophy)

### Clean Architecture Dependency Rule

```
Presentation → Application → Domain ← Infrastructure
```

- **Domain** — trung tâm, không import bất kỳ framework nào
- **Application** — orchestrate use-cases, không chứa business logic
- **Infrastructure** — implement ports, không expose sang Domain
- **Presentation** — validate input, delegate sang Application

### DDD First Principles

1. **Model the domain** trước, technology sau
2. **Ubiquitous Language** — code dùng đúng ngôn ngữ business
3. **Aggregate Boundary** — mọi thay đổi đi qua Aggregate Root
4. **Domain Events** — state change được publish, không poll
5. **Result<T>** — không throw exception ở Domain/Application layer

### SOLID trong NestJS

| Principle                 | Áp dụng                                                         |
| ------------------------- | --------------------------------------------------------------- |
| S — Single Responsibility | Mỗi class một lý do thay đổi (Entity ≠ Repository ≠ Controller) |
| O — Open/Closed           | Extend BaseCommand, không modify nó                             |
| L — Liskov Substitution   | OrderRepository implements OrderRepositoryPort                  |
| I — Interface Segregation | IReadRepository ≠ IWriteRepository                              |
| D — Dependency Inversion  | Depend on Port abstraction, không on Concrete                   |

---

## PHẦN 1 — OVERVIEW

### 1.1 Controllers

**Triết lý:**

- Controllers chỉ xử lý HTTP binding, không chứa business logic
- Tách Command Controller và Query Controller (CQRS Presentation)
- Validate bằng DTO trước khi delegate sang Use Case
- `@Controller()` nhận prefix route; `@Get()`, `@Post()`, etc. nhận path

**Rules:**

- ✅ Dùng `executeWithHooks()` thay vì `execute()` trực tiếp
- ✅ Response qua Interceptor (không format tay trong controller)
- ✅ Parse UUID/Int với Pipe (`@Param('id', ParseUUIDPipe)`)
- ❌ Không inject Repository vào Controller
- ❌ Không chứa `if/else` business logic trong Controller

**Naming Convention:**

| Type               | Pattern                          | Ví dụ                         |
| ------------------ | -------------------------------- | ----------------------------- |
| Command Controller | `<Domain>CommandController`      | `OrderCommandController`      |
| Query Controller   | `<Domain>QueryController`        | `OrderQueryController`        |
| File               | `<domain>-command.controller.ts` | `order-command.controller.ts` |

**Versioning:** URI versioning `v1/`, `v2/` — bật qua `app.enableVersioning()`

### 1.2 Providers

**Triết lý:**

- Provider = bất kỳ class nào được NestJS IoC container quản lý
- Use Cases, Services, Repositories, Factories đều là Providers
- Ưu tiên Constructor Injection — rõ ràng, testable

**Provider Types:**

| Type    | `provide`     | `useClass/useValue/useFactory` | Khi dùng       |
| ------- | ------------- | ------------------------------ | -------------- |
| Class   | token = class | `useClass`                     | Default        |
| Value   | string token  | `useValue`                     | Config objects |
| Factory | string token  | `useFactory`                   | Async init     |
| Alias   | class         | `useExisting`                  | Rename token   |

**Token Conventions:**

- Abstract class làm token cho Port (DDD) — không dùng string literal
- `@Inject(OrderRepositoryPort)` → type-safe injection

### 1.3 Modules

**Triết lý:**

- Module = đơn vị đóng gói theo bounded context
- Mọi thứ trong module là private mặc định — phải `exports[]` để share
- `@Global()` chỉ dùng cho infrastructure (Redis, Logger, Config)

**Module Hierarchy cho DDD:**

```
AppModule
├── PresentationModule
│   └── OrderPresentationModule (controllers)
├── LibDDDModule (global)
├── TypeOrmInfrastructureModule (global)
├── KafkaInfrastructureModule (global)
└── RedisInfrastructureModule (global)

OrderPresentationModule imports:
└── OrderApplicationModule
    imports: OrderInfraModule + OrderDomainModule
    provides: CreateOrderUseCase, GetOrderDetailQuery

OrderInfraModule:
    provides: { provide: OrderRepositoryPort, useClass: OrderRepository }
    exports: [OrderRepositoryPort]
```

**Module Rules:**

- ✅ `forRoot()` cho static config; `forRootAsync()` cho async config
- ✅ `forFeature()` cho TypeORM entities trong submodule
- ❌ Không import Infrastructure Module trực tiếp vào Domain Module

### 1.4 Middleware

**Triết lý:**

- Middleware chạy trước Route Handler và Guard
- Dùng cho cross-cutting: logging, request ID, body transformation

**Middleware trong dự án:**

| Middleware                  | Mục đích                                   |
| --------------------------- | ------------------------------------------ |
| `MergeParamsBodyMiddleware` | Merge URL params vào request body          |
| `RequestContextMiddleware`  | Seed AsyncLocalStorage với request context |

**Apply:** qua `MiddlewareConsumer.apply().forRoutes()`
**Order:** `app.use()` globals → Consumer middleware → Guards → Interceptors → Pipes → Route Handler

### 1.5 Exception Filters

**Triết lý:**

- Exception Filter = last-resort handler cho unhandled exceptions
- Trong dự án: dùng `ExceptionInterceptor` (không phải Filter) để transform mọi exception sang API response chuẩn
- `LibExceptionFilter` nhận notification qua Telegram

**Exception Hierarchy:**

```
ExceptionBase
├── DomainException
│   ├── DomainNotFoundException
│   ├── DomainBadRequestException
│   └── DomainInternalException
├── InfrastructureException
│   ├── InfrastructureNotFoundException
│   └── InfrastructureBadRequestException
├── UsecaseException
│   ├── UsecaseBadRequestException
│   ├── UsecaseNotFoundException
│   └── UsecaseInternalException
└── ThirdPartyException
```

**Exception-to-HTTP Mapping:**

| Nhóm lỗi (Exception)      | Custom Classes (`libs/src/common/exceptions`)                                                                                   | HTTP Status        |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| **Not Found**             | `DomainNotFoundException`, `UsecaseNotFoundException`, `InfrastructureNotFoundException`, `ThirdPartyNotFoundException`         | 404                |
| **Bad Request**           | `DomainBadRequestException`, `UsecaseBadRequestException`, `InfrastructureBadRequestException`, `ThirdPartyBadRequestException` | 400                |
| **Internal Server Error** | `DomainInternalException`, `UsecaseInternalServerErrorException`, `ThirdPartyInternalServerErrorException`                      | 500                |
| **Third Party APIs**      | `ThirdPartyException`, `OmsException`, `SfaException`, `EkycException`                                                          | 200 (hoặc tùy API) |
| **Validation**            | `I18nValidationException`                                                                                                       | 400                |

**Rules:**

- ✅ Ném exception đúng layer:
    - **Domain Layer**: `throw new DomainException` (Dùng `domain-exception.handler.ts`)
    - **Application/UseCase Layer**: `throw new UsecaseException` (Dùng `usecase-exception.handler.ts`)
    - **Infrastructure Layer** (DB, Redis, MQ): `throw new InfrastructureException` (Dùng `infrastructure-exception.handler.ts`)
    - **External APIs**: `throw new ThirdPartyException` (Dùng `third-party-exception.handler.ts`)
    - Dùng `exception.handler.ts` cho các trường hợp chung chung chưa xác định rõ layer.
- ❌ Không ném raw `new Error()` trong application/domain
- ❌ Không dùng các HTTP Exception raw của NestJS (`BadRequestException`, `UnauthorizedException`) trong Core Layers (Domain, Application). Phải dùng class tương ứng như `UsecaseBadRequestException` hoặc `UsecaseUnauthorizedException`.
- ❌ Không leak sensitive data trong error message
- ❌ Không leak sensitive data trong error message

### 1.6 Pipes

**Triết lý:**

- Pipe chạy sau Middleware/Guard, trước Route Handler
- Hai vai trò: **Transformation** (string → UUID) và **Validation** (DTO integrity)
- Dự án dùng `I18nValidationPipe` global cho i18n error messages

**Built-in Pipes quan trọng:**

| Pipe                 | Mục đích                            |
| -------------------- | ----------------------------------- |
| `ParseUUIDPipe`      | Validate và parse UUID param        |
| `ParseIntPipe`       | Parse string → number               |
| `ValidationPipe`     | class-validator + class-transformer |
| `I18nValidationPipe` | ValidationPipe + i18n messages      |
| `ParseFilePipe`      | File upload validation              |

**DTO Validation Stack:**

- `class-validator` decorators (`@IsString`, `@IsUUID`, `@IsNotEmpty`, etc.)
- `class-transformer` (`@Type()`, `@Transform()`, `@Expose()`)
- `whitelist: true` — strip unknown properties
- `transform: true` — auto type conversion

### 1.7 Guards

**Triết lý:**

- Guard = authorization gate (yes/no execution decision)
- Chạy SAU Middleware, TRƯỚC Interceptors
- `canActivate()` trả về `boolean | Observable<boolean> | Promise<boolean>`

**Guard Layers:**

| Guard                 | Khi dùng                      |
| --------------------- | ----------------------------- |
| `JwtAuthGuard`        | Verify JWT token từ SSO       |
| `ApiInternalGuard`    | Check `x-api-internal` header |
| `RolesGuard`          | RBAC authorization            |
| `PublicEndpointGuard` | Whitelist public routes       |

**Guard Factory Pattern:**

- `createAppGuard(modulePath)` — tạo guard dynamically theo module context
- Hỗ trợ multi-strategy: Base / Portal / Mobile

**`@Public()` decorator** — bypass auth guard cho public endpoints

### 1.8 Interceptors

**Triết lý:**

- Interceptor bao quanh route handler (AOP — Aspect Oriented Programming)
- Có access vào cả request và response
- Thứ tự: Interceptor binding → handler → Interceptor cleanup

**Interceptors trong dự án:**

| Interceptor                     | Mục đích                            |
| ------------------------------- | ----------------------------------- |
| `ExceptionInterceptor`          | Transform exceptions → API response |
| `TransformUseSentryInterceptor` | Sentry tracing                      |
| `MergeParamsBodyInterceptor`    | Merge params into body              |
| `IdempotencyInterceptor`        | Idempotency key handling            |
| `UploadFileInterceptor`         | File upload binding                 |

**Global Interceptors** — đăng ký qua `APP_INTERCEPTOR` token trong AppModule

### 1.9 Custom Decorators

**Triết lý:**

- Decorator = declarative cross-cutting concern
- Parameter decorator → extract data từ request
- Method decorator → wrap method logic (logging, caching, retry)
- Combine multiple decorators với `applyDecorators()`

**Custom Decorators trong dự án:**

| Decorator                 | Type         | Mục đích                   |
| ------------------------- | ------------ | -------------------------- |
| `@CurrentUser()`          | Parameter    | Extract JWT payload        |
| `@IdempotencyKey()`       | Parameter    | Extract idempotency header |
| `@Public()`               | Route        | Mark public endpoint       |
| `@LogExecution()`         | Method/Class | Execution tracing          |
| `@TrackPerformance()`     | Method       | CPU/Memory metrics         |
| `@Retry()`                | Method       | Auto retry                 |
| `@MergeParamsBody()`      | Controller   | Merge URL params           |
| `@PlainToInstanceQuery()` | Param        | Transform query to DTO     |
| `@SubscribePattern()`     | Method       | Kafka pattern subscription |

---

## PHẦN 2 — FUNDAMENTALS

### 2.1 Custom Providers

**Provider recipes:**

| Recipe        | Syntax                                         | Use Case                |
| ------------- | ---------------------------------------------- | ----------------------- |
| `useClass`    | `{provide: Token, useClass: Impl}`             | Swap implementation     |
| `useValue`    | `{provide: TOKEN, useValue: obj}`              | Config objects, mocks   |
| `useFactory`  | `{provide: TOKEN, useFactory: fn, inject: []}` | Async init, computed    |
| `useExisting` | `{provide: A, useExisting: B}`                 | Alias existing provider |

**Port → Adapter pattern (DDD):**

- `provide: OrderRepositoryPort` (abstract class = injection token)
- `useClass: OrderRepositoryTypeOrm` (concrete implementation)
- Swap implementation bằng cách đổi `useClass` trong module

### 2.2 Asynchronous Providers

**`useFactory` async:**

- Factory return `Promise<T>` — NestJS chờ resolve trước khi inject
- Dùng cho: database connections, config loading, service warm-up
- `inject: [ConfigService]` — inject dependencies vào factory

**`forRootAsync()` Pattern:**

- Standard pattern cho thư viện (`TypeOrmModule.forRootAsync`, `CacheModule.registerAsync`)
- Cho phép inject `ConfigService` để build options

### 2.3 Dynamic Modules

**Triết lý:**

- Dynamic Module = module config khác nhau theo caller context
- `forRoot()` — singleton global config
- `forFeature()` — per-feature config (TypeORM entities, BullMQ queues)
- `forRootAsync()` — async config với inject

**Pattern trong dự án:**

- `TypeOrmModule.forFeature([OrderOrmEntity])` — register ORM entity
- `LibReportModule.forRootAsync({useFactory, inject})` — async config
- `BullModule.registerQueue({name: 'order-processing'})` — queue registration

### 2.4 Injection Scopes

**Scope types:**

| Scope                 | Lifetime      | Khi dùng                          |
| --------------------- | ------------- | --------------------------------- |
| `DEFAULT` (Singleton) | App lifetime  | Services, Repositories, Use Cases |
| `REQUEST`             | Per request   | RequestContext, User-specific     |
| `TRANSIENT`           | Per injection | Stateful one-shot                 |

**Rule:** Mặc định dùng Singleton. REQUEST scope chỉ khi thực sự cần per-request state (và biết performance cost).

### 2.5 Circular Dependency

**Nguyên nhân:** Module A imports B, B imports A.

**Giải pháp:**

1. `forwardRef(() => ModuleB)` — delay resolution
2. Refactor — extract shared dependency sang module C
3. Event-driven — thay direct import bằng domain events

**Rule: Tránh circular deps.** Nếu cần `forwardRef`, đây là dấu hiệu design cần refactor.

### 2.6 Module Reference

**`ModuleRef`** — programmatic provider lookup:

- `moduleRef.get(Token)` — get singleton
- `moduleRef.resolve(Token)` — resolve scoped provider
- `moduleRef.create(Class)` — create transient instance

**Khi dùng:** Dynamic use case dispatch, plugin systems, runtime provider selection.

### 2.7 Lazy-Loading Modules

**`LazyModuleLoader`** — load module on-demand:

- Giảm startup time
- Dùng cho: heavy modules không cần at startup (Kafka consumer khi worker mode)
- `lazyModuleLoader.load(() => import('./heavy.module'))` returns `LazyModuleLoaderRef`

### 2.8 Execution Context

**`ExecutionContext`** — adapter cho request context:

- `context.switchToHttp()` → `getRequest()`, `getResponse()`
- `context.switchToWs()` → WebSocket context
- `context.switchToRpc()` → Microservice context
- `context.getHandler()` → Route handler method
- `context.getClass()` → Controller class

**`Reflector`** — đọc metadata từ decorators:

- `reflector.get(MetadataKey, handler)`
- `reflector.getAllAndMerge(key, [handler, controller])`

### 2.9 Lifecycle Events

**Lifecycle Order:**

```
onModuleInit() → onApplicationBootstrap()
    (Application running)
onModuleDestroy() → beforeApplicationShutdown() → onApplicationShutdown()
```

**Interfaces:**

| Interface                   | Method                              | Khi dùng                 |
| --------------------------- | ----------------------------------- | ------------------------ |
| `OnModuleInit`              | `onModuleInit()`                    | Setup sau DI resolved    |
| `OnApplicationBootstrap`    | `onApplicationBootstrap()`          | Sau toàn bộ module init  |
| `OnModuleDestroy`           | `onModuleDestroy()`                 | Cleanup resources        |
| `BeforeApplicationShutdown` | `beforeApplicationShutdown(signal)` | Graceful shutdown signal |
| `OnApplicationShutdown`     | `onApplicationShutdown(signal)`     | Final cleanup            |

**Ví dụ:** `DomainEventDispatcher.setDomainEventPublisherService()` trong `LibDDDModule.onModuleInit()`

### 2.10 Discovery Service

**`DiscoveryService`** (từ `@nestjs/core`):

- Scan all providers/controllers tại runtime
- Dùng để: plugin registration, saga registration, event handler discovery
- `discoveryService.getProviders()` → `InstanceWrapper[]`
- Combine với `Reflector` để đọc metadata từ providers

**Saga Registration Pattern:**

- `SagaManager` dùng DiscoveryService để auto-register tất cả `SagaDefinition` subclasses

### 2.11 Platform Agnosticism

**NestJS hỗ trợ:**

- **Express** (default) — battle-tested, ecosystem phong phú
- **Fastify** — 2x performance, không có Express middleware compatibility

**Adapter:** `NestExpressApplication` vs `NestFastifyApplication`

**Rule:** Dự án dùng Express. Chỉ migrate Fastify khi có benchmark evidence rõ ràng. Không mix Express middleware khi đã switch sang Fastify.

### 2.12 Testing

**Testing Strategy (80% coverage minimum):**

| Test Type   | Tool                     | Scope                                     |
| ----------- | ------------------------ | ----------------------------------------- |
| Unit        | Jest + `@nestjs/testing` | Domain entities, value objects, use cases |
| Integration | Jest + `TestingModule`   | Module interactions, repository           |
| E2E         | Jest + Supertest         | Full HTTP flow                            |

**Testing Philosophy:**

- TDD: RED → GREEN → REFACTOR
- Mock Infrastructure layer khi test Application/Domain
- Dùng `TestingModule.createTestingModule()` — không spin up toàn bộ app
- In-memory implementations cho repositories trong unit tests

**Mock Patterns:**

- `jest.fn()` cho simple mocks
- Manual mocks cho complex dependencies
- Spy với `jest.spyOn()` cho partial mocking

---

## PHẦN 3 — TECHNIQUES

### 3.1 Configuration

**Philosophy:**

- Mọi config qua environment variables — không hardcode
- `ConfigModule.forRoot()` với Joi validation schema
- `ConfigService.get<T>('KEY')` — type-safe access
- `expandVariables: true` — hỗ trợ `${OTHER_VAR}` syntax

**Config Categories:**

| Category | Keys                                                           |
| -------- | -------------------------------------------------------------- |
| Server   | `PORT`, `NODE_ENV`                                             |
| Database | `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`      |
| Redis    | `REDIS_HOST`, `REDIS_PORT`, `REDIS_USERNAME`, `REDIS_PASSWORD` |
| Kafka    | `KAFKA_DEFAULT_CLIENT_ID`, `KAFKA_DEFAULT_BROKER`              |
| S3       | `S3_PUBLIC_*`, `S3_PRIVATE_*`                                  |
| Auth     | `JWT_SECRET`, `SSO_URL`                                        |
| Sentry   | `SENTRY_DSN`                                                   |

**Environment Files:** `getEnvFilePath(NODE_ENV)` resolve đúng `.env.development`, `.env.production`

### 3.2 Database (TypeORM)

**Philosophy:**

- ORM Entity ≠ Domain Entity — luôn tách biệt
- Repository Pattern với BaseRepositoryTypeORM
- Mapper tách biệt transform logic
- Transaction qua `@TypeOrmTransaction()` decorator
- Optimistic locking qua `@VersionColumn()`

**ORM Entity Rules:**

| Rule                                                               | Lý do                               |
| ------------------------------------------------------------------ | ----------------------------------- |
| `@PrimaryColumn('uuid')` thay vì `@PrimaryGeneratedColumn('uuid')` | Domain tạo ID, không DB             |
| `@Index(['customerId', 'status'])`                                 | Performance cho queries thường dùng |
| `@VersionColumn()`                                                 | Optimistic locking                  |
| `@CreateDateColumn()` / `@UpdateDateColumn()`                      | Auto timestamp                      |

**Query Builder:** Dùng khi cần JOIN phức tạp, aggregation, hoặc select specific columns.

**Migration:** Schema changes qua TypeORM migrations, không dùng `synchronize: true` production.

### 3.3 MongoDB (Mongoose)

**Khi dùng:** Event Store, Audit Log, Flexible Schema, Time-Series data.

**Pattern:**

- Schema trong `@Schema()` class
- Repository Pattern tương tự TypeORM
- `@InjectModel(ModelName)` để inject model

### 3.4 Validation

**Validation Stack:**

```
Request → I18nValidationPipe → class-validator → class-transformer → DTO instance
```

**Validation Rules:**

| Layer       | Tool                       | Scope                   |
| ----------- | -------------------------- | ----------------------- |
| HTTP Input  | class-validator decorators | DTO validation          |
| Domain      | `Guard` class              | Value object invariants |
| Application | `Result<T>`                | Use case validation     |
| Database    | TypeORM constraints        | Data integrity          |

**i18n Validation:**

- Lỗi validation tự động translate theo `Accept-Language` header
- Error messages trong `src/i18n/` directory

### 3.5 Caching

**Caching Strategy:**

| Layer         | Tool                      | TTL Strategy      |
| ------------- | ------------------------- | ----------------- |
| Method-level  | `@RedisCache()` decorator | Per-method config |
| HTTP Response | CacheInterceptor          | Endpoint-level    |
| Application   | Manual Redis client       | Custom logic      |

**Redis Cache Decorator Rules:**

- `mode: 'GET'` — check cache first, miss → execute → set
- `mode: 'SET'` — execute → clear related patterns
- `relatedPatterns[]` — invalidate related cache keys on mutation
- Key format: `{prefix}:{param1}:{param2}`
- Async error — non-blocking, log only

**Cache Key Convention:** `{domain}:{operation}:{id}` (e.g., `order:detail:uuid-xxx`)

### 3.6 Serialization

**Philosophy:**

- Response transformation qua `@Expose()` + `ClassSerializerInterceptor`
- Hoặc explicit Mapper `.toDto()` method
- Không expose internal domain fields trực tiếp

**Tools:**

- `class-transformer`: `@Expose()`, `@Exclude()`, `@Transform()`
- `@ApiProperty()`: Swagger documentation
- `@AutoApiProperty()`: Auto-generate từ TypeScript types

**Rule:** Dùng explicit Mapper (`toDtoMany`) thay vì auto-serialization khi cần fine-grained control.

### 3.7 Versioning

**Strategy:** URI Versioning — `/v1/orders`, `/v2/orders`

**Config:**

```
enableVersioning({ defaultVersion: '1', type: VersioningType.URI })
```

**Controller Level:** `@Controller({path: 'orders', version: '2'})`
**Route Level:** `@Version('2')` trên specific endpoint

**Deprecation Strategy:**

- Maintain v1 với deprecated notice ít nhất 1 sprint
- Document breaking changes
- Swagger document cả hai versions

### 3.8 Task Scheduling

**Tools:**

- `@nestjs/schedule` — Cron jobs
- BullMQ — Reliable background jobs với retry
- `@nestjs/bull` — Bull queue integration

**Scheduling Patterns:**

| Pattern        | Tool                          | Use Case                          |
| -------------- | ----------------------------- | --------------------------------- |
| Cron           | `@Cron()` / `@Interval()`     | Periodic tasks (reports, cleanup) |
| Delayed Job    | BullMQ `delay` option         | Send notification after 24h       |
| Retry Queue    | BullMQ `attempts` + `backoff` | Transient failure retry           |
| Priority Queue | BullMQ `priority`             | Important jobs first              |

**Rule:** Production jobs chạy qua BullMQ (persistent, retry-able), không phải in-memory cron.

### 3.9 Queues (BullMQ)

**Queue Setup:**

- `BullModule.forRoot()` — Redis connection
- `BullModule.registerQueue({name})` — register queue
- `@Processor(queueName)` — worker class
- `@Process(jobName)` — job handler

**Job Options:**

- `attempts: 3` — retry count
- `backoff: {type: 'exponential', delay: 2000}` — backoff strategy
- `removeOnComplete: 100` — cleanup after success
- `removeOnFail: 500` — retain failed jobs for debugging
- `delay` — scheduled execution

**Concurrency:** `@Processor({name, concurrency: 5})` — parallel workers

### 3.10 Logging

**Logger Stack:**

- Winston — structured, multi-transport logging
- Log4js — alternative (cronjob context)
- Sentry — error tracking + distributed tracing

**Log Levels:** `error`, `warn`, `info`, `debug`, `verbose`

**Structured Log Format:**

```json
{
    "level": "info",
    "message": "Order created",
    "context": "CreateOrderUseCase",
    "requestId": "uuid",
    "duration_ms": 142,
    "timestamp": "ISO-8601"
}
```

**Rules:**

- ✅ Dùng `Logger` từ `@nestjs/common` trong use cases
- ✅ Log ở START và END của use case execution (`@LogExecution()`)
- ❌ Không log sensitive data (passwords, tokens, PII)
- ❌ Không dùng `console.log` trong production code

**`@LogExecution()` decorator** — tự động log method entry/exit + duration

### 3.11 Cookies

**Setup:** `cookie-parser` middleware
**Signed Cookies:** `app.use(cookieParser(SECRET))`
**Access:** `@Req() req` → `req.cookies` hoặc `req.signedCookies`

**Security Rules:**

- `httpOnly: true` — prevent XSS
- `secure: true` — HTTPS only
- `sameSite: 'strict'` — CSRF protection
- Signed cookies cho sensitive data

### 3.12 Events (EventEmitter)

**Local Events:** `@nestjs/event-emitter` — in-process only
**Distributed Events:** Kafka Event Bus (`KafkaEventBus`) — cross-service

**Domain Event Flow:**

```
Aggregate.addDomainEvent()
  → Repository.save()
  → DomainEventDispatcher.dispatchEventsForAggregate()
  → DomainEventPublisherService.emitAllEvent()
  → KafkaEventBus.publish()
  → Kafka Topic
  → Consumer service
```

**Local EventEmitter** dùng cho: notification triggers, audit logging, sync side effects trong same process.

### 3.13 Compression

**Setup:** `compression` middleware — Brotli/Gzip

**Rule:** Chỉ compress response size > 1KB. Static assets nên serve qua CDN với pre-compressed files.

### 3.14 File Upload

**Stack:**

- `multer` — multipart form data
- `@UploadedFile()` / `@UploadedFiles()` — parameter decorator
- `ParseFilePipe` — validate type, size
- S3 Adapter — upload to object storage

**Flow:**

```
Client → Multer middleware → Validation pipe → Controller → Use Case → S3 Adapter
```

**Rules:**

- ✅ Validate file type bằng MIME type, không phải extension
- ✅ Limit file size (default 50MB trong dự án)
- ✅ Stream large files — không load toàn bộ vào memory
- ❌ Không store files trong local filesystem production

### 3.15 Streaming Files

**Tools:**

- `StreamableFile` — NestJS built-in
- Node.js `ReadStream` / `PassThrough`
- RxJS `Observable<Buffer>` cho reactive streaming

**Report Streaming Pattern:**

- `StreamPipelineService` — process large datasets
- Async generator pattern cho memory-efficient processing
- Response `Content-Type: application/octet-stream` + `Content-Disposition`

### 3.16 HTTP Module

**`HttpModule`** (`@nestjs/axios`) — HTTP client cho external APIs

**Resilience Pattern (MANDATORY cho external calls):**

- Circuit Breaker (`opossum`)
- `@Retry(options)` decorator — exponential backoff
- Timeout via `timeout(5000)` RxJS operator

**HTTP Client Rules:**

- ✅ Circuit Breaker cho tất cả external calls
- ✅ Retry ≤ 3 lần với exponential backoff
- ✅ Timeout ≤ 5s cho synchronous flows
- ❌ Không retry non-idempotent operations (POST tạo resource)

### 3.17 Session

**Dự án:** Stateless JWT — không dùng server-side session
**Redis Session:** Nếu cần stateful → `express-session` + `connect-redis`
**Alternative:** Cookie-based session với signed cookies

### 3.18 Model-View-Controller (MVC)

**Dùng khi:** Server-side rendering cho admin panels, email templates

**Stack:**

- `@Render('template')` decorator
- Handlebars/EJS engine
- `BaseAssetsModule.setupAssets(app)` — static asset serving
- Views trong `src/views/`

**Rule:** Dự án ưu tiên REST API + Frontend separation. MVC chỉ cho internal tools, admin panels.

### 3.19 Performance (Fastify)

**Khi migrate sang Fastify:**

- `NestFastifyApplication` thay vì `NestExpressApplication`
- `FastifyAdapter` từ `@nestjs/platform-fastify`
- Không dùng Express-specific middleware
- Fastify plugins thay vì Express middleware

**Rule hiện tại:** Dự án dùng Express. Benchmark trước khi migrate.

### 3.20 Server-Sent Events (SSE)

**Stack:**

- `@Sse(path)` decorator
- Response type `Observable<MessageEvent>`
- `EventSource` client-side

**SSE Module trong dự án:**

- `SseModule` trong `libs/src/core/sse/`
- Subscriber management
- Heartbeat support
- Connection lifecycle management

**Use Cases:** Real-time job progress, notifications, dashboard updates.

---

## PHẦN 4 — SECURITY

### 4.1 Authentication

**Flow:**

```
Request
→ Auth Middleware (extract token)
→ JwtAuthGuard.canActivate()
→ SSO Verify
→ Store user in CLS
→ @CurrentUser() in handler
```

**SSO Integration:**

- Centralized token verification
- Multi-tenant support
- Strategy: Portal / Mobile / Internal

**Token Storage:** ClsService (request-scoped AsyncLocalStorage) — thread-safe

### 4.2 Authorization

**Layers:**

| Layer           | Mechanism                          |
| --------------- | ---------------------------------- |
| Route           | `@UseGuards(JwtAuthGuard)`         |
| Public endpoint | `@Public()` whitelist              |
| Role-based      | `@Roles(Role.Admin)` + RolesGuard  |
| Resource-based  | Use Case level check (owner check) |
| Internal API    | `x-api-internal` header            |

**RBAC Rules:**

- Roles định nghĩa trong domain (không hardcode string)
- Resource ownership check trong Use Case, không trong Controller

### 4.3 Encryption and Hashing

**Hashing:**

- `bcrypt` — password hashing (cost factor ≥ 12)
- `argon2` — modern alternative
- Never store plaintext passwords

**Encryption:**

- `crypto` built-in — AES-256-GCM cho sensitive data at rest
- HTTPS enforced at infrastructure level

**Token Signing:**

- JWT RS256 (asymmetric) thay vì HS256 cho production
- Short expiry + refresh token rotation

### 4.4 Helmet

**`helmet` middleware** — set security HTTP headers:

| Header                      | Protection     |
| --------------------------- | -------------- |
| `X-Frame-Options`           | Clickjacking   |
| `X-XSS-Protection`          | XSS (legacy)   |
| `Content-Security-Policy`   | XSS, injection |
| `Strict-Transport-Security` | Force HTTPS    |
| `X-Content-Type-Options`    | MIME sniffing  |
| `Referrer-Policy`           | Info leakage   |

**Config:** `app.use(helmet())` — default headers. Customize CSP cho specific needs.

### 4.5 CORS

**Config:**

- `app.use(cors())` — enable CORS
- `origin: ['https://app.domain.com']` — whitelist origins
- `credentials: true` — for cookie-based auth
- `methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']`

**Rule:** Không dùng `origin: '*'` production với credential APIs.

### 4.6 CSRF Protection

**Strategy:**

- Stateless JWT APIs — CSRF protection không cần thiết nếu không dùng cookie auth
- Cookie-based auth → `csurf` middleware bắt buộc
- `sameSite: 'strict'` cookie flag giảm thiểu CSRF risk

### 4.7 Rate Limiting

**Stack:** `@nestjs/throttler`

**Strategy:**

| Endpoint      | Limit        |
| ------------- | ------------ |
| Login/Auth    | 5 req/min    |
| Public APIs   | 100 req/min  |
| Internal APIs | 1000 req/min |
| File Upload   | 10 req/min   |

**Redis-backed throttling** cho distributed rate limiting (multi-instance)

---

## PHẦN 5 — WEBSOCKETS

### 5.1 Gateways

**Architecture:**

- Gateway = WebSocket controller analog
- `@WebSocketGateway(port, options)` — declare gateway
- `@WebSocketServer()` — inject server instance
- `@SubscribeMessage(event)` — handle client messages
- Implement `OnGatewayInit`, `OnGatewayConnection`, `OnGatewayDisconnect`

**Redis Adapter (Dự án):**

- `RedisIoAdapter` — scale WebSocket across multiple instances
- Redis pub/sub để broadcast events
- `app.useWebSocketAdapter(redisIoAdapter)`

**Namespace Strategy:**

- `/admin` — admin dashboard
- `/user` — user-specific events
- Room-based targeting: `socket.join(roomId)`

### 5.2 WebSocket Exception Filters

- Implement `WsExceptionFilter` — handle WebSocket exceptions
- `WsException` — WebSocket-specific exception class
- Response via `socket.emit('error', errorObj)`

### 5.3 WebSocket Pipes

- Apply với `@UsePipes()` trên `@SubscribeMessage()` handler
- Validate WebSocket payload (same as HTTP pipes)

### 5.4 WebSocket Guards

- Implement `CanActivate` — check WebSocket auth
- Access token qua handshake: `client.handshake.auth.token`
- Return `false` → disconnect client

### 5.5 WebSocket Interceptors

- Wrap WebSocket message handling
- Logging, metrics, transform response format

### 5.6 WebSocket Adapters

**Dự án dùng:** `socket.io` với Redis adapter

- `createIoServer()` override cho Redis pub/sub
- `@nestjs/platform-socket.io` default
- Alternative: `ws` (raw WebSocket, lighter)

---

## PHẦN 6 — MICROSERVICES

### 6.1 Overview

**Hybrid App Pattern (Dự án):**

```typescript
Promise.race([
    app.startAllMicroservices(), // Kafka consumer
    app.listen(PORT), // HTTP server
]);
```

**Communication Patterns:**

| Pattern          | Tool                    | Use Case                     |
| ---------------- | ----------------------- | ---------------------------- |
| Request-Response | Kafka/RabbitMQ `send()` | Synchronous cross-service    |
| Event-Driven     | Kafka `emit()`          | Domain events, notifications |
| Saga             | SagaManager             | Distributed transactions     |
| Outbox           | OutboxProcessor         | Guaranteed delivery          |

### 6.2 RabbitMQ

**Config:**

- `RabbitMQInfrastructureModule` — global setup
- Exchange types: `direct`, `topic`, `fanout`
- Wildcard routing: `order.*` patterns

**Consumer Pattern:**

- `@SubscribeEventPattern(pattern)` — custom decorator
- `@Payload()` — extract message body
- `@Ctx()` — extract context (headers, ack/nack)
- Manual ack/nack cho reliable processing

**Publisher Pattern:**

- `rabbitmq.publish(exchange, routingKey, payload)`
- Persistent messages: `persistent: true`
- Correlation ID cho request-response

### 6.3 Kafka

**Config:**

- `KafkaInfrastructureModule` — global setup
- Multi-client: primary + secondary clusters
- Consumer groups cho horizontal scaling

**Topic Naming Convention:**

```
{service}.{domain}.{event}
EventBusKafka.ORDER_CREATED
EventBusKafka.DOMAIN_EVENTS
{sagaType}-reply
```

**Saga Kafka Pattern:**

- Orchestrator publishes commands to Kafka
- Participants subscribe to their topic
- Participants publish replies to `{sagaType}-reply` topic
- `SagaCommandHeaders` — saga context metadata

**Event Bus Pattern:**

- `KafkaEventBus.publish(event)` — single event
- `KafkaEventBus.publishAll(events[])` — batch
- `EVENT_TOPIC_MAP` — per-event topic routing

### 6.4 Microservice Interceptors

- Apply với `@UseInterceptors()` trên `@MessagePattern()` / `@EventPattern()` handlers
- Logging, metrics, error transformation
- Same interface as HTTP interceptors

---

## PHẦN 7 — OPENAPI

### 7.1 Introduction

**Swagger Setup:**

- `SwaggerServiceAdapter` — centralized Swagger config
- Multiple Swagger endpoints: `/doc`, `/integrate-document`
- Themes: DARK/LIGHT
- `@ApiTags()` — group endpoints

**Swagger Module:**

```
SwaggerModule.setup(path, app, document, options)
```

### 7.2 Types and Parameters

**Decorators:**

| Decorator                | Mục đích           |
| ------------------------ | ------------------ |
| `@ApiProperty()`         | Document DTO field |
| `@ApiPropertyOptional()` | Optional field     |
| `@ApiQuery()`            | Query parameter    |
| `@ApiParam()`            | Path parameter     |
| `@ApiHeader()`           | Request header     |
| `@ApiBody()`             | Request body       |

**Rules:**

- `@ApiProperty({description, example, type})` — đầy đủ thông tin
- Nested DTOs cần `@ApiProperty({type: () => NestedDto})`
- Enums: `@ApiProperty({enum: MyEnum, enumName: 'MyEnum'})`

### 7.3 Operations

| Decorator                                   | Mục đích               |
| ------------------------------------------- | ---------------------- |
| `@ApiOperation({summary, description})`     | Endpoint description   |
| `@ApiResponse({status, type, description})` | Response documentation |
| `@ApiOkResponse({type})`                    | 200 response           |
| `@ApiCreatedResponse({type})`               | 201 response           |
| `@ApiNotFoundResponse()`                    | 404 response           |

**Dự án dùng:**

- `@SwaggerApiResponse()` — custom decorator combine multiple responses
- Standard error responses: 400, 401, 404, 500, 502

### 7.4 Security

**Bearer Token:**

```
@ApiBearerAuth()
```

**API Key:**

```
@ApiSecurity('x-api-internal')
```

**Global setup:** `addBearerAuth()` trong Swagger config

### 7.5 Mapped Types

**`PartialType(CreateDto)`** — all fields optional (Update DTOs)
**`PickType(CreateDto, ['field1', 'field2'])`** — select fields
**`OmitType(CreateDto, ['field1'])`** — exclude fields
**`IntersectionType(Dto1, Dto2)`** — merge DTOs

**Convention:**

- `CreateOrderDto` — full required
- `UpdateOrderDto` extends `PartialType(CreateOrderDto)`

### 7.6 Decorators (CLI Plugin)

**`@nestjs/swagger` CLI plugin** — auto-generate `@ApiProperty()` từ TypeScript types

**tsconfig Plugin:**

```json
"plugins": [
    {"name": "@nestjs/swagger", "options": {"classValidatorShim": true}}
]
```

**`@AutoApiProperty()`** — dự án custom decorator tương tự

### 7.7 Other Features

**Extra Models:** Register DTOs không trực tiếp trong routes

```
extraModels: [PaginatedResponseDto, ErrorResponseDto]
```

**Document Builder:**

- Title, description, version, tags
- Server URLs for different environments
- Contact + License info

**Swagger Collection Pattern (Dự án):**

- `SwaggerCollection` — group related routes
- Per-module Swagger setup
- Include/exclude specific routes

---

## QUICK REFERENCE — Decision Tables

### Khi nào dùng gì?

| Tình huống                   | Giải pháp                          |
| ---------------------------- | ---------------------------------- |
| Validate HTTP input          | `class-validator` trong DTO        |
| Validate domain invariant    | `Guard` class + `Result<T>`        |
| Handle cross-cutting concern | Interceptor hoặc Decorator         |
| Authorization check          | Guard (`canActivate`)              |
| Transform request param      | Pipe                               |
| Async initialization         | `useFactory` + `async`             |
| Per-request state            | `ClsService` / `AsyncLocalStorage` |
| External API call            | `HttpModule` + Circuit Breaker     |
| Distributed transaction      | Saga Pattern                       |
| Guaranteed event delivery    | Outbox Pattern                     |
| Background job               | BullMQ                             |
| Real-time push               | WebSocket (Socket.io + Redis)      |
| One-way server push          | SSE                                |
| Cache expensive query        | `@RedisCache()` decorator          |
| Schema-less data             | MongoDB                            |
| Relational + ACID            | PostgreSQL + TypeORM               |

### Layer Dependency Rules

| From → To                    | Allowed?              |
| ---------------------------- | --------------------- |
| Presentation → Application   | ✅                    |
| Application → Domain         | ✅                    |
| Infrastructure → Domain      | ✅ (implements ports) |
| Domain → Application         | ❌                    |
| Domain → Infrastructure      | ❌                    |
| Domain → NestJS/TypeORM      | ❌                    |
| Application → Infrastructure | ❌ (only via ports)   |

### Exception Throw Rules

| Layer           | Exception Class           |
| --------------- | ------------------------- |
| Domain          | `DomainException`         |
| Use Case        | `UsecaseException`        |
| Infrastructure  | `InfrastructureException` |
| Third-party     | `ThirdPartyException`     |
| HTTP validation | `class-validator` (auto)  |

### Performance Checklist

| Check             | Tool                    |
| ----------------- | ----------------------- |
| N+1 queries?      | QueryBuilder with joins |
| Repeated lookups? | Map/Set index           |
| Expensive method? | `@RedisCache()`         |
| External call?    | Circuit Breaker + Retry |
| Large response?   | Streaming + Pagination  |
| Background work?  | BullMQ queue            |
| WebSocket scale?  | Redis pub/sub adapter   |

### Security Checklist

| Check              | Implementation              |
| ------------------ | --------------------------- |
| Input validated    | `I18nValidationPipe` global |
| Auth on all routes | `@UseGuards(JwtAuthGuard)`  |
| Sensitive headers  | `helmet()`                  |
| CORS configured    | Whitelist origins           |
| Rate limited       | `@Throttle()`               |
| Secrets in env     | `ConfigService`             |
| Idempotent writes  | `IdempotencyInterceptor`    |
| No SQL injection   | TypeORM parameterized       |
| Audit trail        | Domain events               |
