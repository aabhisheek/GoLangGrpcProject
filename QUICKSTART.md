# Quick Start Guide

Get the Voucher & Payment Service running in 5 minutes!

## Prerequisites

- **Go** 1.21+ installed ([download](https://go.dev/dl/))
- **Docker Desktop** running ([download](https://www.docker.com/products/docker-desktop))
- **Git** installed

## Step 1: Clone the Repository

```bash
git clone <repository-url>
cd Go-lang-Project
```

## Step 2: Automated Setup

### For Linux/macOS:

```bash
./scripts/setup.sh
```

### For Windows:

```cmd
scripts\setup.bat
```

### Or Manual Setup:

```bash
# Install dependencies
go mod download

# Start infrastructure (MySQL, Redis, RabbitMQ, Jaeger)
docker-compose up -d

# Wait 30 seconds for services to start
# Then run migrations
go run cmd/migrate/main.go up

# Generate protobuf files
make proto
# Or manually:
# protoc -I api/proto --go_out=api/generated --go_opt=paths=source_relative \
#   --go-grpc_out=api/generated --go-grpc_opt=paths=source_relative \
#   --grpc-gateway_out=api/generated --grpc-gateway_opt=paths=source_relative \
#   api/proto/*.proto
```

## Step 3: Run the Service

```bash
# Using Make
make run

# Or directly
go run cmd/server/main.go
```

You should see:
```
INFO    starting voucher payment service
INFO    database connected successfully
INFO    Redis connected successfully
INFO    RabbitMQ connected successfully
INFO    starting gRPC server    {"port": 50051}
INFO    starting HTTP gateway on :8080
```

## Step 4: Test the API

### Option 1: Run the Demo Script

```bash
# Linux/macOS
./scripts/demo.sh

# Windows - use curl or Postman with the commands below
```

### Option 2: Manual Testing

**1. Search for vouchers:**
```bash
curl "http://localhost:8080/v1/vouchers/search?query=amazon&page=1&page_size=5"
```

**2. Get voucher details:**
```bash
curl "http://localhost:8080/v1/vouchers/voucher-001"
```

**3. Check wallet balance:**
```bash
curl "http://localhost:8080/v1/users/user-001/balance"
```

**4. Buy a voucher:**
```bash
curl -X POST "http://localhost:8080/v1/vouchers/buy" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "voucher_id": "voucher-001",
    "quantity": 1,
    "payment_method": "wallet"
  }'
```

## Step 5: Explore the System

### Jaeger (Distributed Tracing)
Open: http://localhost:16686

- Search for traces
- View request flows
- Analyze performance

### RabbitMQ Management
Open: http://localhost:15672
- Username: `guest`
- Password: `guest`

View queues, exchanges, and messages.

### API Documentation
Check `API_GUIDE.md` for detailed API documentation.

## Common Commands

```bash
# Start services
make docker-up

# Stop services
make docker-down

# Run tests
make test

# View test coverage
open coverage.html

# Run benchmarks
make benchmark

# Clean generated files
make clean

# View all commands
make help
```

## Sample Data

The database comes pre-loaded with:
- **3 users** (user-001, user-002, user-003)
- **10 vouchers** (Amazon, Flipkart, Starbucks, etc.)
- **Wallets** with initial balance

## Project Structure Overview

```
.
├── cmd/server/         # Main application
├── api/proto/          # gRPC definitions
├── internal/
│   ├── domain/         # Business entities
│   ├── service/        # Business logic
│   ├── repository/     # Data access
│   └── grpc/           # API handlers
├── migrations/         # Database schema
└── docker-compose.yml  # Infrastructure
```

## Next Steps

1. **Explore the Code**: Start with `cmd/server/main.go`
2. **Read Architecture**: Check `ARCHITECTURE.md`
3. **Try Examples**: See `API_GUIDE.md`
4. **Run Tests**: Execute `make test`
5. **View Traces**: Explore in Jaeger UI
6. **Modify**: Add a new feature!

## Troubleshooting

### Service won't start

**Check if ports are available:**
```bash
# Check port 50051 (gRPC)
lsof -i :50051

# Check port 8080 (HTTP)
lsof -i :8080

# Windows
netstat -ano | findstr :50051
netstat -ano | findstr :8080
```

**Check Docker containers:**
```bash
docker-compose ps
```

All containers should show "Up (healthy)".

### Database connection error

**Restart MySQL:**
```bash
docker-compose restart mysql

# Wait for healthy status
docker-compose ps
```

### Proto generation fails

**Install protoc plugins:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

**Check protoc installation:**
```bash
protoc --version
```

### "Module not found" errors

```bash
go mod download
go mod tidy
```

## Performance Tips

- **First run** might be slow (downloading deps, building)
- **Subsequent runs** are faster (caching)
- **Docker** needs 4GB+ RAM allocated
- **Enable caching** for best performance

## Learning Path

1. ✅ **Quick Start** (you are here!)
2. 📖 **README.md** - Full documentation
3. 🏗️ **ARCHITECTURE.md** - System design
4. 📚 **API_GUIDE.md** - API reference
5. 🚀 **DEPLOYMENT.md** - Production deployment
6. 🤝 **CONTRIBUTING.md** - How to contribute

## Support

- **Documentation**: Check `README.md`
- **Issues**: Search closed issues on GitHub
- **Discussions**: Use GitHub Discussions
- **Contact**: Reach out to maintainers

## What You've Built

Congratulations! You now have a running production-grade microservice with:

✅ gRPC & REST APIs
✅ Database (MySQL)
✅ Caching (Redis)
✅ Message Queue (RabbitMQ)
✅ Distributed Tracing (Jaeger)
✅ Clean Architecture
✅ Comprehensive Testing

**Happy Coding! 🚀**

