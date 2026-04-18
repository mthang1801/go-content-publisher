# NestJS Backend Patterns — Code Examples

> **Copy-paste ready.** Mọi triết lý và rules → xem `SKILL.md`.

---

## PHẦN 1 — OVERVIEW

### 1.1 Controllers — Command + Query Split

```typescript
// src/presentation/portal/order/controllers/order-command.controller.ts
import {
    Body,
    Controller,
    Delete,
    HttpCode,
    HttpStatus,
    Param,
    ParseUUIDPipe,
    Post,
} from '@nestjs/common';
import { ApiOperation, ApiTags } from '@nestjs/swagger';
import { CurrentUser, IdempotencyKey, Public } from '@common/decorators';
import { JwtAuthGuard } from '@common/guards';
import { UseGuards, UseInterceptors } from '@nestjs/common';
import { IdempotencyInterceptor } from '@common/interceptors';
import { CreateOrderUseCase, CancelOrderUseCase } from '@application/order/use-cases';
import { CreateOrderDto, CancelOrderDto } from '../dtos';
import { CreateOrderResponse } from '@application/order/responses';
import { UserPayload } from '@common/types';

@Controller('orders')
@ApiTags('Orders')
@UseGuards(JwtAuthGuard)
export class OrderCommandController {
    constructor(
        private readonly createOrder: CreateOrderUseCase,
        private readonly cancelOrder: CancelOrderUseCase,
    ) {}

    @Post()
    @HttpCode(HttpStatus.CREATED)
    @UseInterceptors(IdempotencyInterceptor)
    @ApiOperation({ summary: 'Tạo đơn hàng mới' })
    async create(
        @Body() dto: CreateOrderDto,
        @CurrentUser() user: UserPayload,
        @IdempotencyKey() idempotencyKey: string,
    ): Promise<CreateOrderResponse> {
        return this.createOrder.executeWithHooks({
            ...dto,
            idempotencyKey,
            userId: user.id,
        });
    }

    @Delete(':id')
    @HttpCode(HttpStatus.NO_CONTENT)
    async cancel(
        @Param('id', ParseUUIDPipe) id: string,
        @Body() dto: CancelOrderDto,
        @CurrentUser() user: UserPayload,
    ): Promise<void> {
        await this.cancelOrder.executeWithHooks({ orderId: id, ...dto, userId: user.id });
    }
}

// src/presentation/portal/order/controllers/order-query.controller.ts
@Controller('orders')
@ApiTags('Orders')
@UseGuards(JwtAuthGuard)
export class OrderQueryController {
    constructor(private readonly getOrderDetail: GetOrderDetailQuery) {}

    @Get(':id')
    async getDetail(@Param('id', ParseUUIDPipe) id: string): Promise<OrderDetailResponse> {
        return this.getOrderDetail.queryWithHooks({ orderId: id });
    }

    @Get()
    async list(@Query() query: ListOrdersQueryDto): Promise<PaginatedOrderResponse> {
        return this.listOrders.queryWithHooks(query);
    }
}
```

### 1.2 Providers — Custom Token + Port/Adapter

```typescript
// src/infrastructure/modules/order/order.infra.module.ts
import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { OrderRepositoryPort } from '@domain/order/repositories/order.repository.port';
import { OrderRepository } from './repositories/order.repository';
import { OrderMapper } from './mappers/order.mapper';
import { OrderOrmEntity } from './entities/order.orm.entity';
import { OrderDomainModule } from '@domain/order/order.domain.module';

@Module({
    imports: [TypeOrmModule.forFeature([OrderOrmEntity]), OrderDomainModule],
    providers: [
        OrderMapper,
        {
            provide: OrderRepositoryPort, // Abstract class as token
            useClass: OrderRepository, // Concrete implementation
        },
    ],
    exports: [OrderRepositoryPort],
})
export class OrderInfraModule {}

// --- Config value provider ---
@Module({
    providers: [
        {
            provide: 'APP_CONFIG',
            useValue: { maxRetries: 3, timeout: 5000 },
        },
    ],
})
export class ConfigDemoModule {}

// --- Factory provider (async) ---
@Module({
    providers: [
        {
            provide: 'DATABASE_CONNECTION',
            useFactory: async (configService: ConfigService) => {
                return createConnection({
                    host: configService.get('DB_HOST'),
                    port: configService.get('DB_PORT'),
                });
            },
            inject: [ConfigService],
        },
    ],
})
export class DatabaseDemoModule {}
```

### 1.3 Modules — DDD Layered Module

```typescript
// src/application/order/order.application.module.ts
@Module({
    imports: [
        OrderInfraModule, // Provides: OrderRepositoryPort → OrderRepository
        OrderDomainModule, // Exports: OrderDomainService, OrderFactory, PricingService
    ],
    providers: [CreateOrderUseCase, GetOrderDetailQuery, CancelOrderUseCase],
    exports: [CreateOrderUseCase, GetOrderDetailQuery, CancelOrderUseCase],
})
export class OrderApplicationModule {}

// src/presentation/presentation.module.ts
@Module({
    imports: [OrderApplicationModule, ProductApplicationModule],
    controllers: [OrderCommandController, OrderQueryController],
})
export class PresentationModule {}

// --- Dynamic module with forRootAsync ---
@Module({})
export class NotificationModule {
    static forRootAsync(options: NotificationModuleAsyncOptions): DynamicModule {
        return {
            module: NotificationModule,
            imports: options.imports || [],
            providers: [
                {
                    provide: NOTIFICATION_OPTIONS,
                    useFactory: options.useFactory,
                    inject: options.inject || [],
                },
                NotificationService,
            ],
            exports: [NotificationService],
        };
    }
}
```

### 1.4 Middleware

```typescript
// libs/src/common/middleware/request-context.middleware.ts
import { Injectable, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';
import { RequestContextStorage } from '@core/async-local-storage';
import { v7 as uuidv7 } from 'uuid';

@Injectable()
export class RequestContextMiddleware implements NestMiddleware {
    use(req: Request, res: Response, next: NextFunction) {
        const session = {
            sessionId: uuidv7(),
            ip: req.ip,
            authorization: req.headers.authorization,
            context: { url: req.url, method: req.method },
        };

        RequestContextStorage.run(session, () => next());
    }
}

// libs/src/common/middleware/merge-params-body.middleware.ts
@Injectable()
export class MergeParamsBodyMiddleware implements NestMiddleware {
    use(req: Request, res: Response, next: NextFunction) {
        req.body = { ...req.body, ...req.params };
        next();
    }
}

// Apply in AppModule:
export class AppModule implements NestModule {
    configure(consumer: MiddlewareConsumer) {
        consumer.apply(MergeParamsBodyMiddleware, RequestContextMiddleware).forRoutes('*');

        consumer
            .apply(AuthMiddleware)
            .exclude({ path: 'health', method: RequestMethod.GET })
            .forRoutes('*');
    }
}
```

### 1.5 Exception Filters

```typescript
// libs/src/common/exceptions/exception.interceptor.ts
import { Injectable, NestInterceptor, ExecutionContext, CallHandler } from '@nestjs/common';
import { catchError, Observable, throwError } from 'rxjs';
import { ExceptionBase } from './exception.base';

@Injectable()
export class ExceptionInterceptor implements NestInterceptor {
    intercept(context: ExecutionContext, next: CallHandler): Observable<unknown> {
        return next.handle().pipe(
            catchError((error: unknown) => {
                const response = context.switchToHttp().getResponse();
                const handler = this.getHandler(error);
                const { statusCode, body } = handler.getResponseError(error);

                if (process.env.NODE_ENV !== 'development') {
                    delete body.stack;
                }

                response.status(statusCode).json(body);
                return throwError(() => error);
            }),
        );
    }

    private getHandler(error: unknown) {
        if (error instanceof UsecaseException) return new ExceptionUsecaseHandler();
        if (error instanceof DomainException) return new ExceptionDomainHandler();
        if (error instanceof InfrastructureException) return new ExceptionInfrastructureHandler();
        return new GenericExceptionHandler();
    }
}

// libs/src/common/exceptions/exception.base.ts
export abstract class ExceptionBase extends Error {
    abstract code: string;
    abstract errorCode: string;

    constructor(
        readonly message: string,
        readonly errors: string[] = [],
        readonly stack?: string,
    ) {
        super(message);
        if (stack) this.stack = stack;
        else Error.captureStackTrace(this, this.constructor);
    }
}

// Usage: throw from use-case
throw new UsecaseBadRequestException(['Order must be in PENDING status to cancel']);
throw new UsecaseNotFoundException(`Order ${orderId} not found`);
```

### 1.6 Pipes

