package grpc

import (
	"github.com/agris/user-service/internal/config"
	"github.com/agris/user-service/internal/grpc/interceptor"
	"github.com/agris/user-service/internal/grpc/pb/userservicepb"
	"github.com/agris/user-service/internal/grpc/service_grpc"
	"google.golang.org/grpc"
	"log"
	"net"
)

func NewGRPCServer(
	config *config.Config,
	authGRPCService *service_grpc.AuthGRPCService,
	authInterceptor *interceptor.AuthInterceptor,
) (*grpc.Server, func(), error) {

	// Tạo listener
	lis, err := net.Listen("tcp", ":"+config.Server.GRPCPort)
	if err != nil {
		return nil, nil, err
	}

	// Tạo gRPC server với interceptors
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingInterceptor,
			authInterceptor.Handler(),
		),
	)

	// Đăng ký service
	userservicepb.RegisterUserServiceServer(server, authGRPCService)

	log.Printf("gRPC server configured to listen at %v", lis.Addr())

	// Cleanup function
	cleanup := func() {
		log.Println("Shutting down gRPC server...")
		server.GracefulStop()
		lis.Close()
	}

	// Start server trong goroutine
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return server, cleanup, nil
}
