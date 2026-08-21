package redpanda

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type RedpandaReader struct {
	reader *kafka.Reader
}

func NewRedpandaReader(brokerAddr string, topic string, groupID string) *RedpandaReader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  1 * time.Second,
	})
	return &RedpandaReader{reader: r}
}

func (rr *RedpandaReader) Close() error {
	return rr.reader.Close()
}

func (rr *RedpandaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return rr.reader.FetchMessage(ctx)
}

func (rr *RedpandaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return rr.reader.CommitMessages(ctx, msgs...)
}
