# System Architecture

This document describes the architecture and design decisions for the Voucher & Payment Service.

## Table of Contents

- [Overview](#overview)
- [Architecture Patterns](#architecture-patterns)
- [Component Design](#component-design)
- [Data Flow](#data-flow)
- [Design Decisions](#design-decisions)
- [Scalability](#scalability)

## Overview

The Voucher & Payment Service is a production-grade microservice built with Go that follows clean architecture principles and implements industry-standard patterns for distributed systems.

### Key Characteristics

- **Language**: Go 1.21+
- **Architecture**: Clean Architecture (Hexagonal)
- **Communication**: gRPC + REST (via grpc-gateway)
- **Database**: MySQL 8.0 (ACID transactions)
- **Cache**: Redis (write-through strategy)
- **Message Queue**: RabbitMQ (event-driven)
- **Observability**: OpenTelemetry + Jaeger

## Architecture Patterns

### 1. Clean Architecture

The system follows Uncle Bob's Clean Architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────┐
│           External Interfaces               │
│  (gRPC Handlers, REST Gateway, CLI)         │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│          Application Services               │
│   (Business Logic, Orchestration)           │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│           Domain Layer                      │
│   (Entities, Business Rules, Errors)        │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│        Infrastructure Layer                 │
│  (Database, Cache, Queue, Observability)    │
└─────────────────────────────────────────────┘
```

**Benefits:**
- Independent of frameworks
- Testable business logic
- Independent of UI/Database
- Easy to maintain and extend

### 2. Repository Pattern

Data access is abstracted through repository interfaces:

```go
// Interface in domain/service layer
type VoucherRepository interface {
    GetByID(ctx context.Context, id string) (*Voucher, error)
    Search(ctx context.Context, params SearchParams) ([]*Voucher, int32, error)
}

// Implementation in infrastructure layer
type MySQLVoucherRepository struct {
    db *sql.DB
}
```

**Benefits:**
- Decouples business logic from data source
- Easy to swap implementations
- Simplifies testing with mocks

### 3. Unit of Work Pattern

Database transactions are managed through Unit of Work:

```go
type UnitOfWork interface {
    Begin(ctx context.Context) error
    Commit() error
    Rollback() error
    Vouchers() VoucherRepository
    Wallets() WalletRepository
    // ... other repositories
}
```

**Benefits:**
- Atomic operations across multiple repositories
- Consistent transaction management
- Prevents partial updates

### 4. CQRS (Command Query Responsibility Segregation)

Read and write operations are separated:

- **Commands**: Modify state (BuyVoucher, AddMoney)
- **Queries**: Read state (SearchVouchers, GetBalance)

**Benefits:**
- Optimize reads and writes independently
- Scale read/write workloads differently
- Clear separation of concerns

## Component Design

### API Layer

**Responsibilities:**
- Protocol handling (gRPC/HTTP)
- Request validation
- Response formatting
- Error mapping

**Components:**
- `internal/grpc/handlers.go` - gRPC service implementations
- `internal/grpc/interceptors.go` - Middleware (logging, auth, tracing)
- `internal/grpc/server.go` - Server setup and configuration

### Service Layer

**Responsibilities:**
- Business logic orchestration
- Transaction coordination
- Cache management
- Event publishing

**Components:**
- `internal/service/voucher_service.go` - Voucher operations
- `internal/service/wallet_service.go` - Wallet operations

**Design Patterns:**
- Dependency Injection
- Strategy Pattern (payment methods)
- Template Method (transaction flow)

### Domain Layer

**Responsibilities:**
- Business entities
- Domain rules
- Custom errors
- Value objects

**Components:**
- `internal/domain/models.go` - Entity definitions
- `internal/domain/errors.go` - Business errors

### Repository Layer

**Responsibilities:**
- Data persistence
- Query execution
- Connection management
- Optimistic locking

**Components:**
- `internal/repository/interfaces.go` - Repository contracts
- `internal/repository/mysql/` - MySQL implementations

**Optimization Techniques:**
- Connection pooling
- Prepared statements
- Batch operations
- Row-level locking (`FOR UPDATE`)

### Infrastructure Layer

#### Cache (Redis)

**Strategy**: Write-through with TTL

```
Read Request:
1. Check cache
2. If miss, query database
3. Update cache
4. Return data

Write Request:
1. Update database
2. Invalidate/update cache
3. Return success
```

**Cache Keys:**
- `voucher:{id}` - Individual vouchers
- `wallet:{user_id}` - User wallets
- `vouchers:search:{params}` - Search results

#### Message Queue (RabbitMQ)

**Pattern**: Event-driven with retry and DLQ

```
Producer:
1. Publish event to exchange
2. Event routed to queue
3. Consumer processes

Consumer:
1. Receive message
2. Process (with retries)
3. Ack on success
4. Move to DLQ on failure
```

**Events:**
- `VoucherPurchased` - After successful purchase
- `WalletCredited` - After money added
- `TransactionCompleted` - After any transaction

#### Observability

**Tracing**: OpenTelemetry with Jaeger

- Trace IDs propagate through entire request
- Spans track individual operations
- Context carries trace information

**Logging**: Structured JSON logs with Zap

- Request ID in all logs
- Log levels: DEBUG, INFO, WARN, ERROR
- Contextual information attached

**Metrics**: Prometheus format

- Request rates
- Latency histograms
- Error rates
- Resource utilization

## Data Flow

### Purchase Voucher Flow

```
Client Request
    │
    ▼
gRPC/REST Handler
    │
    ├─► Validate Input
    ├─► Start Trace Span
    └─► Call Service
         │
         ▼
    Voucher Service
         │
         ├─► Begin Transaction
         ├─► Check Voucher (with cache)
         ├─► Verify Stock
         ├─► Deduct Wallet Balance
         ├─► Create Transaction Record
         ├─► Generate Voucher Codes
         ├─► Update Stock
         ├─► Commit Transaction
         ├─► Invalidate Cache
         └─► Publish Event to MQ
              │
              ▼
         Return Response
```

### Cache Strategy

```
Read Path:
    │
    ├─► Check Redis
    │   │
    │   ├─► Hit: Return cached data
    │   │
    │   └─► Miss:
    │       ├─► Query MySQL
    │       ├─► Update Redis (TTL)
    │       └─► Return data

Write Path:
    │
    ├─► Update MySQL (in transaction)
    ├─► Invalidate Redis keys
    └─► Return success
```

### Concurrency Model

```
HTTP/gRPC Request
    │
    ├─► Goroutine per request
    │   │
    │   ├─► Service call
    │   │   │
    │   │   ├─► Database query (connection pool)
    │   │   ├─► Cache operation (connection pool)
    │   │   └─► Queue publish (channel)
    │   │
    │   └─► Response
    │
Worker Pool (background tasks):
    │
    ├─► 10 workers (configurable)
    ├─► Task queue (buffered channel)
    └─► Process async operations
```

## Design Decisions

### Why gRPC + REST?

**gRPC:**
- High performance (Protocol Buffers)
- Strong typing
- Built-in streaming
- Language-agnostic

**REST (via grpc-gateway):**
- Browser compatibility
- Easier debugging
- Wider adoption
- Simple integration

### Why MySQL over NoSQL?

**Reasons:**
- ACID guarantees required
- Complex transactions (wallet + voucher)
- Structured data with relationships
- Strong consistency needs

**Trade-offs:**
- Harder to scale horizontally
- Schema changes require migrations

### Why Redis for Caching?

**Reasons:**
- In-memory performance
- Rich data structures
- TTL support
- Pub/Sub capabilities
- High availability (Redis Sentinel/Cluster)

### Why RabbitMQ over Kafka?

**Reasons:**
- Simpler to setup and maintain
- Built-in retry mechanisms
- Dead letter queues
- Good for task queues
- Lower operational complexity

**When to use Kafka:**
- Event sourcing
- Log aggregation
- High-throughput streaming

### Transaction Isolation

**Level**: READ COMMITTED

**Reasons:**
- Balance between consistency and performance
- Prevents dirty reads
- Allows concurrent reads
- Good for most use cases

**Protection mechanisms:**
- Row-level locking (`FOR UPDATE`)
- Optimistic locking (version checks)
- Atomic operations (UPDATE ... WHERE)

## Scalability

### Horizontal Scaling

**Stateless Design:**
- No session state in service
- All state in database/cache
- Can add replicas freely

**Load Balancing:**
- gRPC: Round-robin/Least-request
- HTTP: Any load balancer

**Considerations:**
- Database connection pool per instance
- Cache hit rate per instance
- Queue consumers coordination

### Vertical Scaling

**Resource Optimization:**
- Connection pooling
- Worker pool sizing
- Memory profiling
- CPU profiling

**Limits:**
- Single machine capacity
- Diminishing returns

### Database Scaling

**Read Replicas:**
- Route queries to replicas
- Write to primary
- Handle replication lag

**Sharding:**
- Partition by user_id
- Hash-based distribution
- Cross-shard queries complexity

### Cache Scaling

**Redis Cluster:**
- Multiple master nodes
- Automatic sharding
- High availability

**Consistent Hashing:**
- Minimize key redistribution
- Add/remove nodes gracefully

## Security Considerations

### Authentication & Authorization

- JWT tokens (to be implemented)
- Role-based access control
- API key management

### Data Protection

- Encrypted connections (TLS)
- Sensitive data encryption
- PII handling compliance

### Input Validation

- Request validation at API layer
- SQL injection prevention (prepared statements)
- XSS prevention
- Rate limiting

## Future Enhancements

1. **Event Sourcing** - Complete audit trail
2. **Saga Pattern** - Distributed transactions
3. **GraphQL API** - Flexible queries
4. **Multi-tenancy** - Isolate customer data
5. **Machine Learning** - Fraud detection
6. **Blockchain** - Voucher authenticity

## References

- [Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/tags/domain%20driven%20design.html)
- [Microservices Patterns](https://microservices.io/patterns/)
- [The Twelve-Factor App](https://12factor.net/)

