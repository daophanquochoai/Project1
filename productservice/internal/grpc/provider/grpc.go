package provider

import (
	"log"
	"productservice/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ProvideGRPCConnection(cfg *config.Config) (*grpc.ClientConn, func(), error) {
	// Lấy địa chỉ từ config
	authServiceAddr := cfg.Server.Grpc.Auth.Host + ":" + cfg.Server.Grpc.Auth.Port
	if authServiceAddr == "" {
		authServiceAddr = "localhost:9005"
	}

	log.Printf("Connecting to Product Service at: %s", authServiceAddr)

	conn, err := grpc.NewClient(
		authServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Failed to connect to Product Service: %v", err)
		return nil, nil, err
	}

	log.Printf("Successfully connected to Product Service")

	// Cleanup function
	cleanup := func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		} else {
			log.Printf("gRPC connection closed")
		}
	}

	return conn, cleanup, nil
}
