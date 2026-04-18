# Design Patterns in Java — Applied, Not Academic

## When to Use What (Decision Tree)

```
Need to CREATE objects?
├── Complex construction (many params)    → Builder
├── Different object families             → Abstract Factory
├── Decouple creation from usage          → Factory Method
└── Only one instance needed              → Singleton (prefer enum)

Need to STRUCTURE code?
├── Add behavior without modifying class  → Decorator
├── Simplify complex subsystem            → Facade
├── Make classes work together            → Adapter
└── Share state between many objects      → Flyweight

Need to CONTROL behavior?
├── Swappable algorithms                  → Strategy
├── Respond to state changes              → Observer / Event
├── Step-by-step algorithm skeleton       → Template Method
├── Request as object                     → Command
└── State machine                         → State
```

---

## Builder — The Most Used Pattern in Java

```java
// ✅ Immutable value object with Builder
public final class CreateOrderCommand {
    private final String customerId;
    private final List<OrderItem> items;
    private final String deliveryAddress;
    private final PaymentMethod paymentMethod;
    private final Instant requestedDeliveryDate; // nullable

    private CreateOrderCommand(Builder builder) {
        this.customerId = Objects.requireNonNull(builder.customerId, "customerId required");
        this.items = List.copyOf(builder.items); // defensive copy
        this.deliveryAddress = Objects.requireNonNull(builder.deliveryAddress);
        this.paymentMethod = Objects.requireNonNull(builder.paymentMethod);
        this.requestedDeliveryDate = builder.requestedDeliveryDate; // optional
    }

    public static Builder builder(String customerId) {
        return new Builder(customerId);
    }

    public static class Builder {
        private final String customerId;
        private List<OrderItem> items = new ArrayList<>();
        private String deliveryAddress;
        private PaymentMethod paymentMethod;
        private Instant requestedDeliveryDate;

        private Builder(String customerId) { this.customerId = customerId; }

        public Builder item(OrderItem item) { this.items.add(item); return this; }
        public Builder deliveryAddress(String addr) { this.deliveryAddress = addr; return this; }
        public Builder paymentMethod(PaymentMethod pm) { this.paymentMethod = pm; return this; }
        public Builder requestedDelivery(Instant date) { this.requestedDeliveryDate = date; return this; }
        public CreateOrderCommand build() { return new CreateOrderCommand(this); }
    }
}
```

---

## Strategy — Runtime Behavior Swapping

```java
// Domain: different shipping cost calculators
public interface ShippingCalculator {
    Money calculate(Order order);
}

@Component("standard")
public class StandardShippingCalculator implements ShippingCalculator { ... }

@Component("express")
public class ExpressShippingCalculator implements ShippingCalculator { ... }

@Component("free")
public class FreeShippingCalculator implements ShippingCalculator { ... }

// Factory that selects strategy
@Service
@RequiredArgsConstructor
public class ShippingService {
    private final Map<String, ShippingCalculator> calculators; // Spring injects all implementations!

    public Money calculateShipping(Order order) {
        String type = determineShippingType(order);
        ShippingCalculator calculator = calculators.get(type);
        if (calculator == null) throw new IllegalStateException("No calculator for type: " + type);
        return calculator.calculate(order);
    }
}
```

---

## Decorator — Adding Cross-Cutting Behavior

```java
// Core interface
public interface OrderRepository {
    Order findById(String id);
    void save(Order order);
}

// Real implementation
@Repository
public class JpaOrderRepository implements OrderRepository { ... }

// Decorator: adds caching
public class CachingOrderRepository implements OrderRepository {
    private final OrderRepository delegate;
    private final Cache<String, Order> cache;

    public Order findById(String id) {
        return cache.get(id, key -> delegate.findById(key));
    }

    public void save(Order order) {
        delegate.save(order);
        cache.invalidate(order.getId()); // invalidate after write
    }
}

// Decorator: adds metrics
public class MeteredOrderRepository implements OrderRepository {
    private final OrderRepository delegate;
    private final MeterRegistry registry;

    public Order findById(String id) {
        return Timer.builder("order.repository.find")
            .register(registry)
            .recordCallable(() -> delegate.findById(id));
    }
}
// Stack: Metered → Caching → JPA
```

---

## Observer / Event — Loose Coupling with Spring

