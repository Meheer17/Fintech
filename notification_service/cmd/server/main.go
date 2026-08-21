package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
	"github.com/RevenueIQ/notification_service/internal/redpanda"
	"github.com/RevenueIQ/notification_service/internal/ses"
	"github.com/RevenueIQ/notification_service/internal/worker"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Configuration
	redpandaAddr := getEnv("REDPANDA_ADDR", "localhost:9092")
	redisAddr := getEnv("REDIS_SERVICE_ADDR", "localhost:50052")
	awsSESMock := getEnv("AWS_SES_MOCK", "false")

	log.Printf("Starting notification_service...")
	log.Printf("Configured Redpanda Address: %s", redpandaAddr)
	log.Printf("Configured Redis Service Address: %s", redisAddr)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Redis Service
	redisConn, err := grpc.Dial(redisAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial redis_service: %v", err)
	}
	defer redisConn.Close()
	redisClient := redis_pb.NewRedisCacheServiceClient(redisConn)
	log.Printf("Connected to redis_service Client.")

	// Initialize SES Client
	var sesClient ses.SESClient
	if awsSESMock == "true" || (os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_SECRET_ACCESS_KEY") == "") {
		log.Printf("Initializing Mock AWS SES Client (AWS_SES_MOCK=%s or no AWS credentials found)", awsSESMock)
		sesClient = ses.NewMockSESClient()
	} else {
		log.Printf("Initializing Real AWS SES Client...")
		cfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("Failed to load AWS config: %v", err)
		}
		sesClient = ses.NewRealSESClient(cfg)
	}

	// Initialize Redpanda Reader
	// Topic: EMAIL, Group ID: notification-service-group
	reader := redpanda.NewRedpandaReader(redpandaAddr, "EMAIL", "notification-service-group")
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing Redpanda reader: %v", err)
		}
	}()

	// Initialize and start worker
	w := worker.NewWorker(reader, redisClient, sesClient)

	// Start a simple HTTP health check server
	healthPort := getEnv("HEALTH_PORT", "8083")
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Printf("Starting health check server on port %s...", healthPort)
		if err := http.ListenAndServe(":"+healthPort, mux); err != nil {
			log.Printf("Health check server failed: %v", err)
		}
	}()

	// Channel to listen for OS signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Run worker in a goroutine
	go w.Start(ctx)

	// Block until signal is received
	<-stopChan
	log.Printf("Shutting down worker and reader gracefully...")
	cancel()

	// Wait briefly for worker to clean up
	time.Sleep(1 * time.Second)
	log.Printf("Notification service stopped.")
}
