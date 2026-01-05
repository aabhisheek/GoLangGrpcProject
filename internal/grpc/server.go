package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	voucherv1 "github.com/voucher-payment-service/api/generated"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server wraps gRPC server
type Server struct {
	server        *grpc.Server
	voucherServer *VoucherServer
	walletServer  *WalletServer
	logger        *zap.Logger
	port          int
}

// Config holds server configuration
type Config struct {
	Port          int
	RequestTimeout time.Duration
}

// NewServer creates a new gRPC server
func NewServer(
	cfg Config,
	voucherServer *VoucherServer,
	walletServer *WalletServer,
	logger *zap.Logger,
) *Server {
	// Get tracer
	tracer := otel.Tracer("grpc-server")

	// Create interceptors
	interceptors := []grpc.UnaryServerInterceptor{
		RecoveryInterceptor(logger),
		LoggingInterceptor(logger),
		TracingInterceptor(tracer),
		AuthInterceptor(logger),
		MetricsInterceptor(),
		TimeoutInterceptor(cfg.RequestTimeout),
	}

	// Create server with chained interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...),
	)

	// Register services
	voucherv1.RegisterVoucherServiceServer(grpcServer, voucherServer)
	voucherv1.RegisterWalletServiceServer(grpcServer, walletServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for grpcurl and other tools
	reflection.Register(grpcServer)

	return &Server{
		server:        grpcServer,
		voucherServer: voucherServer,
		walletServer:  walletServer,
		logger:        logger,
		port:          cfg.Port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("starting gRPC server", zap.Int("port", s.port))

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping gRPC server")

	// Graceful stop with timeout
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	case <-stopped:
		return nil
	}
}

