package main

import (
	"github.com/agris/user-service/internal/config"
	"github.com/agris/user-service/internal/grpc"
	"log"
)

func main() {
	userGRPCService, err := grpc.InitGRPCServer()
	if err != nil {
		log.Fatalf("Failed to initialize gRPC: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Starting gRPC server")
	grpc.StartGRPCServer(cfg, userGRPCService)
}
