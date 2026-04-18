# Clean Architecture Go Rules

- Keep dependency direction one-way: presentation -> application -> domain <- infrastructure.
- Domain packages must not import Gin, GORM, SQL drivers, Redis clients, Kafka clients, RabbitMQ clients, or logging/tracing frameworks that encode infrastructure concerns.
- Application packages may orchestrate transactions and ports, but business invariants still belong in domain types and services.
- Infrastructure packages implement ports and translate external details into internal contracts.
- `cmd/*` is composition only. No business branching beyond startup and wiring.
- `pkg/*` must remain reusable and must not depend on service-specific `internal/*` packages.