```java
// Domain event (plain record)
public record OrderPlacedEvent(String orderId, String customerId, Money total, Instant occurredAt) {}

// Publisher (in domain/service layer)
@Service
@RequiredArgsConstructor
public class OrderService {
    private final ApplicationEventPublisher eventPublisher;

    @Transactional
    public Order placeOrder(PlaceOrderCommand cmd) {
        Order order = Order.create(cmd);
        orderRepository.save(order);
        // Event published AFTER commit (via @TransactionalEventListener)
        eventPublisher.publishEvent(new OrderPlacedEvent(order.getId(), ...));
        return order;
    }
}

// Listener: runs in new transaction after original commits
@Component
public class OrderNotificationListener {

    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    @Async
    public void onOrderPlaced(OrderPlacedEvent event) {
        emailService.sendConfirmation(event.customerId(), event.orderId());
    }
}
```

---

## Template Method — Shared Algorithm, Variable Steps

```java
// Abstract base: defines the skeleton
public abstract class ReportGenerator {

    // Template method — final to prevent override of algorithm structure
    public final Report generate(ReportRequest request) {
        validateRequest(request);
        List<Record> data = fetchData(request);           // hook
        List<Record> processed = processData(data);       // hook
        return formatReport(request, processed);           // hook
    }

    protected void validateRequest(ReportRequest req) {
        // default validation — can be overridden
        Objects.requireNonNull(req.getStartDate());
        Objects.requireNonNull(req.getEndDate());
    }

    protected abstract List<Record> fetchData(ReportRequest request);
    protected abstract List<Record> processData(List<Record> rawData);
    protected abstract Report formatReport(ReportRequest request, List<Record> data);
}

@Component
public class SalesReportGenerator extends ReportGenerator {
    @Override
    protected List<Record> fetchData(ReportRequest req) { ... }
    @Override
    protected List<Record> processData(List<Record> data) { ... }
    @Override
    protected Report formatReport(ReportRequest req, List<Record> data) { ... }
}
```

---

## Saga Pattern — Distributed Transactions

**Problem**: Transferring money across two microservices — what if step 2 fails?

```
Choreography-based Saga (event-driven):

[OrderService]         [PaymentService]       [InventoryService]
    │ OrderCreated ──────────→ │                       │
    │                  PaymentProcessed ───────────────→ │
    │                          │               InventoryReserved
    │ ←─────────────────────────────────────────────── │
    │ OrderConfirmed            │                       │

On failure: each service publishes a compensating event:
PaymentFailed → OrderService cancels order
InventoryFailed → PaymentService refunds → OrderService cancels
```

```java
// Each step has a compensating action
@Component
public class PaymentSagaStep {

    @KafkaListener(topics = "order.created")
    public void handleOrderCreated(OrderCreatedEvent event) {
        try {
            Payment payment = paymentService.charge(event.customerId(), event.total());
            kafkaTemplate.send("payment.processed", new PaymentProcessedEvent(...));
        } catch (PaymentException ex) {
            kafkaTemplate.send("payment.failed", new PaymentFailedEvent(event.orderId(), ex.getMessage()));
        }
    }
}
```

---

## Enum Singleton — Safest Singleton in Java

```java
// Thread-safe, serialization-safe, reflection-safe
public enum DatabaseConnectionFactory {
    INSTANCE;

    private final DataSource dataSource = createDataSource();

    private DataSource createDataSource() { ... }

    public Connection getConnection() throws SQLException {
        return dataSource.getConnection();
    }
}

// Usage
Connection conn = DatabaseConnectionFactory.INSTANCE.getConnection();
```

---

## Anti-Patterns — Recognize and Refactor

| Anti-Pattern | Problem | Solution |
|---|---|---|
| **God Class** | 2000-line service doing everything | Split by business capability (SRP) |
| **Anemic Domain Model** | Entity with only getters/setters; logic in services | Move behavior into entity |
| **Service Locator** | `context.getBean(OrderService.class)` | Constructor injection |
| **Magic Numbers** | `if (status == 3)` | Named constants / enums |
| **Primitive Obsession** | `String email, String phone, String address` everywhere | Value Objects: `Email`, `PhoneNumber` |
| **Feature Envy** | Method uses more data from another class than its own | Move the method |
| **Shotgun Surgery** | One change requires many files | Cohesion problem → restructure |
