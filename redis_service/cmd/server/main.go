package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
	"github.com/RevenueIQ/redis_service/internal/db"
	"github.com/RevenueIQ/redis_service/internal/server"

	"google.golang.org/grpc"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Configuration
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDBStr := getEnv("REDIS_DB", "0")
	port := getEnv("PORT", "50052")

	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Printf("Invalid REDIS_DB config: %s. Defaulting to 0", redisDBStr)
		redisDB = 0
	}

	log.Printf("Starting redis_service server...")
	log.Printf("Configured Redis Address: %s", redisAddr)
	log.Printf("Configured Redis Database: %d", redisDB)
	log.Printf("Configured gRPC Port: %s", port)

	// Context for database connection startup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to Redis
	rdb, err := db.ConnectRedis(ctx, redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Successfully connected to Redis.")

	// Create TCP Listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", err)
	}

	// Create gRPC Server
	grpcServer := grpc.NewServer()
	redisCacheServer := server.NewServer(rdb)
	pb.RegisterRedisCacheServiceServer(grpcServer, redisCacheServer)

	// Channel to listen for OS signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Run gRPC server in a goroutine
	go func() {
		log.Printf("gRPC server listening on port %s...", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Block until signal is received
	<-stopChan
	log.Printf("Shutting down gRPC server gracefully...")

	// Gracefully stop the server
	grpcServer.GracefulStop()
	log.Printf("gRPC server stopped.")
}
