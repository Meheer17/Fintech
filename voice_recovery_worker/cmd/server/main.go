package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

type VoiceTask struct {
	PaymentID    string `json:"payment_id"`
	CustomerName string `json:"customer_name"`
	AmountPaise  int64  `json:"amount_paise"`
	Language     string `json:"language"`
}

func main() {
	httpPort := os.Getenv("HEALTH_PORT")
	if httpPort == "" {
		httpPort = "8009"
	}
	redpandaAddr := os.Getenv("REDPANDA_ADDR")
	if redpandaAddr == "" {
		redpandaAddr = "localhost:9092"
	}

	// Redpanda Consumer
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{redpandaAddr},
		Topic:     "VOICE_RECOVERY",
		GroupID:   "voice-recovery-group",
		MinBytes:  10,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	go func() {
		log.Println("Starting Voice Recovery Worker Redpanda consumer for VOICE_RECOVERY topic...")
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("[ERROR] Redpanda reader error: %v", err)
				continue
			}
			var task VoiceTask
			if err := json.Unmarshal(m.Value, &task); err == nil {
				script := fmt.Sprintf("Namaste %s, aapka ₹%.2f ka payment fail ho gaya tha. Recovery link aapke WhatsApp par bhej diya gaya hai.",
					task.CustomerName, float64(task.AmountPaise)/100.0)
				log.Printf("[VOICE IVR DISPATCH] Call to %s | Script: %s", task.PaymentID, script)
			}
		}
	}()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "voice-recovery-worker"})
	})

	log.Printf("Starting Voice Recovery Worker health check on :%s", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		log.Fatalf("Voice worker HTTP server error: %v", err)
	}
}