```typescript
// libs/src/common/pipes/i18n-validation.pipe.ts
import { Injectable } from '@nestjs/common';
import { I18nValidationPipe as BaseI18nValidationPipe } from 'nestjs-i18n';
import { HttpStatus } from '@nestjs/common';

@Injectable()
export class AppValidationPipe extends BaseI18nValidationPipe {
    constructor() {
        super({
            transform: true,
            transformOptions: { enableImplicitConversion: true },
            whitelist: true,
            forbidNonWhitelisted: false,
            errorHttpStatusCode: HttpStatus.BAD_REQUEST,
        });
    }
}

// main.ts - global pipe
app.useGlobalPipes(new AppValidationPipe());

// Custom pipe example
@Injectable()
export class ParsePositiveIntPipe implements PipeTransform<string, number> {
    transform(value: string): number {
        const val = parseInt(value, 10);
        if (isNaN(val) || val <= 0) {
            throw new UsecaseBadRequestException(`${value} must be a positive integer`);
        }
        return val;
    }
}

// Usage in controller
@Get(':page')
async getPage(@Param('page', ParsePositiveIntPipe) page: number) { ... }
```

### 1.7 Guards

```typescript
// libs/src/common/guards/jwt-auth.guard.ts
import { Injectable, CanActivate, ExecutionContext, UnauthorizedException } from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { ClsService } from 'nestjs-cls';
import { IS_PUBLIC_KEY, USER_REQUEST } from '@common/constants';

@Injectable()
export class JwtAuthGuard implements CanActivate {
    constructor(
        private readonly reflector: Reflector,
        private readonly ssoService: SsoService,
        private readonly cls: ClsService,
    ) {}

    async canActivate(context: ExecutionContext): Promise<boolean> {
        const request = context.switchToHttp().getRequest();

        // Check x-api-internal header
        if (request.headers['x-api-internal'] === process.env.API_INTERNAL_SECRET) {
            return true;
        }

        // Check @Public() decorator
        const isPublic = this.reflector.getAllAndOverride<boolean>(IS_PUBLIC_KEY, [
            context.getHandler(),
            context.getClass(),
        ]);
        if (isPublic) return true;

        // Extract and verify token
        const token = this.extractBearerToken(request.headers.authorization);
        if (!token) throw new UsecaseUnauthorizedException('Token required');

        const user = await this.ssoService.verify(token);
        if (!user) throw new UsecaseUnauthorizedException('Invalid token');

        this.cls.set(USER_REQUEST, user);
        return true;
    }

    private extractBearerToken(authHeader?: string): string | null {
        if (!authHeader?.startsWith('Bearer ')) return null;
        return authHeader.slice(7);
    }
}

// Guard factory pattern
export const createAppGuard = (modulePath: string) => {
    @Injectable()
    class DynamicGuard extends JwtAuthGuard {
        canActivate(context: ExecutionContext): Promise<boolean> {
            const request = context.switchToHttp().getRequest();
            const url: string = request.url;
            const pattern = new RegExp(`/v\\d+/${modulePath}`);
            if (!pattern.test(url)) return Promise.resolve(true);
            return super.canActivate(context);
        }
    }
    return DynamicGuard;
};
```

### 1.8 Interceptors

```typescript
// libs/src/common/interceptors/idempotency.interceptor.ts
import { Injectable, NestInterceptor, ExecutionContext, CallHandler } from '@nestjs/common';
import { Observable, of } from 'rxjs';
import { tap } from 'rxjs/operators';
import { RedisService } from '@core/redis';

@Injectable()
export class IdempotencyInterceptor implements NestInterceptor {
    private readonly TTL = 86400; // 24 hours

    constructor(private readonly redis: RedisService) {}

    async intercept(context: ExecutionContext, next: CallHandler): Promise<Observable<unknown>> {
        const request = context.switchToHttp().getRequest();
        const key = request.headers['x-idempotency-key'];

        if (!key) return next.handle();

        const cached = await this.redis.get(`idempotency:${key}`);
        if (cached) return of(JSON.parse(cached));

        return next.handle().pipe(
            tap(async (response) => {
                await this.redis.set(`idempotency:${key}`, JSON.stringify(response), this.TTL);
            }),
        );
    }
}

// libs/src/common/transform/transform-sentry.interceptor.ts
@Injectable()
export class TransformUseSentryInterceptor implements NestInterceptor {
    intercept(context: ExecutionContext, next: CallHandler): Observable<unknown> {
        const request = context.switchToHttp().getRequest();
        const { method, url } = request;

        return RequestContextStorage.runWithSentry(
            { sessionId: uuidv7(), ip: request.ip },
            `${method} ${url}`,
            () => next.handle(),
        );
    }
}
```

### 1.9 Custom Decorators

```typescript
// libs/src/common/decorators/current-user.decorator.ts
import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { ClsService } from 'nestjs-cls';
import { USER_REQUEST } from '@common/constants';

export const CurrentUser = createParamDecorator(
    (data: keyof UserPayload | undefined, ctx: ExecutionContext) => {
        const cls = ctx.switchToHttp().getRequest().cls as ClsService;
        const user = cls.get<UserPayload>(USER_REQUEST);
        return data ? user?.[data] : user;
    },
);

// Public route decorator
export const IS_PUBLIC_KEY = 'isPublic';
export const Public = () => SetMetadata(IS_PUBLIC_KEY, true);

// Idempotency key decorator
export const IdempotencyKey = createParamDecorator(
    (_data: unknown, ctx: ExecutionContext): string => {
        const request = ctx.switchToHttp().getRequest();
        const key = request.headers['x-idempotency-key'];
        if (!key) throw new UsecaseBadRequestException('x-idempotency-key header required');
        return key as string;
    },
);

// Method logging decorator
export const LogExecution = (options?: LogExecutionOptions): MethodDecorator & ClassDecorator => {
    return (target: any, key?: string | symbol, descriptor?: TypedPropertyDescriptor<any>) => {
        if (descriptor) {
            // Method decorator
            const original = descriptor.value;
            descriptor.value = async function (...args: any[]) {
                const start = Date.now();
                const logger = new Logger(target.constructor.name);
                logger.log(`[START] ${String(key)}`);
                try {
                    const result = await original.apply(this, args);
                    logger.log(`[END] ${String(key)} duration=${Date.now() - start}ms`);
                    return result;
                } catch (err) {
                    logger.error(
                        `[ERROR] ${String(key)} duration=${Date.now() - start}ms`,
                        err.stack,
                    );
                    throw err;
                }
            };
            return descriptor;
        }
        // Class decorator — apply to all methods
        for (const methodName of Object.getOwnPropertyNames(target.prototype)) {
            if (methodName === 'constructor') continue;
            const desc = Object.getOwnPropertyDescriptor(target.prototype, methodName);
            if (desc?.value instanceof Function) {
                LogExecution(options)(target.prototype, methodName, desc);
                Object.defineProperty(target.prototype, methodName, desc);
            }
        }
        return target;
    };
};

// Retry decorator
export const Retry = (options: RetryOptions = {}): MethodDecorator => {
    const { retries = 3, factor = 2, minTimeout = 1000 } = options;
    return (_target, _key, descriptor: PropertyDescriptor) => {
        const original = descriptor.value;
        descriptor.value = async function (...args: any[]) {
            return asyncRetry(() => original.apply(this, args), {
                retries,
                factor,
                minTimeout,
                onRetry: (err, attempt) => {
                    Logger.warn(`Retry attempt ${attempt}: ${err.message}`, 'RetryDecorator');
                },
            });
        };
        return descriptor;
    };
};
```

---

## PHẦN 2 — FUNDAMENTALS

### 2.1 Custom Providers — useFactory + useExisting

```typescript
// Async factory with inject
@Module({
    providers: [
        {
            provide: KafkaEventBus,
            useFactory: async (kafkaService: KafkaServiceImpl, config: ConfigService) => {
                const bus = new KafkaEventBus(kafkaService, config);
                await bus.connect();
                return bus;
            },
            inject: [KafkaServiceImpl, ConfigService],
        },
        // Alias
        {
            provide: IEventBus,
            useExisting: KafkaEventBus,
        },
    ],
})
export class EventBusModule {}

// Value provider for testing
const mockOrderRepo: OrderRepositoryPort = {
    findById: jest.fn(),
    save: jest.fn(),
    delete: jest.fn(),
    deleteByIds: jest.fn(),
    findByCustomer: jest.fn(),
};

const module = await Test.createTestingModule({
    providers: [CreateOrderUseCase, { provide: OrderRepositoryPort, useValue: mockOrderRepo }],
}).compile();
```

### 2.2 Dynamic Modules — forRootAsync

