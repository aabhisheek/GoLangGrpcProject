#!/bin/bash

# Test script for Voucher & Payment Service

set -e

echo "🧪 Running tests for Voucher & Payment Service..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_message() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Run unit tests
print_message $YELLOW "\n📝 Running unit tests..."
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
print_message $YELLOW "\n📊 Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Show coverage summary
print_message $YELLOW "\n📈 Coverage summary:"
go tool cover -func=coverage.out | grep total

# Run benchmarks
print_message $YELLOW "\n⚡ Running benchmarks..."
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...

print_message $GREEN "\n✅ All tests completed!"
print_message $YELLOW "\nGenerated files:"
echo "  - coverage.out      (coverage data)"
echo "  - coverage.html     (coverage report)"
echo "  - cpu.prof          (CPU profile)"
echo "  - mem.prof          (memory profile)"
echo ""
print_message $YELLOW "View coverage report:"
echo "  open coverage.html"
echo ""
print_message $YELLOW "Analyze profiles:"
echo "  go tool pprof cpu.prof"
echo "  go tool pprof mem.prof"
echo "  go tool pprof -http=:8081 cpu.prof"

