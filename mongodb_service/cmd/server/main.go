package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RevenueIQ/mongo_service/internal/db"
	"github.com/RevenueIQ/mongo_service/internal/grpc_clients"
	"github.com/RevenueIQ/mongo_service/internal/grpc_servers"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Configuration
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGO_DB", "mongodb_service_db")
	colName := getEnv("MONGO_COLLECTION", "users")
	port := getEnv("PORT", "50051")
	redisCacheAddr := getEnv("REDIS_CACHE_ADDR", "localhost:50052")

	log.Printf("Starting mongodb_service server...")
	log.Printf("Configured MongoDB URI: %s", mongoURI)
	log.Printf("Configured Database Name: %s", dbName)
	log.Printf("Configured Collection Name: %s", colName)
	log.Printf("Configured Redis Cache Address: %s", redisCacheAddr)
	log.Printf("Configured gRPC Port: %s", port)

	// Context for database connection startup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	mongoDB, err := db.ConnectMongo(ctx, mongoURI, dbName, colName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoDB.Client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()
	log.Printf("Successfully connected to MongoDB.")

	// Connect to Redis Cache gRPC Service using the client helper (optional fallback)
	var redisClient redis_pb.RedisCacheServiceClient
	redisClient, redisConn, err := grpcclients.NewRedisCacheClient(redisCacheAddr)
	if err != nil {
		log.Printf("[WARNING] Redis cache service unavailable (%v). Operating without optional cache layer.", err)
		redisClient = nil
	} else {
		defer redisConn.Close()
		log.Printf("Successfully connected to Redis cache client.")
	}

	// Start gRPC Server using the server helper
	grpcServer, lis, err := grpcservers.StartGRPCServer(port, mongoDB, redisClient)
	if err != nil {
		log.Fatalf("Failed to start gRPC Server: %v", err)
	}

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