```typescript
// libs/src/common/modules/report/report.module.ts
export interface ReportModuleOptions {
    redisStorage: RedisOptions;
}

export interface ReportModuleAsyncOptions {
    imports?: any[];
    useFactory: (...args: any[]) => Promise<ReportModuleOptions> | ReportModuleOptions;
    inject?: any[];
}

@Global()
@Module({})
export class LibReportModule {
    static forRootAsync(options: ReportModuleAsyncOptions): DynamicModule {
        return {
            module: LibReportModule,
            global: true,
            imports: options.imports || [],
            providers: [
                {
                    provide: REPORT_MODULE_OPTIONS,
                    useFactory: options.useFactory,
                    inject: options.inject || [],
                },
                ReportService,
                ReportStorageService,
            ],
            exports: [ReportService],
        };
    }
}

// Usage in AppModule:
LibReportModule.forRootAsync({
    useFactory: (configService: ConfigService) => ({
        redisStorage: {
            host: configService.get('REDIS_HOST'),
            port: configService.get('REDIS_PORT'),
            prefix: 'REPORT',
        },
    }),
    inject: [ConfigService],
}),
```

### 2.3 Execution Context + Reflector

```typescript
// Read metadata from guard
@Injectable()
export class RolesGuard implements CanActivate {
    constructor(private readonly reflector: Reflector) {}

    canActivate(context: ExecutionContext): boolean {
        const requiredRoles = this.reflector.getAllAndMerge<Role[]>(ROLES_KEY, [
            context.getHandler(),
            context.getClass(),
        ]);
        if (!requiredRoles?.length) return true;

        const request = context.switchToHttp().getRequest();
        const user = request.user as UserPayload;
        return requiredRoles.some((role) => user.roles?.includes(role));
    }
}

// WebSocket context switch
@Injectable()
export class WsAuthGuard implements CanActivate {
    canActivate(context: ExecutionContext): boolean {
        const client = context.switchToWs().getClient();
        const token = client.handshake?.auth?.token;
        return this.verifyToken(token);
    }
}
```

### 2.4 Lifecycle Hooks

```typescript
// libs/src/ddd/ddd.module.ts
@Global()
@Module({
    imports: [LibKafkaModule, EventBusModule],
    providers: [
        { provide: DomainEventPublisherService, useClass: DomainEventPublisherEventBusService },
    ],
    exports: [DomainEventPublisherService, EventBusModule],
})
export class LibDDDModule implements OnModuleInit {
    constructor(
        @Inject(DomainEventPublisherService)
        private readonly publisher: DomainEventPublisherService,
    ) {}

    onModuleInit(): void {
        DomainEventDispatcher.setDomainEventPublisherService(this.publisher);
    }
}

// Graceful shutdown
@Injectable()
export class KafkaConsumerService implements OnApplicationShutdown {
    constructor(private readonly kafka: KafkaServiceImpl) {}

    async onApplicationShutdown(signal?: string): Promise<void> {
        Logger.log(`Shutting down Kafka consumer (signal: ${signal})`, 'KafkaConsumer');
        await this.kafka.disconnect();
    }
}

// main.ts
app.enableShutdownHooks(); // Required for shutdown events
```

### 2.5 Module Reference — Dynamic Provider Resolution

```typescript
// Dynamic use case dispatch
@Injectable()
export class UseCaseDispatcher {
    constructor(private readonly moduleRef: ModuleRef) {}

    async dispatch<T>(useCaseName: string, request: unknown): Promise<T> {
        const useCase = await this.moduleRef.resolve<BaseCommand<any, T>>(useCaseName, undefined, {
            strict: false,
        });
        return useCase.executeWithHooks(request);
    }
}

// Saga manager registers dynamically
@Injectable()
export class SagaManager implements OnModuleInit {
    private readonly definitions = new Map<string, SagaDefinition<any>>();

    constructor(
        private readonly discovery: DiscoveryService,
        private readonly reflector: Reflector,
    ) {}

    onModuleInit(): void {
        const wrappers = this.discovery.getProviders();
        for (const wrapper of wrappers) {
            const instance = wrapper.instance;
            if (instance instanceof SagaDefinition) {
                this.definitions.set(instance.sagaType, instance);
            }
        }
    }
}
```

### 2.6 Testing

```typescript
// Unit test — use case
describe('CreateOrderUseCase', () => {
    let useCase: CreateOrderUseCase;
    let orderRepo: jest.Mocked<OrderRepositoryPort>;

    beforeEach(async () => {
        const mockRepo = {
            findById: jest.fn(),
            save: jest.fn().mockResolvedValue({ id: 'order-uuid' }),
        };

        const module = await Test.createTestingModule({
            providers: [
                CreateOrderUseCase,
                { provide: OrderRepositoryPort, useValue: mockRepo },
                { provide: PricingDomainService, useValue: { calculateDiscount: jest.fn() } },
            ],
        }).compile();

        useCase = module.get(CreateOrderUseCase);
        orderRepo = module.get(OrderRepositoryPort);
    });

    it('should create order successfully', async () => {
        const result = await useCase.executeWithHooks({
            customerId: 'customer-uuid',
            items: [{ productId: 'prod-1', quantity: 2, price: 100000 }],
            idempotencyKey: 'key-123',
        });

        expect(result.orderId).toBeDefined();
        expect(orderRepo.save).toHaveBeenCalledTimes(1);
    });
});

// Integration test — module
describe('OrderModule (integration)', () => {
    let app: INestApplication;

    beforeAll(async () => {
        const module = await Test.createTestingModule({
            imports: [
                TypeOrmModule.forRoot({
                    type: 'sqlite',
                    database: ':memory:',
                    entities: [OrderOrmEntity],
                    synchronize: true,
                }),
                OrderApplicationModule,
            ],
        }).compile();

        app = module.createNestApplication();
        await app.init();
    });

    it('should persist and retrieve order', async () => {
        const useCase = app.get(CreateOrderUseCase);
        const result = await useCase.executeWithHooks({ customerId: 'test', items: [] });
        expect(result.orderId).toBeDefined();
    });
});

// E2E test
describe('POST /v1/orders', () => {
    it('should return 201 with valid request', () => {
        return request(app.getHttpServer())
            .post('/v1/orders')
            .set('Authorization', `Bearer ${validToken}`)
            .set('x-idempotency-key', 'unique-key-123')
            .send({ customerId: 'uuid', items: [{ productId: 'uuid', quantity: 1, price: 50000 }] })
            .expect(201)
            .expect((res) => {
                expect(res.body.orderId).toBeDefined();
            });
    });
});
```

---

## PHẦN 3 — TECHNIQUES

### 3.1 Configuration

```typescript
// src/app.module.ts
ConfigModule.forRoot({
    isGlobal: true,
    envFilePath: getEnvFilePath(env.NODE_ENV),
    expandVariables: true,
    validationSchema: joi.object({
        PORT: joi.number().required(),
        DB_HOST: joi.string().required(),
        DB_PORT: joi.number().required(),
        DB_NAME: joi.string().required(),
        DB_USER: joi.string().required(),
        DB_PASSWORD: joi.string().required(),
        REDIS_HOST: joi.string().required(),
        REDIS_PORT: joi.number().required(),
        KAFKA_DEFAULT_BROKER: joi.string().required(),
        JWT_SECRET: joi.string().min(32).required(),
        NODE_ENV: joi.string().valid('development', 'staging', 'production').required(),
    }),
});

// Typed config service usage
@Injectable()
export class AppConfigService {
    constructor(private readonly config: ConfigService) {}

    get port(): number {
        return this.config.get<number>('PORT');
    }
    get redisHost(): string {
        return this.config.get<string>('REDIS_HOST');
    }
    get jwtSecret(): string {
        return this.config.get<string>('JWT_SECRET');
    }
}

// Namespace config
export default registerAs('database', () => ({
    host: process.env.DB_HOST,
    port: parseInt(process.env.DB_PORT, 10),
    name: process.env.DB_NAME,
}));

// Usage: configService.get('database.host')
```

### 3.2 Database — TypeORM Patterns

