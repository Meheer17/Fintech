package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/webhook"
)

type server struct {
	pb.UnimplementedWebhookServiceServer
}

// Supported Active Razorpay Webhook Events
var SupportedWebhookEvents = map[string]bool{
	"payment.authorized":        true,
	"payment.failed":            true,
	"payment.captured":          true,
	"payment.dispute.created":   true,
	"order.paid":                 true,
	"subscription.pending":       true,
	"subscription.charged":       true,
	"subscription.charged.failed": true,
	"subscription.cancelled":     true,
	"settlement.processed":       true,
	"refund.created":            true,
}

func verifySignature(body []byte, signature, secret string) bool {
	if secret == "" {
		return true // skip verification in test mode
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

func handleWebhookRequest(c *gin.Context, webhookSecret string, kafkaWriter *kafka.Writer) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	signature := c.GetHeader("X-Razorpay-Signature")
	isValid := verifySignature(body, signature, webhookSecret)
	sigStatus := "VALID"
	if !isValid {
		sigStatus = "INVALID"
		log.Printf("[WARNING] Invalid Razorpay webhook signature received for signature: %s", signature)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	eventType, _ := payload["event"].(string)
	eventID, _ := payload["account_id"].(string)

	isKnown := SupportedWebhookEvents[eventType]

	eventMsg := map[string]interface{}{
		"event_id":         eventID,
		"event_type":       eventType,
		"is_supported":     isKnown,
		"signature_valid": sigStatus,
		"payload":          payload,
		"received_at":      time.Now().Format(time.RFC3339),
	}

	msgBytes, _ := json.Marshal(eventMsg)
	err = kafkaWriter.WriteMessages(c.Request.Context(), kafka.Message{
		Key:   []byte(eventType),
		Value: msgBytes,
	})

	if err != nil {
		log.Printf("[ERROR] Failed to publish webhook event %s to Redpanda: %v", eventType, err)
	} else {
		log.Printf("[INFO] Ingested & Published Razorpay event %s (Signature: %s, Supported: %v)", eventType, sigStatus, isKnown)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "RECEIVED",
		"event_type":      eventType,
		"supported":       isKnown,
		"signature_valid": sigStatus,
	})
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8001"
	}
	grpcPort := os.Getenv("PORT")
	if grpcPort == "" {
		grpcPort = "50001"
	}
	webhookSecret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	redpandaAddr := os.Getenv("REDPANDA_ADDR")
	if redpandaAddr == "" {
		redpandaAddr = "localhost:9092"
	}

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(redpandaAddr),
		Topic:    "PAYMENT_EVENTS",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	// Start HTTP Server
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":            "ok",
			"service":           "webhook-receiver",
			"supported_events":  len(SupportedWebhookEvents),
			"webhook_secret_set": bool(webhookSecret != ""),
		})
	})

	// Webhook Endpoints
	r.POST("/webhook/razorpay", func(c *gin.Context) {
		handleWebhookRequest(c, webhookSecret, kafkaWriter)
	})
	r.POST("/webhook", func(c *gin.Context) {
		handleWebhookRequest(c, webhookSecret, kafkaWriter)
	})

	go func() {
		log.Printf("Starting Webhook Receiver HTTP server on :%s (Endpoints: /webhook, /webhook/razorpay)", httpPort)
		if err := r.Run(":" + httpPort); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWebhookServiceServer(grpcServer, &server{})

	log.Printf("Starting Webhook Receiver gRPC server on :%s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
