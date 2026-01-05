@echo off
REM Setup script for Voucher & Payment Service (Windows)

echo ====================================
echo Setting up Voucher Payment Service
echo ====================================

REM Check Go installation
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed. Please install Go 1.21+
    exit /b 1
)
echo [OK] Go is installed

REM Check Docker installation
where docker >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Docker is not installed. Please install Docker Desktop
    exit /b 1
)
echo [OK] Docker is installed

REM Install Go tools
echo.
echo Installing Go tools...
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
echo [OK] Go tools installed

REM Download dependencies
echo.
echo Downloading dependencies...
go mod download
go mod tidy
echo [OK] Dependencies downloaded

REM Start infrastructure
echo.
echo Starting infrastructure services...
docker-compose up -d

echo.
echo Waiting for services to start (30 seconds)...
timeout /t 30 /nobreak >nul

REM Run migrations
echo.
echo Running database migrations...
go run cmd/migrate/main.go up
echo [OK] Migrations completed

REM Generate protobuf files
echo.
echo Generating protobuf files...
if not exist "api\generated" mkdir api\generated
protoc -I api/proto --go_out=api/generated --go_opt=paths=source_relative --go-grpc_out=api/generated --go-grpc_opt=paths=source_relative --grpc-gateway_out=api/generated --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=generate_unbound_methods=true api/proto/*.proto
echo [OK] Protobuf files generated

REM Build application
echo.
echo Building application...
go build -o bin\voucher-payment-service.exe cmd\server\main.go
echo [OK] Application built

echo.
echo ====================================
echo Setup completed successfully!
echo ====================================
echo.
echo Quick Start:
echo   1. Run the service:    go run cmd/server/main.go
echo   2. Jaeger UI:          http://localhost:16686
echo   3. RabbitMQ UI:        http://localhost:15672 (guest/guest)
echo   4. gRPC endpoint:      localhost:50051
echo   5. REST endpoint:      http://localhost:8080
echo.
echo For more info, see README.md
echo.
pause