```typescript
// ORM Entity
@Entity('orders')
@Index(['customerId', 'status'])
@Index(['createdAt'])
export class OrderOrmEntity {
    @PrimaryColumn('uuid')
    id: string;

    @Column('uuid')
    customerId: string;

    @Column({ type: 'enum', enum: ENUM_ORDER_STATUS })
    status: ENUM_ORDER_STATUS;

    @Column({ type: 'decimal', precision: 18, scale: 2 })
    totalAmount: number;

    @Column({ length: 3 })
    currency: string;

    @OneToMany(() => OrderItemOrmEntity, (item) => item.order, { cascade: true })
    items: OrderItemOrmEntity[];

    @CreateDateColumn()
    createdAt: Date;

    @UpdateDateColumn()
    updatedAt: Date;

    @VersionColumn()
    version: number;

    @Index({ unique: true })
    @Column({ nullable: true })
    idempotencyKey: string;
}

// Repository implementation
@Injectable()
export class OrderRepository
    extends BaseRepositoryTypeORM<Order, OrderOrmEntity, OrderMapper>
    implements OrderRepositoryPort
{
    constructor(@InjectDataSource() dataSource: DataSource, mapper: OrderMapper) {
        super(OrderOrmEntity, dataSource, mapper);
    }

    async findByCustomer(customerId: string): Promise<Order[]> {
        const orms = await this.repository
            .createQueryBuilder('order')
            .select(['order.id', 'order.status', 'order.totalAmount', 'order.createdAt'])
            .where('order.customerId = :customerId', { customerId })
            .orderBy('order.createdAt', 'DESC')
            .getMany();
        return orms.map((orm) => this.mapper.toDomain(orm));
    }

    async findByStatusWithItems(status: ENUM_ORDER_STATUS): Promise<Order[]> {
        const orms = await this.repository
            .createQueryBuilder('order')
            .leftJoinAndSelect('order.items', 'item')
            .where('order.status = :status', { status })
            .getMany();
        return orms.map((orm) => this.mapper.toDomain(orm));
    }

    async deleteByIds(ids: string[]): Promise<void> {
        await this.repository.delete({ id: In(ids) });
    }
}

// Transaction pattern
@Injectable()
export class OrderTransactionService {
    constructor(@InjectDataSource() private readonly dataSource: DataSource) {}

    async createOrderWithItems(order: Order, items: OrderItem[]): Promise<Order> {
        return this.dataSource.transaction(async (manager) => {
            const orderRepo = manager.getRepository(OrderOrmEntity);
            const itemRepo = manager.getRepository(OrderItemOrmEntity);

            const savedOrder = await orderRepo.save(this.orderMapper.toOrm(order));
            await itemRepo.save(items.map((i) => this.itemMapper.toOrm(i)));

            return this.orderMapper.toDomain(savedOrder);
        });
    }
}
```

### 3.3 Validation — DTO + class-validator

```typescript
// Full DTO example
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import {
    IsArray,
    IsEnum,
    IsInt,
    IsNotEmpty,
    IsOptional,
    IsPositive,
    IsString,
    IsUUID,
    MaxLength,
    Min,
    MinLength,
    ValidateNested,
} from 'class-validator';
import { Type, Transform } from 'class-transformer';

export class CreateOrderItemDto {
    @ApiProperty({ description: 'Product UUID', example: '550e8400-e29b-41d4-a716-446655440000' })
    @IsUUID()
    @IsNotEmpty()
    productId: string;

    @ApiProperty({ description: 'Quantity', minimum: 1, example: 2 })
    @IsInt()
    @Min(1)
    quantity: number;

    @ApiPropertyOptional({ description: 'Custom note', maxLength: 200 })
    @IsString()
    @MaxLength(200)
    @IsOptional()
    note?: string;
}

export class CreateOrderDto {
    @ApiProperty({ example: '550e8400-e29b-41d4-a716-446655440000' })
    @IsUUID()
    @IsNotEmpty()
    customerId: string;

    @ApiProperty({ type: [CreateOrderItemDto] })
    @IsArray()
    @ValidateNested({ each: true })
    @Type(() => CreateOrderItemDto)
    @ArrayMinSize(1)
    items: CreateOrderItemDto[];

    @ApiProperty({ enum: OrderChannel, enumName: 'OrderChannel' })
    @IsEnum(OrderChannel)
    channel: OrderChannel;

    @ApiPropertyOptional()
    @IsString()
    @MaxLength(500)
    @IsOptional()
    note?: string;
}

// Query DTO with pagination
export class ListOrdersQueryDto {
    @ApiPropertyOptional({ default: 1 })
    @IsInt()
    @Min(1)
    @IsOptional()
    @Type(() => Number)
    page?: number = 1;

    @ApiPropertyOptional({ default: 20 })
    @IsInt()
    @Min(1)
    @IsOptional()
    @Type(() => Number)
    limit?: number = 20;

    @ApiPropertyOptional({ enum: ENUM_ORDER_STATUS })
    @IsEnum(ENUM_ORDER_STATUS)
    @IsOptional()
    status?: ENUM_ORDER_STATUS;

    @ApiPropertyOptional()
    @IsUUID()
    @IsOptional()
    customerId?: string;
}
```

### 3.4 Caching — Redis Cache Decorator

```typescript
// libs/src/core/redis/decorator/redis-cache.decorator.ts
import { RedisCacheMethodDecorator } from '@core/redis';

@Injectable()
export class ProductQueryService {
    constructor(private readonly repo: ProductRepositoryPort) {}

    @RedisCacheMethodDecorator({
        mode: 'GET',
        prefix: 'product:detail',
        ttl: 3600, // 1 hour
        parameters: [0], // Use first parameter (productId) in cache key
    })
    async getProduct(productId: string): Promise<ProductDetailResponse> {
        return this.repo.findById(productId);
    }

    @RedisCacheMethodDecorator({
        mode: 'SET',
        prefix: 'product',
        relatedPatterns: ['product:detail:*', 'product:list:*'],
    })
    async updateProduct(productId: string, data: UpdateProductDto): Promise<Product> {
        return this.repo.save(productId, data);
    }

    @RedisCacheMethodDecorator({
        mode: 'GET',
        prefix: 'product:list',
        ttl: 300, // 5 minutes
        keyGenerator: (args) => `page:${args[0].page}:limit:${args[0].limit}`,
    })
    async listProducts(query: ListProductsQuery): Promise<PaginatedResponse<Product>> {
        return this.repo.findMany(query);
    }
}

// Manual Redis usage
@Injectable()
export class SessionService {
    constructor(@InjectRedis() private readonly redis: Redis) {}

    async setSession(sessionId: string, data: SessionData): Promise<void> {
        await this.redis.setEx(`session:${sessionId}`, 3600, JSON.stringify(data));
    }

    async getSession(sessionId: string): Promise<SessionData | null> {
        const raw = await this.redis.get(`session:${sessionId}`);
        return raw ? JSON.parse(raw) : null;
    }

    async invalidateSession(sessionId: string): Promise<void> {
        await this.redis.del(`session:${sessionId}`);
    }
}
```

### 3.5 Serialization — Response Mapping

```typescript
// Response DTO
export class OrderDetailResponse {
    @ApiProperty()
    @Expose()
    orderId: string;

    @ApiProperty()
    @Expose()
    status: ENUM_ORDER_STATUS;

    @ApiProperty({ type: [OrderItemResponse] })
    @Expose()
    @Type(() => OrderItemResponse)
    items: OrderItemResponse[];

    @ApiProperty()
    @Expose()
    @Transform(({ value }) => value?.toISOString())
    createdAt: string;
}

// Mapper toDto
export class OrderMapper extends BaseMapper<Order, OrderOrmEntity, OrderDetailResponse> {
    toDomain(orm: OrderOrmEntity): Order {
        /* ... */
    }
    toOrm(domain: Order): OrderOrmEntity {
        /* ... */
    }

    toDto(domain: Order): OrderDetailResponse {
        return plainToInstance(
            OrderDetailResponse,
            {
                orderId: domain.id.toValue(),
                status: domain.props.status,
                items: domain.props.items.map((item) => this.itemMapper.toDto(item)),
                createdAt: domain.props.createdAt,
            },
            { excludeExtraneousValues: true },
        );
    }
}

// Paginated response
export class ApiResponseData<TData> {
    success: boolean;
    statusCode: number;
    data: TData;
    metadata?: PaginationMetadata;
    message?: string;

    static createResponse<T>(data: T, context: ApiResponseContext): ApiResponseData<T> {
        return {
            success: true,
            statusCode: context.statusCode || 200,
            data,
            metadata: context.metadata,
        };
    }
}
```

### 3.6 Queues — BullMQ

```typescript
// Queue module setup
@Module({
    imports: [
        BullModule.forRootAsync({
            useFactory: (config: ConfigService) => ({
                connection: {
                    host: config.get('REDIS_HOST'),
                    port: config.get('REDIS_PORT'),
                },
            }),
            inject: [ConfigService],
        }),
        BullModule.registerQueue({ name: 'order-processing' }),
        BullModule.registerQueue({ name: 'notification' }),
    ],
    providers: [OrderProcessingQueue, OrderProcessingWorker],
    exports: [OrderProcessingQueue],
})
export class QueueModule {}

// Queue producer
@Injectable()
export class OrderProcessingQueue {
    constructor(@InjectQueue('order-processing') private readonly queue: Queue) {}

    async enqueueConfirmation(orderId: string, delay = 0): Promise<void> {
        await this.queue.add(
            'confirm-order',
            { orderId },
            {
                delay,
                attempts: 3,
                backoff: { type: 'exponential', delay: 2000 },
                removeOnComplete: 100,
                removeOnFail: 500,
            },
        );
    }

    async enqueueShipment(orderId: string, scheduledAt: Date): Promise<void> {
        const delay = scheduledAt.getTime() - Date.now();
        await this.queue.add(
            'ship-order',
            { orderId },
            {
                delay: Math.max(0, delay),
                attempts: 5,
                backoff: { type: 'exponential', delay: 5000 },
                priority: 1,
            },
        );
    }
}

// Queue worker
@Processor({ name: 'order-processing', concurrency: 5 })
export class OrderProcessingWorker {
    private readonly logger = new Logger(OrderProcessingWorker.name);

    constructor(
        private readonly confirmOrder: ConfirmOrderUseCase,
        private readonly shipOrder: ShipOrderUseCase,
    ) {}

    @Process('confirm-order')
    async handleConfirmation(job: Job<{ orderId: string }>): Promise<void> {
        const { orderId } = job.data;
        this.logger.log(`Confirming order ${orderId} (attempt ${job.attemptsMade + 1})`);
        await this.confirmOrder.executeWithHooks({ orderId });
    }

    @OnQueueFailed()
    onFailed(job: Job, error: Error): void {
        this.logger.error(`Job ${job.id} failed: ${error.message}`, error.stack);
    }

    @OnQueueCompleted()
    onCompleted(job: Job): void {
        this.logger.log(`Job ${job.id} completed successfully`);
    }
}
```

