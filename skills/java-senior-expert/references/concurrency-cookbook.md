# Java Concurrency Cookbook — Production-Tested Recipes

## Mental Model First

Before writing any concurrent code, answer:
1. **Visibility**: Can thread B see changes made by thread A?
2. **Atomicity**: Can the operation be interrupted mid-way?
3. **Ordering**: Can the compiler/CPU reorder these instructions?

If "yes" to any → you need synchronization.

---

## CompletableFuture — The Right Way

```java
// ✅ Fan-out: call 3 services in parallel, combine results
public OrderSummary buildOrderSummary(String orderId) {
    CompletableFuture<OrderDetails>  detailsFuture  = CompletableFuture
        .supplyAsync(() -> orderClient.getDetails(orderId), ioExecutor);
    CompletableFuture<PaymentStatus> paymentFuture  = CompletableFuture
        .supplyAsync(() -> paymentClient.getStatus(orderId), ioExecutor);
    CompletableFuture<ShipmentInfo>  shipmentFuture = CompletableFuture
        .supplyAsync(() -> shipmentClient.getInfo(orderId), ioExecutor);

    return CompletableFuture.allOf(detailsFuture, paymentFuture, shipmentFuture)
        .thenApply(v -> OrderSummary.builder()
            .details(detailsFuture.join())
            .payment(paymentFuture.join())
            .shipment(shipmentFuture.join())
            .build())
        .orTimeout(5, TimeUnit.SECONDS)          // never block indefinitely
        .exceptionally(ex -> OrderSummary.degraded(orderId, ex.getMessage()))
        .join();
}
```

**Executor matters:**
- `ForkJoinPool.commonPool()` — for CPU-bound tasks only
- Custom `ThreadPoolTaskExecutor` — for IO-bound tasks (DB, HTTP)

---

## Atomic Operations

```java
// ✅ Thread-safe counter without synchronized
private final AtomicLong requestCount = new AtomicLong(0);
private final AtomicReference<Status> status = new AtomicReference<>(Status.IDLE);

// CAS — compare-and-swap (lock-free)
public boolean tryStart() {
    return status.compareAndSet(Status.IDLE, Status.RUNNING);
}

// AtomicLong for high-throughput counters; LongAdder for even higher throughput
private final LongAdder errorCount = new LongAdder();
errorCount.increment();
long total = errorCount.sum();
```

---

## ReentrantLock — When synchronized Isn't Enough

```java
public class OrderProcessor {
    private final ReentrantLock lock = new ReentrantLock(true); // fair lock

    public boolean tryProcess(Order order, long timeoutMs) {
        try {
            if (!lock.tryLock(timeoutMs, TimeUnit.MILLISECONDS)) {
                log.warn("Could not acquire lock within {}ms for order={}", timeoutMs, order.getId());
                return false;
            }
            try {
                // critical section
                process(order);
                return true;
            } finally {
                lock.unlock(); // ALWAYS in finally
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt(); // restore interrupt status
            return false;
        }
    }
}
```

---

## ReadWriteLock — Read-Heavy Workloads

```java
public class ConfigCache {
    private final ReadWriteLock rwLock = new ReentrantReadWriteLock();
    private final Lock readLock  = rwLock.readLock();
    private final Lock writeLock = rwLock.writeLock();
    private volatile Map<String, String> config = new HashMap<>();

    public String get(String key) {
        readLock.lock();
        try {
            return config.get(key);  // multiple readers in parallel
        } finally {
            readLock.unlock();
        }
    }

    public void refresh(Map<String, String> newConfig) {
        writeLock.lock();
        try {
            this.config = Map.copyOf(newConfig);  // exclusive write
        } finally {
            writeLock.unlock();
        }
    }
}
```

---

## Semaphore — Rate Limiting / Resource Pool

```java
public class ExternalApiClient {
    // Allow max 10 concurrent calls to external API
    private final Semaphore semaphore = new Semaphore(10);

    public Response call(Request request) {
        try {
            semaphore.acquire();
            try {
                return httpClient.execute(request);
            } finally {
                semaphore.release(); // ALWAYS release in finally
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new ServiceException("Interrupted while waiting for permit");
        }
    }
}
```

---

