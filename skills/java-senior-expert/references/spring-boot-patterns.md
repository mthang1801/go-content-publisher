# Spring Boot Architecture Patterns & Anti-Patterns

## Layered Architecture

```
Controller (HTTP boundary)
    ↓
Service (Business logic + @Transactional)
    ↓
Repository (Data access — Spring Data JPA / custom)
    ↓
Domain Model (Entities, Value Objects)
```

**Rule**: Dependencies only point **downward**. Never inject a Controller into a Service.

---

## Constructor Injection — Always

```java
// ✅ Correct
@Service
@RequiredArgsConstructor  // Lombok
public class OrderService {
    private final OrderRepository orderRepository;
    private final PaymentGateway paymentGateway;
    private final ApplicationEventPublisher eventPublisher;
}

// ❌ Never do this
@Service
public class OrderService {
    @Autowired
    private OrderRepository orderRepository; // field injection — untestable, hides deps
}
```

---

## @Transactional — Understand What You're Doing

```java
@Service
public class OrderService {

    // ✅ Propagation.REQUIRED (default): joins existing tx or creates new one
    @Transactional
    public Order placeOrder(PlaceOrderCommand cmd) { ... }

    // ✅ REQUIRES_NEW: always create a new tx (useful for audit logs that must survive rollback)
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void saveAuditLog(AuditEntry entry) { ... }

    // ✅ Read-only: enables query optimizations, prevents accidental flush
    @Transactional(readOnly = true)
    public Page<Order> findOrders(Pageable pageable) { ... }

    // ❌ Transactional on private method — Spring AOP won't intercept it!
    @Transactional
    private void internalMethod() { ... }
}
```

**Common pitfall**: self-invocation bypasses the proxy — `this.placeOrder()` won't start a transaction.

---

## Exception Handling — Global Pattern

```java
@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(OrderNotFoundException.class)
    public ProblemDetail handleOrderNotFound(OrderNotFoundException ex, HttpServletRequest req) {
        ProblemDetail pd = ProblemDetail.forStatusAndDetail(HttpStatus.NOT_FOUND, ex.getMessage());
        pd.setTitle("Order Not Found");
        pd.setInstance(URI.create(req.getRequestURI()));
        return pd;
    }

    @ExceptionHandler(ConstraintViolationException.class)
    public ProblemDetail handleValidation(ConstraintViolationException ex) {
        ProblemDetail pd = ProblemDetail.forStatus(HttpStatus.BAD_REQUEST);
        pd.setTitle("Validation Failed");
        pd.setProperty("violations", ex.getConstraintViolations().stream()
            .map(v -> v.getPropertyPath() + ": " + v.getMessage())
            .collect(Collectors.toList()));
        return pd;
    }

    @ExceptionHandler(Exception.class)
    public ProblemDetail handleUnexpected(Exception ex) {
        log.error("Unhandled exception", ex);
        return ProblemDetail.forStatusAndDetail(HttpStatus.INTERNAL_SERVER_ERROR,
            "An unexpected error occurred. Please contact support.");
    }
}
```

---

## ConfigurationProperties — Type-Safe Config

```java
@ConfigurationProperties(prefix = "payment")
@Validated
public record PaymentConfig(
    @NotBlank String apiUrl,
    @NotBlank String apiKey,
    @Min(1) @Max(30) int timeoutSeconds,
    @Min(1) @Max(5)  int maxRetries
) {}

// application.yml
// payment:
//   api-url: https://payment.example.com
//   api-key: ${PAYMENT_API_KEY}  # from env var
//   timeout-seconds: 10
//   max-retries: 3
```

---

## Async with ThreadPoolTaskExecutor

```java
@Configuration
@EnableAsync
public class AsyncConfig {

    @Bean("orderProcessingExecutor")
    public Executor orderProcessingExecutor() {
        ThreadPoolTaskExecutor executor = new ThreadPoolTaskExecutor();
        executor.setCorePoolSize(10);
        executor.setMaxPoolSize(50);
        executor.setQueueCapacity(100);
        executor.setThreadNamePrefix("order-proc-");
        executor.setRejectedExecutionHandler(new ThreadPoolExecutor.CallerRunsPolicy());
        executor.initialize();
        return executor;
    }
}

@Service
public class NotificationService {

    @Async("orderProcessingExecutor")
    public CompletableFuture<Void> sendOrderConfirmation(Order order) {
        // non-blocking email send
        emailClient.send(order.getCustomerEmail(), buildTemplate(order));
        return CompletableFuture.completedFuture(null);
    }
}
```

---

## Spring Security — Modern Config

```java
@Configuration
@EnableWebSecurity
public class SecurityConfig {

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        return http
            .csrf(csrf -> csrf.disable())  // OK for stateless JWT APIs
            .sessionManagement(sm -> sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/actuator/health", "/actuator/info").permitAll()
                .requestMatchers(HttpMethod.POST, "/api/auth/**").permitAll()
                .requestMatchers("/api/admin/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            .addFilterBefore(jwtAuthFilter, UsernamePasswordAuthenticationFilter.class)
            .build();
    }
}
```

---

## Outbox Pattern — Reliable Event Publishing

**Problem**: You save to DB and publish to Kafka — what if Kafka is down?

```java
@Transactional
public Order placeOrder(PlaceOrderCommand cmd) {
    Order order = Order.create(cmd);
    orderRepository.save(order);

    // Save event in SAME transaction — never lost
    OutboxEvent event = OutboxEvent.of("OrderPlaced", toJson(order));
    outboxRepository.save(event);

    // Separate scheduler polls outbox and publishes to Kafka
    return order;
}
```

---

## Anti-Patterns (Call These Out in Code Review)

| Anti-Pattern | Problem | Fix |
|---|---|---|
| `@Autowired` field injection | Hides dependencies, breaks unit tests | Constructor injection |
| Catching `Exception` silently | Hides bugs, hard to debug | Specific exceptions with context |
| `@Transactional` on private methods | Spring AOP won't intercept | Move to public method |
| Returning `null` from service | Forces null checks everywhere | `Optional<T>` or throw |
| `SELECT *` in JPQL | Fetches unnecessary data | Projections / DTOs |
| Loading all records without pagination | OOM at scale | `Pageable` |
| Hardcoded credentials | Security risk | `@ConfigurationProperties` + env vars |
| Business logic in Controllers | Untestable, violates SRP | Move to Service layer |