### 3.7 Task Scheduling — Cron

```typescript
@Injectable()
export class OrderCleanupScheduler {
    private readonly logger = new Logger(OrderCleanupScheduler.name);

    constructor(private readonly cleanupOrders: CleanupExpiredOrdersUseCase) {}

    @Cron(CronExpression.EVERY_DAY_AT_MIDNIGHT)
    async cleanupExpiredOrders(): Promise<void> {
        this.logger.log('Running expired orders cleanup...');
        const result = await this.cleanupOrders.executeWithHooks({
            olderThanDays: 30,
        });
        this.logger.log(`Cleaned up ${result.count} expired orders`);
    }

    @Interval(60000) // Every minute
    async checkPendingPayments(): Promise<void> {
        await this.syncPaymentStatus.executeWithHooks({});
    }

    @Timeout(5000) // Run once after 5 seconds
    async warmUpCache(): Promise<void> {
        await this.cacheWarmUp.executeWithHooks({});
    }
}

// Register in module
@Module({
    imports: [ScheduleModule.forRoot()],
    providers: [OrderCleanupScheduler],
})
export class SchedulingModule {}
```

### 3.8 Logging — Winston + Structured

```typescript
// libs/src/core/logger/winston/winston.factory.ts
export class WinstonLoggerFactory {
    static create(options: WinstonLoggerOptions): WinstonLogger {
        return WinstonModule.createLogger({
            level: options.logLevels || 'info',
            format: winston.format.combine(
                winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss.SSS' }),
                winston.format.ms(),
                winston.format.json(),
            ),
            transports: [
                new winston.transports.Console({
                    format: winston.format.combine(
                        winston.format.colorize({ all: true }),
                        winston.format.printf(
                            ({ timestamp, level, message, context, ms }) =>
                                `[${timestamp}] ${level} [${context}] ${message} ${ms}`,
                        ),
                    ),
                }),
                ...(options.enableSentry && options.sentryTransportOptions?.dsn
                    ? [new SentryTransport(options.sentryTransportOptions)]
                    : []),
            ],
        });
    }
}

// Usage in service
@Injectable()
export class CreateOrderUseCase extends BaseCommand<CreateOrderRequest, CreateOrderResponse> {
    private readonly logger = new Logger(CreateOrderUseCase.name);

    async execute(request: CreateOrderRequest): Promise<CreateOrderResponse> {
        this.logger.log(`Creating order for customer ${request.customerId}`);

        // Business logic...

        this.logger.log({
            message: 'Order created',
            orderId: result.id,
            customerId: request.customerId,
        });
        return result;
    }
}
```

### 3.9 File Upload + Streaming

```typescript
// File upload controller
@Controller('files')
@UseGuards(JwtAuthGuard)
export class FileUploadController {
    constructor(private readonly uploadFile: UploadFileUseCase) {}

    @Post('upload')
    @UseInterceptors(FileInterceptor('file'))
    async upload(
        @UploadedFile(new ParseFilePipe({
            validators: [
                new MaxFileSizeValidator({ maxSize: 50 * 1024 * 1024 }), // 50MB
                new FileTypeValidator({ fileType: /(jpeg|png|pdf|xlsx)$/ }),
            ],
        })) file: Express.Multer.File,
    ): Promise<UploadFileResponse> {
        return this.uploadFile.executeWithHooks({
            buffer: file.buffer,
            originalName: file.originalname,
            mimeType: file.mimetype,
            size: file.size,
        });
    }

    @Post('upload-multiple')
    @UseInterceptors(FilesInterceptor('files', 10))
    async uploadMultiple(
        @UploadedFiles() files: Express.Multer.File[],
    ): Promise<UploadMultipleResponse> {
        return this.uploadFiles.executeWithHooks({ files });
    }
}

// Streaming file response
@Get('export/orders')
async exportOrders(@Query() query: ExportOrdersQuery, @Res() res: Response): Promise<void> {
    res.setHeader('Content-Type', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
    res.setHeader('Content-Disposition', 'attachment; filename="orders.xlsx"');

    const stream = await this.exportOrdersUseCase.executeWithHooks(query);
    stream.pipe(res);
}

// StreamableFile
@Get('report/:id')
async getReport(@Param('id') id: string): Promise<StreamableFile> {
    const buffer = await this.reportService.generatePdf(id);
    return new StreamableFile(buffer, {
        type: 'application/pdf',
        disposition: `attachment; filename="report-${id}.pdf"`,
    });
}
```

### 3.10 HTTP Module — External API

```typescript
// HTTP client với Circuit Breaker
@Injectable()
export class PaymentHttpClient {
    private readonly circuitBreaker: CircuitBreaker;
    private readonly logger = new Logger(PaymentHttpClient.name);

    constructor(
        private readonly httpService: HttpService,
        private readonly config: ConfigService,
    ) {
        this.circuitBreaker = new CircuitBreaker(this.doInitPayment.bind(this), {
            timeout: 5000,
            errorThresholdPercentage: 50,
            resetTimeout: 30000,
            name: 'PaymentService',
        });

        this.circuitBreaker.on('open', () => this.logger.warn('Payment circuit breaker OPEN'));
        this.circuitBreaker.on('close', () => this.logger.log('Payment circuit breaker CLOSED'));
    }

    @Retry({ retries: 3, factor: 2, minTimeout: 1000 })
    async initPayment(request: InitPaymentRequest): Promise<InitPaymentResponse> {
        return this.circuitBreaker.fire(request);
    }

    private async doInitPayment(request: InitPaymentRequest): Promise<InitPaymentResponse> {
        const response = await firstValueFrom(
            this.httpService
                .post<InitPaymentResponse>(
                    `${this.config.get('PAYMENT_SERVICE_URL')}/payments`,
                    request,
                )
                .pipe(
                    timeout(5000),
                    catchError((err: AxiosError) => {
                        this.logger.error(`Payment service error: ${err.message}`);
                        throw new ThirdPartyException(
                            `Payment service unavailable: ${err.message}`,
                        );
                    }),
                ),
        );
        return response.data;
    }
}
```

### 3.11 Server-Sent Events

```typescript
// SSE Gateway
@Controller('events')
@UseGuards(JwtAuthGuard)
export class EventsController {
    constructor(private readonly sseService: SseService) {}

    @Sse('progress/:jobId')
    streamJobProgress(@Param('jobId') jobId: string): Observable<MessageEvent> {
        return new Observable((subscriber) => {
            const subscription = this.sseService.subscribe(jobId, (event) => {
                subscriber.next({
                    data: JSON.stringify(event),
                    type: event.type,
                    id: event.id,
                });

                if (event.type === 'completed' || event.type === 'failed') {
                    subscriber.complete();
                }
            });

            return () => subscription.unsubscribe();
        });
    }

    @Sse('notifications')
    streamNotifications(@CurrentUser() user: UserPayload): Observable<MessageEvent> {
        return interval(30000).pipe(
            map(() => ({ data: JSON.stringify({ type: 'heartbeat' }) })),
            mergeWith(this.sseService.getUserEvents(user.id)),
        );
    }
}
```

---

## PHẦN 4 — SECURITY

### 4.1 Authentication — JWT + SSO

