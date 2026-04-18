# Clean Architecture & DDD in Go

## Table of Contents
1. [Project Structure Deep Dive](#structure)
2. [Repository Pattern](#repository)
3. [Application Layer / Use Cases](#application)
4. [CQRS in Go](#cqrs)
5. [Domain Events](#events)
6. [Anti-Patterns](#antipatterns)

---

## 1. Project Structure Deep Dive

### domain/ — Zero External Dependencies

```go
// domain/entity/order.go
package entity

// Unexported fields — encapsulation enforced at package level
type Order struct {
    id         OrderID
    customerID CustomerID
    items      []OrderItem
    status     OrderStatus
    total      valueobject.Money
    events     []event.DomainEvent
}

// Constructor enforces invariants
func NewOrder(id OrderID, customerID CustomerID) (*Order, error) {
    if id == ""         { return nil, ErrInvalidOrderID }
    if customerID == "" { return nil, ErrInvalidCustomerID }
    return &Order{
        id:         id,
        customerID: customerID,
        status:     StatusDraft,
        items:      make([]OrderItem, 0),
    }, nil
}

// Getters — expose only what's needed
func (o *Order) ID() OrderID           { return o.id }
func (o *Order) Status() OrderStatus   { return o.status }
func (o *Order) Total() Money          { return o.total }
func (o *Order) Items() []OrderItem    { return append([]OrderItem{}, o.items...) } // defensive copy
```

### domain/repository/ — Interfaces Only

```go
// domain/repository/order_repository.go
package repository

import (
    "context"
    "yourapp/internal/domain/entity"
)

// OrderRepository — defined in domain, implemented in infrastructure
type OrderRepository interface {
    FindByID(ctx context.Context, id entity.OrderID) (*entity.Order, error)
    FindByCustomerID(ctx context.Context, customerID entity.CustomerID, page Pagination) ([]*entity.Order, int64, error)
    Save(ctx context.Context, order *entity.Order) error
    Update(ctx context.Context, order *entity.Order) error
    Delete(ctx context.Context, id entity.OrderID) error
}

// Pagination value object
type Pagination struct {
    Page     int
    PageSize int
}

func (p Pagination) Offset() int { return (p.Page - 1) * p.PageSize }
func (p Pagination) Limit() int  { return p.PageSize }
```

---

## 2. Repository Pattern — Infrastructure Implementation

```go
// infrastructure/persistence/gorm_order_repository.go
package persistence

type GormOrderRepository struct {
    db     *gorm.DB
    mapper OrderMapper
}

func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
    return &GormOrderRepository{db: db, mapper: OrderMapper{}}
}

func (r *GormOrderRepository) FindByID(ctx context.Context, id entity.OrderID) (*entity.Order, error) {
    var model OrderModel
    result := r.db.WithContext(ctx).
        Preload("Items").
        Where("id = ? AND deleted_at IS NULL", string(id)).
        First(&model)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, repository.ErrOrderNotFound
        }
        return nil, fmt.Errorf("GormOrderRepository.FindByID: %w", result.Error)
    }
    return r.mapper.ToDomain(&model)
}

// Persistence Model — separate from domain entity (anti-corruption layer)
type OrderModel struct {
    ID         string         `gorm:"primaryKey"`
    CustomerID string         `gorm:"not null;index"`
    Status     string         `gorm:"not null"`
    TotalCents int64          `gorm:"not null"`
    Currency   string         `gorm:"not null"`
    Items      []OrderItemModel `gorm:"foreignKey:OrderID"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// Mapper — converts between domain entity and persistence model
type OrderMapper struct{}

func (m OrderMapper) ToDomain(model *OrderModel) (*entity.Order, error) {
    // reconstruct entity from persistence model
    // use Reconstitute (not NewXxx) to bypass creation invariants
    return entity.ReconstitutueOrder(entity.OrderID(model.ID), ...)
}

func (m OrderMapper) ToModel(order *entity.Order) *OrderModel { ... }
```

---

## 3. Application Layer / Use Cases

```go
// application/command/place_order_handler.go
package command

type PlaceOrderCommand struct {
    CustomerID string
    Items      []PlaceOrderItemDTO
    AddressID  string
}

type PlaceOrderResult struct {
    OrderID string
    Total   float64
}

type PlaceOrderHandler struct {
    orderRepo    repository.OrderRepository
    productRepo  repository.ProductRepository
    eventPub     event.Publisher
    uow          UnitOfWork
}

func NewPlaceOrderHandler(
    orderRepo   repository.OrderRepository,
    productRepo repository.ProductRepository,
    eventPub    event.Publisher,
    uow         UnitOfWork,
) *PlaceOrderHandler {
    return &PlaceOrderHandler{orderRepo, productRepo, eventPub, uow}
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrderCommand) (*PlaceOrderResult, error) {
    return h.uow.Execute(ctx, func(ctx context.Context) (*PlaceOrderResult, error) {
        order, err := entity.NewOrder(entity.NewOrderID(), entity.CustomerID(cmd.CustomerID))
        if err != nil {
            return nil, fmt.Errorf("PlaceOrderHandler: create order: %w", err)
        }

        for _, item := range cmd.Items {
            product, err := h.productRepo.FindByID(ctx, entity.ProductID(item.ProductID))
            if err != nil {
                return nil, fmt.Errorf("PlaceOrderHandler: fetch product %s: %w", item.ProductID, err)
            }
            if err := order.AddItem(product, item.Qty); err != nil {
                return nil, fmt.Errorf("PlaceOrderHandler: add item: %w", err)
            }
        }

        if err := order.Place(); err != nil {
            return nil, fmt.Errorf("PlaceOrderHandler: place: %w", err)
        }

        if err := h.orderRepo.Save(ctx, order); err != nil {
            return nil, fmt.Errorf("PlaceOrderHandler: save: %w", err)
        }

        // Publish domain events AFTER successful save, within same transaction (Outbox)
        for _, evt := range order.PullEvents() {
            if err := h.eventPub.Publish(ctx, evt); err != nil {
                return nil, fmt.Errorf("PlaceOrderHandler: publish event: %w", err)
            }
        }

        return &PlaceOrderResult{OrderID: string(order.ID()), Total: order.Total().AsFloat()}, nil
    })
}
```

---

## 4. CQRS in Go

```go
// Separate read model — optimized for queries, no domain logic
// application/query/get_order_handler.go

type GetOrderQuery struct {
    OrderID    string
    CustomerID string // for authorization check
}

type OrderDetailDTO struct {
    ID         string
    Status     string
    Total      float64
    Currency   string
    Items      []OrderItemDTO
    CreatedAt  time.Time
}

// Read repository — different interface from write repository
type OrderReadRepository interface {
    FindDetailByID(ctx context.Context, orderID, customerID string) (*OrderDetailDTO, error)
}

type GetOrderHandler struct {
    readRepo OrderReadRepository
}

func (h *GetOrderHandler) Handle(ctx context.Context, q GetOrderQuery) (*OrderDetailDTO, error) {
    dto, err := h.readRepo.FindDetailByID(ctx, q.OrderID, q.CustomerID)
    if err != nil {
        return nil, fmt.Errorf("GetOrderHandler: %w", err)
    }
    return dto, nil
}

// Read-side uses raw SQL / sqlx for performance — no ORM overhead
type SqlxOrderReadRepository struct {
    db *sqlx.DB
}

func (r *SqlxOrderReadRepository) FindDetailByID(ctx context.Context, orderID, customerID string) (*OrderDetailDTO, error) {
    const query = `
        SELECT o.id, o.status, o.total_cents, o.currency, o.created_at,
               oi.product_id, oi.name, oi.qty, oi.unit_price_cents
        FROM orders o
        JOIN order_items oi ON oi.order_id = o.id
        WHERE o.id = $1 AND o.customer_id = $2 AND o.deleted_at IS NULL
    `
    // ... scan and map to DTO
}
```

---

## 5. Domain Events

```go
// domain/event/events.go
package event

type DomainEvent interface {
    AggregateID() string
    EventType()   string
    OccurredAt()  time.Time
}

type OrderPlacedEvent struct {
    aggregateID string
    CustomerID  string
    Total       int64
    Currency    string
    occurredAt  time.Time
}

func (e OrderPlacedEvent) AggregateID() string { return e.aggregateID }
func (e OrderPlacedEvent) EventType()   string { return "order.placed" }
func (e OrderPlacedEvent) OccurredAt()  time.Time { return e.occurredAt }

// Publisher interface — in domain
type Publisher interface {
    Publish(ctx context.Context, event DomainEvent) error
}

// Outbox implementation — in infrastructure
type OutboxPublisher struct {
    db *gorm.DB
}

func (p *OutboxPublisher) Publish(ctx context.Context, evt DomainEvent) error {
    payload, err := json.Marshal(evt)
    if err != nil {
        return fmt.Errorf("OutboxPublisher: marshal: %w", err)
    }
    outbox := OutboxMessage{
        ID:          uuid.New().String(),
        EventType:   evt.EventType(),
        AggregateID: evt.AggregateID(),
        Payload:     payload,
        CreatedAt:   time.Now(),
    }
    return p.db.WithContext(ctx).Create(&outbox).Error
}
```

---

## 6. Anti-Patterns to Call Out

| Anti-Pattern | Problem | Fix |
|---|---|---|
| Anemic domain model | Entity has only getters/setters; logic in service | Move invariant logic into entity methods |
| Fat repository | Repository contains business rules | Business logic belongs in domain service |
| Importing infrastructure from domain | Creates circular deps, breaks testability | Domain defines interfaces; infra implements |
| Using `interface{}` for everything | Loses type safety | Use generics (Go 1.18+) or concrete types |
| Skipping UnitOfWork | Multiple repos, partial commit | Use transaction wrapper (UoW pattern) |
| HTTP handler calling repo directly | Bypasses domain logic | Handler → Use case → Domain → Repository |
| Global state / init() side effects | Hard to test, hidden deps | Explicit wiring in main.go |
