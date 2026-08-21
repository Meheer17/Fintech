package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

type Session struct {
	SessionID  string    `json:"session_id"`
	MerchantID string    `json:"merchant_id"`
	Amount     int64     `json:"amount_paise"`
	LastSeen   time.Time `json:"last_seen"`
	Completed  bool      `json:"completed"`
	Abandoned  bool      `json:"abandoned"`
}

var (
	sessions = make(map[string]*Session)
	mu       sync.Mutex
)

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8008"
	}
	redpandaAddr := os.Getenv("REDPANDA_ADDR")
	if redpandaAddr == "" {
		redpandaAddr = "localhost:9092"
	}

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(redpandaAddr),
		Topic:    "CHECKOUT_EVENTS",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	// Background abandonment scanner
	go func() {
		for {
			time.Sleep(10 * time.Second)
			mu.Lock()
			now := time.Now()
			for id, sess := range sessions {
				if !sess.Completed && !sess.Abandoned && now.Sub(sess.LastSeen) > 30*time.Second {
					sess.Abandoned = true
					log.Printf("[CHECKOUT] Abandonment detected for session: %s (Merchant: %s)", id, sess.MerchantID)

					eventMsg := map[string]interface{}{
						"event_type":          "checkout.abandoned",
						"checkout_session_id": id,
						"merchant_id":         sess.MerchantID,
						"amount_paise":        sess.Amount,
						"abandoned_at":        now.Format(time.RFC3339),
					}
					msgBytes, _ := json.Marshal(eventMsg)
					_ = kafkaWriter.WriteMessages(nil, kafka.Message{
						Key:   []byte(id),
						Value: msgBytes,
					})
				}
			}
			mu.Unlock()
		}
	}()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "checkout-tracker"})
	})

	r.POST("/checkout/start", func(c *gin.Context) {
		var req struct {
			SessionID  string `json:"session_id"`
			MerchantID string `json:"merchant_id"`
			Amount     int64  `json:"amount_paise"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		mu.Lock()
		sessions[req.SessionID] = &Session{
			SessionID:  req.SessionID,
			MerchantID: req.MerchantID,
			Amount:     req.Amount,
			LastSeen:   time.Now(),
		}
		mu.Unlock()

		c.JSON(http.StatusOK, gin.H{"status": "STARTED", "session_id": req.SessionID})
	})

	r.POST("/checkout/heartbeat", func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		mu.Lock()
		if sess, exists := sessions[req.SessionID]; exists {
			sess.LastSeen = time.Now()
		}
		mu.Unlock()

		c.JSON(http.StatusOK, gin.H{"status": "HEARTBEAT_ACK"})
	})

	r.POST("/checkout/complete", func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		mu.Lock()
		if sess, exists := sessions[req.SessionID]; exists {
			sess.Completed = true
		}
		mu.Unlock()

		c.JSON(http.StatusOK, gin.H{"status": "COMPLETED"})
	})

	log.Printf("Starting RevenueIQ Checkout Tracker on :%s", httpPort)
	if err := r.Run(fmt.Sprintf(":%s", httpPort)); err != nil {
		log.Fatalf("Checkout tracker server error: %v", err)
	}
}