```typescript
// SSO verification service
@Injectable()
export class SsoService {
    constructor(
        private readonly httpService: HttpService,
        private readonly config: ConfigService,
        @InjectRedis() private readonly redis: Redis,
    ) {}

    async verify(token: string): Promise<UserPayload | null> {
        const cacheKey = `auth:token:${this.hashToken(token)}`;
        const cached = await this.redis.get(cacheKey);
        if (cached) return JSON.parse(cached);

        try {
            const response = await firstValueFrom(
                this.httpService.post<UserPayload>(`${this.config.get('SSO_URL')}/verify`, {
                    token,
                }),
            );
            await this.redis.setEx(cacheKey, 300, JSON.stringify(response.data));
            return response.data;
        } catch {
            return null;
        }
    }

    private hashToken(token: string): string {
        return createHash('sha256').update(token).digest('hex');
    }
}

// Auth guard full implementation
@Injectable()
export class JwtAuthGuard implements CanActivate {
    constructor(
        private readonly reflector: Reflector,
        private readonly ssoService: SsoService,
        private readonly cls: ClsService,
    ) {}

    async canActivate(context: ExecutionContext): Promise<boolean> {
        const request = context.switchToHttp().getRequest<Request>();

        if (this.isInternalRequest(request)) return true;
        if (this.isPublicEndpoint(context)) return true;

        const token = this.extractBearerToken(request);
        if (!token) throw new UsecaseUnauthorizedException('Authorization token required');

        const user = await this.ssoService.verify(token);
        if (!user) throw new UsecaseUnauthorizedException('Invalid or expired token');

        this.cls.set('user', user);
        return true;
    }

    private isInternalRequest(req: Request): boolean {
        return req.headers['x-api-internal'] === process.env.API_INTERNAL_SECRET;
    }

    private isPublicEndpoint(ctx: ExecutionContext): boolean {
        return (
            this.reflector.getAllAndOverride<boolean>('isPublic', [
                ctx.getHandler(),
                ctx.getClass(),
            ]) ?? false
        );
    }

    private extractBearerToken(req: Request): string | null {
        const auth = req.headers.authorization;
        return auth?.startsWith('Bearer ') ? auth.slice(7) : null;
    }
}
```

### 4.2 Authorization — RBAC

```typescript
// Role decorator
export const Roles = (...roles: Role[]) => SetMetadata('roles', roles);

// Permission decorator
export const Permissions = (...perms: Permission[]) => SetMetadata('permissions', perms);

// RBAC guard
@Injectable()
export class RolesGuard implements CanActivate {
    constructor(
        private readonly reflector: Reflector,
        private readonly cls: ClsService,
    ) {}

    canActivate(context: ExecutionContext): boolean {
        const requiredRoles = this.reflector.getAllAndMerge<Role[]>('roles', [
            context.getHandler(),
            context.getClass(),
        ]);
        if (!requiredRoles?.length) return true;

        const user = this.cls.get<UserPayload>('user');
        return requiredRoles.some((role) => user?.roles?.includes(role));
    }
}

// Usage in controller
@Controller('admin')
@UseGuards(JwtAuthGuard, RolesGuard)
@Roles(Role.Admin)
export class AdminController {
    @Get('users')
    listUsers() {
        /* ... */
    }

    @Delete('users/:id')
    @Roles(Role.SuperAdmin) // Override class-level
    deleteUser(@Param('id') id: string) {
        /* ... */
    }
}
```

### 4.3 Encryption

```typescript
// Encryption service
@Injectable()
export class EncryptionService {
    private readonly algorithm = 'aes-256-gcm';
    private readonly key: Buffer;

    constructor(private readonly config: ConfigService) {
        const secret = this.config.get<string>('ENCRYPTION_KEY');
        this.key = createHash('sha256').update(secret).digest();
    }

    encrypt(plaintext: string): EncryptedData {
        const iv = randomBytes(16);
        const cipher = createCipheriv(this.algorithm, this.key, iv);
        const encrypted = Buffer.concat([cipher.update(plaintext, 'utf8'), cipher.final()]);
        const tag = cipher.getAuthTag();

        return {
            encrypted: encrypted.toString('base64'),
            iv: iv.toString('base64'),
            tag: tag.toString('base64'),
        };
    }

    decrypt(data: EncryptedData): string {
        const decipher = createDecipheriv(this.algorithm, this.key, Buffer.from(data.iv, 'base64'));
        decipher.setAuthTag(Buffer.from(data.tag, 'base64'));
        return decipher.update(data.encrypted, 'base64', 'utf8') + decipher.final('utf8');
    }

    async hashPassword(password: string): Promise<string> {
        return bcrypt.hash(password, 12);
    }

    async verifyPassword(password: string, hash: string): Promise<boolean> {
        return bcrypt.compare(password, hash);
    }
}
```

### 4.4 Rate Limiting

```typescript
// Throttle setup
@Module({
    imports: [
        ThrottlerModule.forRootAsync({
            useFactory: (config: ConfigService, redis: Redis) => ({
                throttlers: [
                    { name: 'short', ttl: 60000, limit: 20 },
                    { name: 'long', ttl: 3600000, limit: 1000 },
                ],
                storage: new ThrottlerStorageRedisService(redis),
            }),
            inject: [ConfigService, getRedisToken()],
        }),
    ],
})
export class AppModule {}

// Custom throttle per endpoint
@Controller('auth')
export class AuthController {
    @Post('login')
    @Throttle({ short: { ttl: 60000, limit: 5 } }) // 5 per minute
    async login(@Body() dto: LoginDto) {
        /* ... */
    }

    @Post('register')
    @SkipThrottle() // Public, no throttle
    async register(@Body() dto: RegisterDto) {
        /* ... */
    }
}

// Custom throttle guard with custom key
@Injectable()
export class UserThrottleGuard extends ThrottlerGuard {
    protected async getTracker(req: Record<string, any>): Promise<string> {
        // Use user ID instead of IP for authenticated routes
        const user = req.user as UserPayload;
        return user ? `user:${user.id}` : req.ip;
    }
}
```

---

## PHẦN 5 — WEBSOCKETS

### 5.1 Gateway — Full Implementation

```typescript
// src/presentation/websocket/notification.gateway.ts
@WebSocketGateway({
    cors: {
        origin: process.env.CORS_ORIGINS?.split(',') || '*',
        credentials: true,
    },
    namespace: '/notifications',
})
@UseGuards(WsAuthGuard)
export class NotificationGateway
    implements OnGatewayInit, OnGatewayConnection, OnGatewayDisconnect
{
    @WebSocketServer()
    server: Server;

    private readonly logger = new Logger(NotificationGateway.name);

    constructor(
        private readonly notificationService: NotificationService,
        private readonly cls: ClsService,
    ) {}

    afterInit(server: Server): void {
        this.logger.log('WebSocket Gateway initialized');
    }

    async handleConnection(client: Socket): Promise<void> {
        const userId = client.handshake.auth?.userId;
        if (!userId) {
            client.disconnect();
            return;
        }

        await client.join(`user:${userId}`);
        this.logger.log(`Client connected: ${client.id} (user: ${userId})`);
    }

    handleDisconnect(client: Socket): void {
        this.logger.log(`Client disconnected: ${client.id}`);
    }

    @SubscribeMessage('join-room')
    async handleJoinRoom(
        @ConnectedSocket() client: Socket,
        @MessageBody() data: { roomId: string },
    ): Promise<WsResponse<{ joined: boolean }>> {
        await client.join(`room:${data.roomId}`);
        return { event: 'room-joined', data: { joined: true } };
    }

    @SubscribeMessage('subscribe-job')
    handleSubscribeJob(
        @ConnectedSocket() client: Socket,
        @MessageBody() data: { jobId: string },
    ): void {
        client.join(`job:${data.jobId}`);
    }

    // Emit to specific user
    emitToUser(userId: string, event: string, data: unknown): void {
        this.server.to(`user:${userId}`).emit(event, data);
    }

    // Emit to room
    emitToRoom(roomId: string, event: string, data: unknown): void {
        this.server.to(`room:${roomId}`).emit(event, data);
    }
}

// WebSocket Auth Guard
@Injectable()
export class WsAuthGuard implements CanActivate {
    constructor(private readonly ssoService: SsoService) {}

    async canActivate(context: ExecutionContext): Promise<boolean> {
        const client = context.switchToWs().getClient<Socket>();
        const token =
            client.handshake?.auth?.token || client.handshake?.headers?.authorization?.slice(7);
        if (!token) return false;
        const user = await this.ssoService.verify(token);
        if (!user) return false;
        client.data.user = user;
        return true;
    }
}

// Redis adapter for scaling
export class RedisIoAdapter extends IoAdapter {
    private adapterConstructor: ReturnType<typeof createAdapter>;

    async connectToRedis(options: RedisOptions): Promise<void> {
        const pubClient = new Redis(options);
        const subClient = pubClient.duplicate();
        this.adapterConstructor = createAdapter(pubClient, subClient);
    }

    createIOServer(port: number, options?: ServerOptions): Server {
        const server = super.createIOServer(port, options) as Server;
        server.adapter(this.adapterConstructor);
        return server;
    }
}
```

---

## PHẦN 6 — MICROSERVICES

### 6.1 Hybrid App + Kafka Consumer

