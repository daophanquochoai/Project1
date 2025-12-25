package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agris/user-service/internal"
	"github.com/agris/user-service/internal/grpc"
)

func main() {
	port := flag.Int("port", 8005, "Server port")
	flag.Parse()
	app, err := internal.New()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	go func() {
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
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 5. Start HTTP server trong goroutine
	go func() {
		addr := fmt.Sprintf(":%d", *port)
		log.Printf("HTTP server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("All servers exited gracefully")
}
