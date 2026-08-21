package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDatabase) CreateOrder(ctx context.Context, order *models.OrderDoc) (*models.OrderDoc, error) {
	now := time.Now().Unix()
	if order.CreatedAt == 0 {
		order.CreatedAt = now
	}
	order.UpdatedAt = now

	_, err := m.OrdersCol.InsertOne(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}
	return order, nil
}

func (m *MongoDatabase) GetOrderByID(ctx context.Context, orderID string) (*models.OrderDoc, error) {
	var order models.OrderDoc
	err := m.OrdersCol.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find order %q: %w", orderID, err)
	}
	return &order, nil
}

func (m *MongoDatabase) UpdateOrderStatus(ctx context.Context, orderID, status, assignedDroneID, note string) (*models.OrderDoc, error) {
	now := time.Now().Unix()

	historyItem := models.OrderHistoryItemDoc{
		Status:    status,
		Timestamp: now,
		Note:      note,
	}

	updateFields := bson.M{
		"status":     status,
		"updated_at": now,
	}
	if assignedDroneID != "" {
		updateFields["assigned_drone_id"] = assignedDroneID
	}

	update := bson.M{
		"$set": updateFields,
		"$push": bson.M{
			"history": historyItem,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedOrder models.OrderDoc

	err := m.OrdersCol.FindOneAndUpdate(ctx, bson.M{"order_id": orderID}, update, opts).Decode(&updatedOrder)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return &updatedOrder, nil
}

func (m *MongoDatabase) ListOrders(ctx context.Context, customerID, status string, limit int32) ([]models.OrderDoc, error) {
	filter := bson.M{}
	if customerID != "" {
		filter["customer_id"] = customerID
	}
	if status != "" {
		filter["status"] = status
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}

	cursor, err := m.OrdersCol.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []models.OrderDoc
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode orders: %w", err)
	}

	return orders, nil
}
