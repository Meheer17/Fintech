package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/namespaces"
	notification_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/notification_service"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
	"github.com/RevenueIQ/notification_service/internal/redpanda"
	"github.com/RevenueIQ/notification_service/internal/ses"
)

type QueuePayload struct {
	Topic  string `json:"topic"`
	Type   string `json:"type"`
	DataID string `json:"data_id"`
}

type Worker struct {
	reader      *redpanda.RedpandaReader
	redisClient redis_pb.RedisCacheServiceClient
	sesClient   ses.SESClient
}

func NewWorker(reader *redpanda.RedpandaReader, redisClient redis_pb.RedisCacheServiceClient, sesClient ses.SESClient) *Worker {
	return &Worker{
		reader:      reader,
		redisClient: redisClient,
		sesClient:   sesClient,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Printf("Worker started, listening for events...")
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker stopping...")
			return
		default:
			msg, err := w.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Worker reader closed or context done: %v", err)
				return
			}

			log.Printf("Fetched event from Redpanda. Partition: %d, Offset: %d, Key: %s", msg.Partition, msg.Offset, string(msg.Key))

			err = w.processEvent(ctx, msg.Value)
			if err != nil {
				log.Printf("Failed to process event: %v", err)
			}

			if err := w.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
		}
	}
}

func (w *Worker) processEvent(ctx context.Context, val []byte) error {
	var payload QueuePayload
	if err := json.Unmarshal(val, &payload); err != nil {
		return err
	}

	log.Printf("Processing event - Topic: %s, Type: %s, DataID: %s", payload.Topic, payload.Type, payload.DataID)

	if payload.Topic != "EMAIL" {
		log.Printf("Ignoring event with topic %q. We only process EMAIL topic.", payload.Topic)
		return nil
	}

	// Fetch data from Redis via gRPC
	redisResp, err := w.redisClient.Get(ctx, &redis_pb.GetRequest{
		Key:       payload.DataID,
		Namespace: string(namespaces.QueueData),
	})
	if err != nil {
		return err
	}
	if !redisResp.GetFound() {
		log.Printf("Warning: Key %s not found in Redis (namespace: %s)", payload.DataID, namespaces.QueueData)
		return nil
	}

	// Unmarshal value from Redis
	var emailPayload notification_pb.EmailPayload
	if err := redisResp.GetValue().UnmarshalTo(&emailPayload); err != nil {
		return err
	}

	// Act based on the type
	switch payload.Type {
	case "signup":
		log.Printf("Sending signup email...")
		err = w.sesClient.SendEmail(ctx,
			emailPayload.GetToAddress(),
			emailPayload.GetSubject(),
			emailPayload.GetBodyHtml(),
			emailPayload.GetBodyText(),
		)
		if err != nil {
			return err
		}
	case "otp":
		log.Printf("Sending OTP email...")
		err = w.sesClient.SendEmail(ctx,
			emailPayload.GetToAddress(),
			emailPayload.GetSubject(),
			emailPayload.GetBodyHtml(),
			emailPayload.GetBodyText(),
		)
		if err != nil {
			return err
		}
	default:
		log.Printf("Unknown or unsupported event type %q for email topic", payload.Type)
	}

	return nil
}