## BlockingQueue — Producer-Consumer

```java
public class EventProcessor {
    private final BlockingQueue<DomainEvent> queue = new LinkedBlockingQueue<>(1000);
    private final ExecutorService consumer = Executors.newSingleThreadExecutor();

    @PostConstruct
    public void start() {
        consumer.submit(() -> {
            while (!Thread.currentThread().isInterrupted()) {
                try {
                    DomainEvent event = queue.poll(1, TimeUnit.SECONDS); // don't block forever
                    if (event != null) process(event);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                }
            }
        });
    }

    public boolean submit(DomainEvent event) {
        return queue.offer(event); // non-blocking; returns false if full
    }

    @PreDestroy
    public void stop() {
        consumer.shutdownNow();
    }
}
```

---

## ThreadLocal — Use with Extreme Care

```java
// ✅ Correct: used in filter/interceptor, ALWAYS remove in finally
public class RequestContextFilter implements Filter {
    private static final ThreadLocal<RequestContext> CONTEXT = new ThreadLocal<>();

    public static RequestContext get() { return CONTEXT.get(); }

    @Override
    public void doFilter(ServletRequest req, ServletResponse res, FilterChain chain)
            throws IOException, ServletException {
        try {
            CONTEXT.set(new RequestContext((HttpServletRequest) req));
            chain.doFilter(req, res);
        } finally {
            CONTEXT.remove(); // ⚠️ CRITICAL — thread pool reuses threads; leak = cross-request contamination
        }
    }
}
```

---

## Volatile — Visibility Only

```java
// ✅ volatile ensures visibility across threads (no caching in CPU registers)
public class CircuitBreaker {
    private volatile boolean open = false;      // reads always see latest write
    private volatile long lastFailureTime = 0;

    // ❌ volatile does NOT make compound operations atomic:
    // open = !open  is NOT atomic even with volatile
    // Use AtomicBoolean for that
}
```

---

## Common Concurrency Bugs (Production War Stories)

### Bug 1: HashMap in Shared Context → ConcurrentModificationException

```java
// ❌ Race condition
private Map<String, Session> sessions = new HashMap<>();

// ✅ Fix
private Map<String, Session> sessions = new ConcurrentHashMap<>();
```

### Bug 2: SimpleDateFormat as Static Field

```java
// ❌ SimpleDateFormat is NOT thread-safe
private static final SimpleDateFormat SDF = new SimpleDateFormat("yyyy-MM-dd");

// ✅ Option 1: DateTimeFormatter (thread-safe, immutable)
private static final DateTimeFormatter FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd");

// ✅ Option 2: ThreadLocal
private static final ThreadLocal<SimpleDateFormat> SDF = 
    ThreadLocal.withInitial(() -> new SimpleDateFormat("yyyy-MM-dd"));
```

### Bug 3: Calling Blocking Code in Reactive Pipeline

```java
// ❌ Blocks Netty event loop thread
Mono.fromCallable(() -> jdbcTemplate.query(...))  // blocking JDBC!
    .subscribe(...);

// ✅ Schedule on bounded elastic scheduler
Mono.fromCallable(() -> jdbcTemplate.query(...))
    .subscribeOn(Schedulers.boundedElastic())
    .subscribe(...);
```

### Bug 4: Double-Checked Locking Without volatile

```java
// ❌ Broken — object can be partially initialized
private static ExpensiveObject instance;
if (instance == null) {
    synchronized (ExpensiveObject.class) {
        if (instance == null) instance = new ExpensiveObject();
    }
}

// ✅ Fix — volatile prevents instruction reordering
private static volatile ExpensiveObject instance;

// ✅ Better — use enum singleton or holder pattern
private static class Holder {
    static final ExpensiveObject INSTANCE = new ExpensiveObject();
}
```

### Bug 5: Executor Not Shut Down → Thread Leak

```java
// ✅ Always register shutdown hook or use @PreDestroy
@PreDestroy
public void shutdown() {
    executor.shutdown();
    try {
        if (!executor.awaitTermination(30, TimeUnit.SECONDS)) {
            executor.shutdownNow();
        }
    } catch (InterruptedException e) {
        executor.shutdownNow();
        Thread.currentThread().interrupt();
    }
}
```