```typescript
// main.ts — hybrid app
async function bootstrap() {
    const app = await NestFactory.create<NestExpressApplication>(AppModule, {
        logger: WinstonLoggerFactory.create({ appName: 'DataVisualizer' }),
    });

    // HTTP setup
    app.useGlobalPipes(new AppValidationPipe());
    app.enableVersioning({ defaultVersion: '1', type: VersioningType.URI });

    // Start both HTTP and microservices
    await Promise.race([
        app.startAllMicroservices(),
        app.listen(configService.get('PORT'), () => {
            Logger.log(`HTTP server running on ${app.getUrl()}`);
        }),
    ]);
}

// Kafka consumer registration
@Module({
    imports: [
        ClientsModule.registerAsync([
            {
                name: 'KAFKA_CLIENT',
                useFactory: (config: ConfigService) => ({
                    transport: Transport.KAFKA,
                    options: {
                        client: {
                            clientId: config.get('KAFKA_CLIENT_ID'),
                            brokers: [config.get('KAFKA_BROKER')],
                        },
                        consumer: { groupId: config.get('KAFKA_CONSUMER_GROUP') },
                    },
                }),
                inject: [ConfigService],
            },
        ]),
    ],
})
export class KafkaModule {}
```

### 6.2 Kafka — Event Publishing + Consuming

```typescript
// Domain event publisher
@Injectable()
export class KafkaEventBus implements IEventBus {
    private readonly topicMap: Map<string, string> = new Map([
        ['OrderCreatedEvent', 'EventBusKafka.ORDER_CREATED'],
        ['OrderConfirmedEvent', 'EventBusKafka.ORDER_CONFIRMED'],
        ['PaymentCompletedEvent', 'EventBusKafka.PAYMENT_COMPLETED'],
    ]);

    constructor(private readonly kafkaService: KafkaServiceImpl) {}

    async publish<T extends BaseDomainEvents<any>>(event: T): Promise<void> {
        const topic = this.topicMap.get(event.eventName) || 'EventBusKafka.DOMAIN_EVENTS';
        const envelope: DomainEventEnvelope = {
            eventType: event.eventName,
            eventId: event.eventId,
            aggregateId: event.aggregateId,
            occurredAt: event.occurredAt.toISOString(),
            props: event.props,
            metadata: {},
        };

        await this.kafkaService.emit(
            { topic, action: event.eventName, clientName: 'primary' },
            { key: event.aggregateId, value: JSON.stringify(envelope) },
        );
    }
}

// Domain event consumer
@Controller()
export class OrderEventConsumer {
    constructor(private readonly handleOrderCreated: HandleOrderCreatedUseCase) {}

    @SubscribeEventPattern('EventBusKafka.ORDER_CREATED')
    async onOrderCreated(
        @Payload() envelope: DomainEventEnvelope,
        @Ctx() ctx: KafkaContext,
    ): Promise<void> {
        const heartbeat = ctx.getHeartbeat();
        await heartbeat(); // Prevent timeout for long processing

        await this.handleOrderCreated.executeWithHooks({
            orderId: envelope.aggregateId,
            ...envelope.props,
        });
    }
}
```

### 6.3 RabbitMQ — Publisher + Consumer

```typescript
// RabbitMQ publisher
@Injectable()
export class OrderEventPublisher {
    constructor(private readonly rabbitmq: RabbitMQService) {}

    async publishOrderCreated(event: OrderCreatedDomainEvent): Promise<void> {
        await this.rabbitmq.publish('order.exchange', 'order.created', {
            orderId: event.aggregateId,
            customerId: event.props.customerId,
            totalAmount: event.props.totalAmount,
            occurredAt: event.occurredAt.toISOString(),
        });
    }
}

// RabbitMQ consumer
@Controller()
export class NotificationConsumer {
    constructor(private readonly sendNotification: SendOrderNotificationUseCase) {}

    @MessagePattern('order.created')
    async handleOrderCreated(@Payload() data: OrderCreatedMessage): Promise<void> {
        await this.sendNotification.executeWithHooks({
            orderId: data.orderId,
            customerId: data.customerId,
            notificationType: 'ORDER_CONFIRMATION',
        });
    }

    @EventPattern('order.*')
    async handleOrderEvents(@Payload() data: unknown, @Ctx() ctx: RmqContext): Promise<void> {
        const channel = ctx.getChannelRef();
        const message = ctx.getMessage();

        try {
            await this.processEvent(data);
            channel.ack(message);
        } catch (error) {
            channel.nack(message, false, false); // Dead letter queue
        }
    }
}
```

### 6.4 Saga Pattern — Full Example

```typescript
// Saga definition
@Injectable()
export class OrderPlacementSaga extends SagaDefinition<OrderPlacementSagaData> {
    readonly sagaType = 'OrderPlacementSaga';

    steps(): ISagaStep<OrderPlacementSagaData>[] {
        return [
            SagaStep.create<OrderPlacementSagaData>('ReservePayment')
                .invoke(async (data) => ({
                    destination: 'payment',
                    commandType: 'InitPaymentCommand',
                    payload: { orderId: data.orderId, amount: data.totalAmount },
                }))
                .handleReply((data, reply) => ({
                    ...data,
                    paymentReservationId: reply.data?.paymentReservationId as string,
                }))
                .compensate(async (data) => {
                    if (!data.paymentReservationId) return null;
                    return {
                        destination: 'payment',
                        commandType: 'CancelPaymentCommand',
                        payload: { paymentReservationId: data.paymentReservationId },
                    };
                })
                .build(),

            SagaStep.create<OrderPlacementSagaData>('DecreaseInventory')
                .invoke(async (data) => ({
                    destination: 'inventory',
                    commandType: 'ReserveInventoryCommand',
                    payload: { orderId: data.orderId, items: data.items },
                }))
                .compensate(async (data) => ({
                    destination: 'inventory',
                    commandType: 'ReleaseInventoryCommand',
                    payload: { orderId: data.orderId },
                }))
                .build(),

            SagaStep.create<OrderPlacementSagaData>('ConfirmOrder')
                .withLocalInvoke(async (data) => {
                    // Local step — no Kafka
                    await this.confirmOrderUseCase.executeWithHooks({ orderId: data.orderId });
                    return data;
                })
                .build(),
        ];
    }

    onCompleted = (ctx) => Logger.log(`Saga completed: orderId=${ctx.sagaData.orderId}`);
    onFailed = (ctx) => Logger.error(`Saga failed: orderId=${ctx.sagaData.orderId}`);
}

// Trigger saga from use case
@Injectable()
export class CreateOrderUseCase extends BaseCommand<CreateOrderRequest, CreateOrderResponse> {
    constructor(
        private readonly orderRepo: OrderRepositoryPort,
        private readonly sagaManager: SagaManager,
    ) {
        super();
    }

    async execute(request: CreateOrderRequest): Promise<CreateOrderResponse> {
        const order = await this.orderRepo.save(/* ... */);

        const saga = await this.sagaManager.create<OrderPlacementSagaData>({
            sagaType: 'OrderPlacementSaga',
            initialData: {
                orderId: order.id.toValue(),
                customerId: request.customerId,
                totalAmount: request.totalAmount,
                items: request.items,
            },
            idempotencyId: request.idempotencyKey,
        });

        return { orderId: order.id.toValue(), sagaId: saga.sagaId };
    }
}
```

---

## PHẦN 7 — OPENAPI

### 7.1 Swagger Setup

```typescript
// libs/src/core/swagger/swagger-service.adapter.ts
@Injectable()
export class SwaggerServiceAdapter {
    constructor(private readonly app: NestExpressApplication) {}

    processRouters(routers: SwaggerRouter[]): void {
        const document = SwaggerModule.createDocument(
            this.app,
            new DocumentBuilder()
                .setTitle('Data Visualizer API')
                .setDescription('REST API for Data Visualizer Studio')
                .setVersion('1.0')
                .addBearerAuth()
                .addApiKey(
                    { type: 'apiKey', name: 'x-api-internal', in: 'header' },
                    'x-api-internal',
                )
                .addServer(process.env.API_URL || 'http://localhost:3000')
                .build(),
            {
                extraModels: [ApiResponseData, ApiErrorResponse, PaginatedResponse],
            },
        );

        SwaggerModule.setup('doc', this.app, document, {
            swaggerOptions: {
                persistAuthorization: true,
                filter: true,
                showRequestDuration: true,
            },
            customSiteTitle: 'Data Visualizer API',
        });
    }
}

// main.ts usage
const swaggerAdapter = new SwaggerServiceAdapter(app);
swaggerAdapter.processRouters([]);
```

### 7.2 Full DTO with Swagger

