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

func StartGRPCServer(config *config.Config, userGRPCService *service_grpc.AuthGRPCService) {
	lis, err := net.Listen("tcp", ":"+config.Server.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingInterceptor,
		),
	)

	userservicepb.RegisterUserServiceServer(server, userGRPCService)

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
