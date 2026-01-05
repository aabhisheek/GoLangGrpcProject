# Voucher & Payment Service - Production-Grade Capstone Project

A production-ready microservice built with Go that handles voucher management, wallet operations, and payments. This project demonstrates advanced Go concepts including concurrency, gRPC, clean architecture, observability, and distributed systems.

## 🎯 Project Overview

This is the capstone project for an Advanced Golang Training Track 2, implementing a real-world backend system similar to those used in fintech and e-commerce platforms.

### Key Features

- ✅ **gRPC & REST API** - Dual protocol support with grpc-gateway
- ✅ **Clean Architecture** - Modular, testable, and maintainable code
- ✅ **Database** - MySQL with ACID transactions
- ✅ **Caching** - Redis for high-performance reads
- ✅ **Message Queue** - RabbitMQ for async event processing
- ✅ **Observability** - OpenTelemetry tracing, structured logging
- ✅ **Concurrency** - Worker pools for parallel processing
- ✅ **Testing** - Unit tests, integration tests, and benchmarks
- ✅ **Production-Ready** - Interceptors, health checks, graceful shutdown

## 📋 Table of Contents

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Development](#development)
- [Testing](#testing)
- [Performance](#performance)
- [Deployment](#deployment)
- [Observability](#observability)

## 🏗️ Architecture

### System Components

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ├─────► gRPC (Port 50051)
       │
       └─────► REST/HTTP (Port 8080)
              │
              ▼
       ┌──────────────┐
       │  API Layer   │
       │ (Handlers)   │
       └──────┬───────┘
              │
              ▼
       ┌──────────────┐
       │Service Layer │
       │(Business     │
       │ Logic)       │
       └──────┬───────┘
              │
       ┌──────┴───────┐
       │              │
       ▼              ▼
┌──────────┐   ┌──────────┐
│Repository│   │  Cache   │
│  Layer   │   │  (Redis) │
└────┬─────┘   └──────────┘
     │
     ▼
┌──────────┐   ┌───────────┐   ┌─────────────┐
│  MySQL   │   │ RabbitMQ  │   │ Jaeger      │
│ Database │   │   Queue   │   │ (Tracing)   │
└──────────┘   └───────────┘   └─────────────┘
```

### Clean Architecture Layers

1. **Domain Layer** - Business entities and rules
2. **Repository Layer** - Data access abstraction
3. **Service Layer** - Business logic orchestration
4. **API Layer** - gRPC/REST handlers
5. **Infrastructure** - External services (cache, queue, observability)

## 📦 Prerequisites

- **Go** 1.21 or higher
- **Docker** & Docker Compose
- **Protocol Buffers** compiler (protoc)
- **MySQL** 8.0+
- **Redis** 7.0+
- **RabbitMQ** 3.12+

### Install Required Tools

```bash
# Install protoc plugins
make install-tools

# Or manually:
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

## 🚀 Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd Go-lang-Project

# Install dependencies
make deps
```

### 2. Start Infrastructure Services

```bash
# Start MySQL, Redis, RabbitMQ, Jaeger
make docker-up

# Wait for services to be healthy (30 seconds)
```

### 3. Run Database Migrations

```bash
make migrate-up
```

### 4. Generate Protocol Buffers

```bash
make proto
```

### 5. Run the Service

```bash
make run

# Or build and run
make build
./bin/voucher-payment-service
```

The service will start on:
- **gRPC**: `localhost:50051`
- **HTTP/REST**: `localhost:8080`
- **Jaeger UI**: `http://localhost:16686`
- **RabbitMQ Management**: `http://localhost:15672`

## 📚 API Documentation

### gRPC Services

#### VoucherService

- `SearchVouchers` - Search for available vouchers
- `GetVoucher` - Get voucher details by ID
- `BuyVoucher` - Purchase vouchers
- `ListUserVouchers` - List user's purchased vouchers

#### WalletService

- `GetBalance` - Get wallet balance
- `AddMoney` - Add money to wallet
- `ListTransactions` - List wallet transactions

### REST API Endpoints

All gRPC methods are also exposed via REST:

#### Vouchers

```bash
# Search vouchers
GET /v1/vouchers/search?query=amazon&category=E-Commerce&page=1&page_size=10

# Get voucher
GET /v1/vouchers/{voucher_id}

# Buy voucher
POST /v1/vouchers/buy
{
  "user_id": "user-001",
  "voucher_id": "voucher-001",
  "quantity": 2,
  "payment_method": "wallet"
}

# List user vouchers
GET /v1/users/{user_id}/vouchers?status=active&page=1&page_size=10
```

#### Wallet

```bash
# Get balance
GET /v1/users/{user_id}/balance

# Add money
POST /v1/users/{user_id}/wallet/add
{
  "amount": 1000.0,
  "payment_method": "upi",
  "transaction_reference": "UPI123456"
}

# List transactions
GET /v1/users/{user_id}/transactions?type=all&page=1&page_size=10
```

### Example: Buy Voucher with cURL

```bash
curl -X POST http://localhost:8080/v1/vouchers/buy \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "voucher_id": "voucher-001",
    "quantity": 1,
    "payment_method": "wallet"
  }'
```

## 📁 Project Structure

```
.
├── api/
│   ├── proto/              # Protocol Buffer definitions
│   └── generated/          # Generated gRPC code
├── cmd/
│   ├── server/             # Main application entry point
│   └── migrate/            # Database migration tool
├── config/                 # Configuration files
├── internal/
│   ├── domain/             # Domain models and errors
│   ├── repository/         # Data access layer
│   │   └── mysql/          # MySQL implementations
│   ├── service/            # Business logic
│   ├── grpc/               # gRPC handlers and interceptors
│   ├── cache/              # Redis caching
│   ├── messaging/          # RabbitMQ integration
│   ├── observability/      # Tracing and logging
│   └── worker/             # Worker pool implementation
├── migrations/             # SQL migration files
├── docker-compose.yml      # Infrastructure services
├── Makefile               # Build and development tasks
└── README.md              # This file
```

## 🔧 Development

### Available Make Commands

```bash
make help              # Show all available commands
make proto             # Generate protobuf files
make build             # Build the application
make run               # Run the application
make test              # Run tests with coverage
make benchmark         # Run benchmarks
make docker-up         # Start docker services
make docker-down       # Stop docker services
make migrate-up        # Run database migrations
make migrate-down      # Rollback migrations
make clean             # Clean generated files
```

### Development Workflow

1. **Make code changes**
2. **Generate protobuf** (if proto files changed): `make proto`
3. **Run tests**: `make test`
4. **Run locally**: `make run`
5. **Check observability**: Visit Jaeger UI at `http://localhost:16686`

## 🧪 Testing

### Run All Tests

```bash
make test
```

This generates a coverage report at `coverage.html`.

### Run Benchmarks

```bash
make benchmark
```

This generates profiling data:
- `cpu.prof` - CPU profile
- `mem.prof` - Memory profile

### View Profiles

```bash
# CPU profile
go tool pprof cpu.prof

# Memory profile
go tool pprof mem.prof

# Web interface
go tool pprof -http=:8081 cpu.prof
```

### Integration Tests

Integration tests require running infrastructure:

```bash
make docker-up
# Wait 30 seconds
go test -tags=integration ./...
```

## ⚡ Performance

### Optimization Features

- **Connection Pooling** - MySQL and Redis connection pools
- **Caching Strategy** - Multi-level caching with TTL
- **Worker Pools** - Concurrent task processing
- **Batch Operations** - Bulk inserts for efficiency
- **Prepared Statements** - SQL query optimization
- **Index Optimization** - Database indexes on hot paths

### Benchmarks

Run benchmarks to measure performance:

```bash
make benchmark
```

Expected throughput (on standard hardware):
- Voucher search: ~10,000 req/sec
- Wallet operations: ~5,000 req/sec
- Purchase flow: ~2,000 req/sec

## 🚢 Deployment

### Docker Build

```bash
# Build Docker image
docker build -t voucher-payment-service:latest .

# Run with Docker
docker run -p 50051:50051 -p 8080:8080 voucher-payment-service:latest
```

### Environment Variables

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=voucheruser
export DB_PASSWORD=voucherpass
export DB_NAME=voucher_db
export REDIS_HOST=localhost
export REDIS_PORT=6379
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

### Health Checks

```bash
# gRPC health check
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# HTTP health check
curl http://localhost:8080/health
```

## 📊 Observability

### Distributed Tracing

View traces in Jaeger UI:
```
http://localhost:16686
```

Traces show:
- Request flow through system
- Database query times
- Cache hit/miss rates
- RabbitMQ message processing

### Logging

Structured logs with:
- Request IDs
- User IDs
- Operation types
- Error details
- Performance metrics

### Metrics

Key metrics tracked:
- Request latency (p50, p95, p99)
- Throughput (requests/second)
- Error rates
- Cache hit ratio
- Database connection pool stats

## 🎓 Learning Outcomes

This project demonstrates:

1. **Go Internals** - Memory management, escape analysis
2. **Concurrency** - Goroutines, channels, worker pools
3. **Context** - Cancellation, timeouts, metadata
4. **Interfaces & Generics** - Clean abstractions
5. **Error Handling** - Custom errors, wrapping
6. **gRPC** - Unary calls, interceptors, metadata
7. **Testing** - Unit tests, benchmarks, mocks
8. **Profiling** - CPU and memory profiling
9. **Caching** - Redis integration, cache strategies
10. **Message Queues** - RabbitMQ, event-driven architecture
11. **Databases** - Transactions, connection pooling, migrations
12. **Observability** - Tracing, logging, metrics

## 📖 References

- [Go Documentation](https://go.dev/doc/)
- [gRPC Go Quickstart](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## 🤝 Contributing

This is a training capstone project. For improvements:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📝 License

This project is for educational purposes.

## 👥 Authors

Created as part of Advanced Golang Training Track 2.

---

**Happy Coding! 🚀**

For questions or issues, please refer to the course materials or contact your instructor.