```typescript
// Response DTOs hierarchy
export class PaginationMetaDto {
    @ApiProperty({ example: 20 })
    perPage: number;

    @ApiProperty({ example: 1 })
    currentPage: number;

    @ApiProperty({ example: 100 })
    totalItems: number;

    @ApiProperty({ example: 5 })
    totalPages: number;
}

export class PaginatedOrderResponse {
    @ApiProperty({ type: [OrderDetailResponse] })
    data: OrderDetailResponse[];

    @ApiProperty({ type: PaginationMetaDto })
    metadata: PaginationMetaDto;

    @ApiProperty({ example: true })
    success: boolean;

    @ApiProperty({ example: 200 })
    statusCode: number;
}

// Controller with full Swagger docs
@Controller({ path: 'orders', version: '1' })
@ApiTags('Orders')
@ApiBearerAuth()
@UseGuards(JwtAuthGuard)
export class OrderCommandController {
    @Post()
    @HttpCode(HttpStatus.CREATED)
    @ApiOperation({
        summary: 'Create new order',
        description: 'Creates a new order with idempotency support',
    })
    @ApiCreatedResponse({ type: CreateOrderResponseDto, description: 'Order created successfully' })
    @ApiBadRequestResponse({ description: 'Validation failed' })
    @ApiUnauthorizedResponse({ description: 'Token invalid' })
    @ApiHeader({ name: 'x-idempotency-key', required: true, description: 'Unique idempotency key' })
    async create(@Body() dto: CreateOrderDto): Promise<CreateOrderResponseDto> {
        return this.createOrder.executeWithHooks(dto);
    }

    @Get()
    @ApiOperation({ summary: 'List orders with pagination' })
    @ApiOkResponse({ type: PaginatedOrderResponse })
    @ApiQuery({ name: 'page', required: false, type: Number, example: 1 })
    @ApiQuery({ name: 'limit', required: false, type: Number, example: 20 })
    @ApiQuery({ name: 'status', required: false, enum: ENUM_ORDER_STATUS })
    async list(@Query() query: ListOrdersQueryDto): Promise<PaginatedOrderResponse> {
        return this.listOrders.queryWithHooks(query);
    }
}
```

### 7.3 Mapped Types

```typescript
// Create DTO
export class CreateProductDto {
    @ApiProperty()
    @IsString()
    @MinLength(3)
    name: string;

    @ApiProperty()
    @IsPositive()
    price: number;

    @ApiProperty({ enum: ProductCategory })
    @IsEnum(ProductCategory)
    category: ProductCategory;

    @ApiPropertyOptional()
    @IsString()
    @IsOptional()
    description?: string;
}

// Update DTO — all optional
export class UpdateProductDto extends PartialType(CreateProductDto) {}

// Patch name only
export class PatchProductNameDto extends PickType(CreateProductDto, ['name']) {}

// Without price
export class CreateProductWithoutPriceDto extends OmitType(CreateProductDto, ['price']) {}

// Merge two DTOs
export class CreateProductWithMetaDto extends IntersectionType(CreateProductDto, MetadataDto) {}
```

---

## Domain Layer Examples (DDD Core)

### Aggregate Root

```typescript
// src/domain/order/entities/order.entity.ts
import { BaseAggregateRoot, Guard, Result, UniqueEntityID } from '@ddd/domain';
import { OrderCreatedEvent, OrderConfirmedEvent, OrderCancelledEvent } from '../events';

interface OrderProps {
    customerId: string;
    status: ENUM_ORDER_STATUS;
    items: OrderItemVO[];
    totalAmount: Money;
    note?: string;
    createdAt: Date;
    updatedAt: Date;
}

export class Order extends BaseAggregateRoot<OrderProps> {
    get customerId(): string {
        return this.props.customerId;
    }
    get status(): ENUM_ORDER_STATUS {
        return this.props.status;
    }
    get totalAmount(): Money {
        return this.props.totalAmount;
    }

    public static create(props: CreateOrderProps, id?: UniqueEntityID): Result<Order> {
        const guard = Guard.combine([
            Guard.againstNullOrUndefined(props.customerId, 'customerId'),
            Guard.againstEmptyArray(props.items, 'items'),
        ]);
        if (!guard.succeeded) return Result.fail(guard.message);

        const total = props.items.reduce((sum, item) => sum + item.price * item.quantity, 0);
        const order = new Order(
            {
                ...props,
                status: ENUM_ORDER_STATUS.PENDING,
                totalAmount: Money.create(total, 'VND').getValue(),
                createdAt: new Date(),
                updatedAt: new Date(),
            },
            id || new UniqueEntityID(),
        );

        order.addDomainEvent(
            OrderCreatedEvent.create(order.id.toValue(), {
                customerId: props.customerId,
                totalAmount: total,
            }),
        );

        return Result.ok(order);
    }

    public confirm(): Result<void> {
        if (this.props.status !== ENUM_ORDER_STATUS.PENDING) {
            return Result.fail(`Cannot confirm order in ${this.props.status} status`);
        }
        this._props.status = ENUM_ORDER_STATUS.CONFIRMED;
        this._props.updatedAt = new Date();
        this.addDomainEvent(OrderConfirmedEvent.create(this.id.toValue(), {}));
        return Result.ok();
    }

    public cancel(reason: string, policy: CancellationPolicyPort): Result<void> {
        const check = policy.canCancel(this, reason);
        if (check.isFailure) return check;

        this._props.status = ENUM_ORDER_STATUS.CANCELLED;
        this._props.updatedAt = new Date();
        this.addDomainEvent(OrderCancelledEvent.create(this.id.toValue(), { reason }));
        return Result.ok();
    }

    // Reconstitute from persistence
    public static reconstitute(props: OrderProps, id: UniqueEntityID): Order {
        return new Order(props, id);
    }
}

// Value Object
export class Money extends BaseValueObject<{ amount: number; currency: string }> {
    get amount(): number {
        return this.props.amount;
    }
    get currency(): string {
        return this.props.currency;
    }

    static create(amount: number, currency: string): Result<Money> {
        const guard = Guard.combine([
            Guard.againstNullOrUndefined(amount, 'amount'),
            Guard.againstAtMinValue(amount, 0, 'amount'),
            Guard.againstEnum(currency, ['VND', 'USD'], 'currency'),
        ]);
        if (!guard.succeeded) return Result.fail(guard.message);
        return Result.ok(new Money({ amount, currency }));
    }

    static zero(currency: string): Money {
        return new Money({ amount: 0, currency });
    }

    add(other: Money): Result<Money> {
        if (this.currency !== other.currency) return Result.fail('Currency mismatch');
        return Money.create(this.amount + other.amount, this.currency);
    }

    multiply(factor: number): Result<Money> {
        return Money.create(Math.round(this.amount * factor), this.currency);
    }
}
```

### Use Case Pattern

```typescript
// src/application/order/use-cases/create-order.use-case.ts
import { Injectable, Inject, Logger } from '@nestjs/common';
import { BaseCommand } from '@ddd/application';
import { UsecaseBadRequestException } from '@common/exceptions';
import { OrderRepositoryPort } from '@domain/order/repositories/order.repository.port';
import { Order } from '@domain/order/entities/order.entity';
import { SagaManager } from '@ddd/saga';

@Injectable()
export class CreateOrderUseCase extends BaseCommand<CreateOrderRequest, CreateOrderResponse> {
    private readonly logger = new Logger(CreateOrderUseCase.name);

    constructor(
        @Inject(OrderRepositoryPort)
        private readonly orderRepo: OrderRepositoryPort,
        private readonly pricingService: PricingDomainService,
        private readonly sagaManager: SagaManager,
    ) {
        super();
    }

    async execute(request: CreateOrderRequest): Promise<CreateOrderResponse> {
        // 1. Create domain entity
        const orderResult = Order.create({
            customerId: request.customerId,
            items: request.items.map((i) => OrderItemVO.create(i).getValue()),
        });
        if (orderResult.isFailure) {
            throw new UsecaseBadRequestException(orderResult.errors);
        }
        const order = orderResult.getValue();

        // 2. Apply discount via domain service
        const discountResult = this.pricingService.calculateDiscount(order, request.customerId);
        if (discountResult.isSuccess) {
            order.applyDiscount(discountResult.getValue());
        }

        // 3. Persist
        const saved = await this.orderRepo.save(order);

        // 4. Start saga for distributed transaction
        const saga = await this.sagaManager.create<OrderPlacementSagaData>({
            sagaType: 'OrderPlacementSaga',
            initialData: {
                orderId: saved.id.toValue(),
                totalAmount: saved.totalAmount.amount,
                items: request.items,
            },
            idempotencyId: request.idempotencyKey,
        });

        return { orderId: saved.id.toValue(), sagaId: saga.sagaId };
    }

    protected async validate(request: CreateOrderRequest): Promise<Result<void>> {
        if (!request.customerId) return Result.fail('customerId is required');
        if (!request.items?.length) return Result.fail('At least one item required');
        return Result.ok();
    }
}

// Query use case
@Injectable()
export class GetOrderDetailQuery extends BaseQuery<GetOrderRequest, OrderDetailResponse> {
    constructor(
        @Inject(OrderRepositoryPort)
        private readonly orderRepo: OrderRepositoryPort,
        private readonly orderMapper: OrderMapper,
    ) {
        super();
    }

    async query(request: GetOrderRequest): Promise<OrderDetailResponse> {
        const order = await this.orderRepo.findById(request.orderId);
        if (!order) throw new UsecaseNotFoundException(`Order ${request.orderId} not found`);
        return this.orderMapper.toDto(order);
    }
}
```
