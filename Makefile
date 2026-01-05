.PHONY: proto build run test docker-up docker-down clean benchmark

# Variables
PROTO_DIR=api/proto
PROTO_OUT=api/generated
SERVICE_NAME=voucher-payment-service

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@mkdir -p $(PROTO_OUT)
	protoc -I $(PROTO_DIR) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		$(PROTO_DIR)/*.proto

# Build the application
build:
	@echo "Building..."
	go build -o bin/$(SERVICE_NAME) cmd/server/main.go

# Run the application
run:
	@echo "Running..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...

# Start docker services
docker-up:
	@echo "Starting docker services..."
	docker-compose up -d

# Stop docker services
docker-down:
	@echo "Stopping docker services..."
	docker-compose down

# Clean generated files
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf $(PROTO_OUT)/*.pb.go
	rm -f coverage.out coverage.html
	rm -f *.prof *.pprof

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run migrations
migrate-up:
	@echo "Running migrations..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back migrations..."
	go run cmd/migrate/main.go down

# Install tools
install-tools:
	@echo "Installing tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Help
help:
	@echo "Available targets:"
	@echo "  proto          - Generate protobuf files"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests with coverage"
	@echo "  benchmark      - Run benchmarks"
	@echo "  docker-up      - Start docker services"
	@echo "  docker-down    - Stop docker services"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"
	@echo "  clean          - Clean generated files"
	@echo "  deps           - Download dependencies"
	@echo "  install-tools  - Install required tools"

