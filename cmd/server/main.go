package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	voucherv1 "github.com/voucher-payment-service/api/generated"
	"github.com/voucher-payment-service/internal/cache"
	grpcserver "github.com/voucher-payment-service/internal/grpc"
	"github.com/voucher-payment-service/internal/messaging"
	"github.com/voucher-payment-service/internal/observability"
	"github.com/voucher-payment-service/internal/repository/mysql"
	"github.com/voucher-payment-service/internal/service"
	"github.com/voucher-payment-service/internal/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize logger
	logger, err := observability.InitLogger(observability.LoggerConfig{
		Level:  "info",
		Format: "console",
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting voucher payment service")

	// Initialize tracing
	shutdownTracing, err := observability.InitTracing(observability.TracingConfig{
		JaegerEndpoint: "http://localhost:14268/api/traces",
		ServiceName:    "voucher-payment-service",
		Environment:    "development",
		Enabled:        true,
	})
	if err != nil {
		logger.Fatal("failed to initialize tracing", zap.Error(err))
	}
	defer shutdownTracing(context.Background())

	// Initialize database
	db, err := mysql.NewDB(mysql.Config{
		Host:               "localhost",
		Port:               3306,
		User:               "voucheruser",
		Password:           "voucherpass",
		Database:           "voucher_db",
		MaxConnections:     25,
		MaxIdleConnections: 5,
		ConnectionLifetime: 5 * time.Minute,
	})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("database connected successfully")

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cache.Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
		CacheTTL: 5 * time.Minute,
	})
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer redisCache.Close()
	logger.Info("Redis connected successfully")

	// Initialize RabbitMQ
	mq, err := messaging.NewRabbitMQ(messaging.Config{
		Host:          "localhost",
		Port:          5672,
		User:          "guest",
		Password:      "guest",
		Exchange:      "voucher.events",
		Queue:         "voucher.purchases",
		DLQ:           "voucher.purchases.dlq",
		PrefetchCount: 10,
	})
	if err != nil {
		logger.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}
	defer mq.Close()
	logger.Info("RabbitMQ connected successfully")

	// Initialize repositories
	uow := mysql.NewUnitOfWork(db)

	// Initialize services
	voucherService := service.NewVoucherService(uow, redisCache, mq, logger)
	walletService := service.NewWalletService(uow, redisCache, logger)

	// Initialize worker pool
	workerPool := worker.NewPool(10, 100, logger)
	workerPool.Start()
	defer workerPool.Stop()
	logger.Info("worker pool started")

	// Start message queue consumer in background
	go func() {
		ctx := context.Background()
		logger.Info("starting message queue consumer")
		
		err := mq.ConsumeWithRetry(ctx, func(ctx context.Context, body []byte) error {
			logger.Info("received message", zap.ByteString("body", body))
			// Process message here
			return nil
		}, 3)
		
		if err != nil {
			logger.Error("message queue consumer error", zap.Error(err))
		}
	}()

	// Initialize gRPC handlers
	voucherHandler := grpcserver.NewVoucherServer(voucherService, logger)
	walletHandler := grpcserver.NewWalletServer(walletService, logger)

	// Create gRPC server
	grpcSrv := grpcserver.NewServer(
		grpcserver.Config{
			Port:          50051,
			RequestTimeout: 30 * time.Second,
		},
		voucherHandler,
		walletHandler,
		logger,
	)

	// Start gRPC server in background
	go func() {
		if err := grpcSrv.Start(); err != nil {
			logger.Fatal("failed to start gRPC server", zap.Error(err))
		}
	}()

	// Setup gRPC gateway (REST API)
	go func() {
		if err := startGateway(logger); err != nil {
			logger.Fatal("failed to start gateway", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := grpcSrv.Stop(ctx); err != nil {
		logger.Error("gRPC server shutdown error", zap.Error(err))
	}

	logger.Info("server stopped")
}

func startGateway(logger *zap.Logger) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create gRPC gateway mux
	mux := runtime.NewServeMux()

	// Setup connection to gRPC server
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	
	err := voucherv1.RegisterVoucherServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		return fmt.Errorf("failed to register voucher service: %w", err)
	}

	err = voucherv1.RegisterWalletServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		return fmt.Errorf("failed to register wallet service: %w", err)
	}

	// Start HTTP server
	logger.Info("starting HTTP gateway on :8081")
	
	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	return server.ListenAndServe()
}

