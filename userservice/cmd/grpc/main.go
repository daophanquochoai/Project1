package main

import (
	"github.com/agris/user-service/internal/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Wire inject - tạo gRPC server và tất cả dependencies
	_, cleanup, err := grpc.InitGRPCServer()
	if err != nil {
		log.Fatalf("Failed to initialize gRPC server: %v", err)
	}
	defer cleanup()

	log.Println("gRPC server started successfully")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Received shutdown signal, cleaning up...")
}
