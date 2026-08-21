package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDatabase) SaveDelivery(ctx context.Context, del *pb.MongoDeliveryData) (*pb.MongoDeliveryData, error) {
	now := time.Now().Unix()
	if del.CreatedAt == 0 {
		del.CreatedAt = now
	}
	if del.DeliveryId == "" {
		del.DeliveryId = fmt.Sprintf("del-%d", now)
	}

	doc := bson.M{
		"_id":            del.DeliveryId,
		"order_id":       del.OrderId,
		"drone_id":       del.DroneId,
		"status":         del.Status,
		"photo_s3_key":   del.PhotoS3Key,
		"photo_url":      del.PhotoUrl,
		"otp":            del.Otp,
		"otp_verified":   del.OtpVerified,
		"failure_reason": del.FailureReason,
		"created_at":     del.CreatedAt,
		"completed_at":   del.CompletedAt,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.DeliveryCol.ReplaceOne(ctx, bson.M{"_id": del.DeliveryId}, doc, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save delivery record: %w", err)
	}
	return del, nil
}

func (m *MongoDatabase) GetDelivery(ctx context.Context, deliveryID, orderID string) (*pb.MongoDeliveryData, error) {
	filter := bson.M{}
	if deliveryID != "" {
		filter["_id"] = deliveryID
	} else if orderID != "" {
		filter["order_id"] = orderID
	} else {
		return nil, errors.New("delivery_id or order_id required")
	}

	var doc struct {
		ID            string `bson:"_id"`
		OrderID       string `bson:"order_id"`
		DroneID       string `bson:"drone_id"`
		Status        string `bson:"status"`
		PhotoS3Key    string `bson:"photo_s3_key"`
		PhotoURL      string `bson:"photo_url"`
		OTP           string `bson:"otp"`
		OTPVerified   bool   `bson:"otp_verified"`
		FailureReason string `bson:"failure_reason"`
		CreatedAt     int64  `bson:"created_at"`
		CompletedAt   int64  `bson:"completed_at"`
	}

	err := m.DeliveryCol.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch delivery record: %w", err)
	}

	return &pb.MongoDeliveryData{
		DeliveryId:    doc.ID,
		OrderId:       doc.OrderID,
		DroneId:       doc.DroneID,
		Status:        doc.Status,
		PhotoS3Key:    doc.PhotoS3Key,
		PhotoUrl:      doc.PhotoURL,
		Otp:           doc.OTP,
		OtpVerified:   doc.OTPVerified,
		FailureReason: doc.FailureReason,
		CreatedAt:     doc.CreatedAt,
		CompletedAt:   doc.CompletedAt,
	}, nil
}
