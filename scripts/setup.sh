#!/bin/bash

# Setup script for Voucher & Payment Service
# This script helps set up the development environment

set -e

echo "🚀 Setting up Voucher & Payment Service..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Print colored message
print_message() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Check prerequisites
print_message $YELLOW "📋 Checking prerequisites..."

if ! command_exists go; then
    print_message $RED "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi
print_message $GREEN "✅ Go is installed: $(go version)"

if ! command_exists docker; then
    print_message $RED "❌ Docker is not installed. Please install Docker."
    exit 1
fi
print_message $GREEN "✅ Docker is installed: $(docker --version)"

if ! command_exists docker-compose; then
    print_message $RED "❌ Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi
print_message $GREEN "✅ Docker Compose is installed: $(docker-compose --version)"

if ! command_exists protoc; then
    print_message $YELLOW "⚠️  protoc is not installed. Attempting to install..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install protobuf
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo apt-get update && sudo apt-get install -y protobuf-compiler
    else
        print_message $RED "❌ Please install protoc manually: https://grpc.io/docs/protoc-installation/"
        exit 1
    fi
fi
print_message $GREEN "✅ protoc is installed: $(protoc --version)"

# Install Go tools
print_message $YELLOW "\n📦 Installing Go tools..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
print_message $GREEN "✅ Go tools installed"

# Download dependencies
print_message $YELLOW "\n📥 Downloading Go dependencies..."
go mod download
go mod tidy
print_message $GREEN "✅ Dependencies downloaded"

# Start infrastructure services
print_message $YELLOW "\n🐳 Starting infrastructure services (MySQL, Redis, RabbitMQ, Jaeger)..."
docker-compose up -d

print_message $YELLOW "⏳ Waiting for services to be ready (30 seconds)..."
sleep 30

# Check if services are healthy
print_message $YELLOW "\n🔍 Checking service health..."

if docker-compose ps | grep -q "Up (healthy).*mysql"; then
    print_message $GREEN "✅ MySQL is healthy"
else
    print_message $RED "❌ MySQL is not healthy"
fi

if docker-compose ps | grep -q "Up (healthy).*redis"; then
    print_message $GREEN "✅ Redis is healthy"
else
    print_message $RED "❌ Redis is not healthy"
fi

if docker-compose ps | grep -q "Up (healthy).*rabbitmq"; then
    print_message $GREEN "✅ RabbitMQ is healthy"
else
    print_message $RED "❌ RabbitMQ is not healthy"
fi

# Run migrations
print_message $YELLOW "\n🗃️  Running database migrations..."
go run cmd/migrate/main.go up
print_message $GREEN "✅ Migrations completed"

# Generate protobuf files
print_message $YELLOW "\n🔧 Generating protobuf files..."
make proto
print_message $GREEN "✅ Protobuf files generated"

# Build the application
print_message $YELLOW "\n🔨 Building the application..."
make build
print_message $GREEN "✅ Application built successfully"

# Success message
print_message $GREEN "\n🎉 Setup completed successfully!"
print_message $YELLOW "\n📚 Quick Start:"
echo "  1. Run the service:    make run"
echo "  2. View Jaeger UI:     http://localhost:16686"
echo "  3. View RabbitMQ UI:   http://localhost:15672 (guest/guest)"
echo "  4. gRPC endpoint:      localhost:50051"
echo "  5. REST endpoint:      http://localhost:8080"
echo ""
print_message $YELLOW "📖 For more information, see README.md"
echo ""
print_message $GREEN "Happy coding! 🚀"

