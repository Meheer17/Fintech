package db

import (
	"context"
	"fmt"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDatabase) SaveWorkflow(ctx context.Context, wf *pb.MongoWorkflowData) error {
	now := time.Now().Unix()
	if wf.CreatedAt == 0 {
		wf.CreatedAt = now
	}
	if wf.WorkflowId == "" {
		wf.WorkflowId = fmt.Sprintf("wf_%s_%d", wf.PaymentId, now)
	}

	doc := bson.M{
		"_id":                 wf.WorkflowId,
		"workflow_id":         wf.WorkflowId,
		"payment_id":          wf.PaymentId,
		"amount_paise":        wf.AmountPaise,
		"status":              wf.Status,
		"action":              wf.Action,
		"payment_link_id":     wf.PaymentLinkId,
		"short_url":           wf.ShortUrl,
		"reason":              wf.Reason,
		"guardrails_checked": wf.GuardrailsChecked,
		"created_at":          wf.CreatedAt,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.WorkflowsCol.ReplaceOne(ctx, bson.M{"payment_id": wf.PaymentId}, doc, opts)
	if err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}
	return nil
}

func (m *MongoDatabase) ListWorkflows(ctx context.Context, status string, limit int32) ([]*pb.MongoWorkflowData, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := m.WorkflowsCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	defer cursor.Close(ctx)

	var workflows []*pb.MongoWorkflowData
	for cursor.Next(ctx) {
		var doc struct {
			WorkflowID        string   `bson:"workflow_id"`
			PaymentID         string   `bson:"payment_id"`
			AmountPaise       float64  `bson:"amount_paise"`
			Status            string   `bson:"status"`
			Action            string   `bson:"action"`
			PaymentLinkID     string   `bson:"payment_link_id"`
			ShortURL          string   `bson:"short_url"`
			Reason            string   `bson:"reason"`
			GuardrailsChecked []string `bson:"guardrails_checked"`
			CreatedAt         int64    `bson:"created_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		workflows = append(workflows, &pb.MongoWorkflowData{
			WorkflowId:        doc.WorkflowID,
			PaymentId:         doc.PaymentID,
			AmountPaise:       doc.AmountPaise,
			Status:            doc.Status,
			Action:            doc.Action,
			PaymentLinkId:     doc.PaymentLinkID,
			ShortUrl:          doc.ShortURL,
			Reason:            doc.Reason,
			GuardrailsChecked: doc.GuardrailsChecked,
			CreatedAt:         doc.CreatedAt,
		})
	}
	return workflows, nil
}
