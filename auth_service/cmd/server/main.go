package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	auth_pb "github.com/RevenueIQ/auth_service/auth_service"
	"github.com/RevenueIQ/auth_service/internal/handler"
	"github.com/RevenueIQ/auth_service/internal/middleware"
	"github.com/RevenueIQ/auth_service/internal/server"
	mongo_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	queue_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/queue_service"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"

	pb "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Configuration
	mongoAddr := getEnv("MONGO_SERVICE_ADDR", "localhost:50051")
	redisAddr := getEnv("REDIS_SERVICE_ADDR", "localhost:50052")
	queueAddr := getEnv("QUEUE_SERVICE_ADDR", "localhost:50054")
	port := getEnv("PORT", "50053")
	httpPort := getEnv("HTTP_PORT", "8081")

	log.Printf("Starting auth_service server...")
	log.Printf("Configured MongoDB Service Address: %s", mongoAddr)
	log.Printf("Configured Redis Service Address: %s", redisAddr)
	log.Printf("Configured Queue Service Address: %s", queueAddr)
	log.Printf("Configured gRPC Port: %s", port)
	log.Printf("Configured HTTP Port: %s", httpPort)

	// Dial MongoDB Service
	mongoConn, err := grpc.Dial(mongoAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial mongodb_service: %v", err)
	}
	defer mongoConn.Close()
	mongoClient := mongo_pb.NewUserServiceClient(mongoConn)
	log.Printf("Connected to mongodb_service Client.")

	// Dial Redis Service
	redisConn, err := grpc.Dial(redisAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial redis_service: %v", err)
	}
	defer redisConn.Close()
	redisClient := redis_pb.NewRedisCacheServiceClient(redisConn)
	log.Printf("Connected to redis_service Client.")

	// Dial Queue Service
	queueConn, err := grpc.Dial(queueAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial queue_service: %v", err)
	}
	defer queueConn.Close()
	queueClient := queue_pb.NewQueueServiceClient(queueConn)
	log.Printf("Connected to queue_service Client.")

	// Create TCP Listener for gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", err)
	}

	// Create and register gRPC Server
	grpcServer := grpc.NewServer()
	authServer := server.NewServer(mongoClient, redisClient, queueClient)
	auth_pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterAuthorizationServer(grpcServer, authServer)

	// Channel to listen for OS signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Run gRPC server in a goroutine
	go func() {
		log.Printf("gRPC server listening on port %s...", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server serve ended: %v", err)
		}
	}()

	// Wait a moment for gRPC server to start before dialing locally
	time.Sleep(100 * time.Millisecond)

	// Connect to local gRPC Auth Service client for Gin HTTP router
	authConn, err := grpc.Dial(fmt.Sprintf("localhost:%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to local gRPC auth_service: %v", err)
	}
	defer authConn.Close()

	authClient := auth_pb.NewAuthServiceClient(authConn)
	authHandler := handler.NewAuthHandler(authClient)

	// Create gin router
	r := gin.Default()

	// Public routes
	api := r.Group("/api/auth")
	{
		api.GET("/health", authHandler.Health)
		api.POST("/signup/user", authHandler.Signup)
		api.POST("/otp/send", authHandler.SendOTP)
		api.POST("/otp/verify", authHandler.VerifyOTP)
		api.POST("/login", authHandler.Login)
		api.POST("/refresh-tokens", authHandler.RefreshTokens)
		api.POST("/forgot-password/send", authHandler.ForgotPasswordSend)
		api.POST("/forgot-password/reset", authHandler.ForgotPasswordReset)
	}

	// Protected routes
	protected := r.Group("/api/auth", middleware.AuthMiddleware(authClient))
	{
		protected.POST("/logout", authHandler.Logout)
		protected.GET("/verify", authHandler.VerifyToken)
		protected.POST("/verify", authHandler.VerifyToken)
		protected.GET("/me", authHandler.Me)
	}

	// Run Gin HTTP server in a goroutine
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: r,
	}

	go func() {
		log.Printf("HTTP Gin server listening on port %s...", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP Gin server serve ended: %v", err)
		}
	}()

	// Block until signal is received
	<-stopChan
	log.Printf("Shutting down servers gracefully...")

	// Gracefully stop gRPC and HTTP servers
	grpcServer.GracefulStop()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP Shutdown error: %v", err)
	}

	log.Printf("Servers stopped.")
}
