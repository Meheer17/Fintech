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

func (m *MongoDatabase) RegisterDrone(ctx context.Context, drone *models.DroneDoc) (*models.DroneDoc, error) {
	now := time.Now().Unix()
	if drone.RegisteredAt == 0 {
		drone.RegisteredAt = now
	}
	drone.LastUpdatedAt = now
	if drone.Status == "" {
		drone.Status = "AVAILABLE"
	}

	_, err := m.DronesCol.InsertOne(ctx, drone)
	if err != nil {
		return nil, fmt.Errorf("failed to register drone: %w", err)
	}
	return drone, nil
}

func (m *MongoDatabase) GetDroneByID(ctx context.Context, droneID string) (*models.DroneDoc, error) {
	var drone models.DroneDoc
	err := m.DronesCol.FindOne(ctx, bson.M{"drone_id": droneID}).Decode(&drone)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find drone %q: %w", droneID, err)
	}
	return &drone, nil
}

func (m *MongoDatabase) UpdateDroneStatus(ctx context.Context, droneID, status string, batteryLevel float64, loc *models.LocationDoc, currentOrderID string) (*models.DroneDoc, error) {
	now := time.Now().Unix()

	updateFields := bson.M{
		"last_updated_at": now,
	}

	if status != "" {
		updateFields["status"] = status
	}
	if batteryLevel > 0 {
		updateFields["battery_level"] = batteryLevel
	}
	if loc != nil {
		updateFields["current_location"] = loc
	}
	if currentOrderID != "" {
		updateFields["current_order_id"] = currentOrderID
	}

	update := bson.M{"$set": updateFields}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedDrone models.DroneDoc
	err := m.DronesCol.FindOneAndUpdate(ctx, bson.M{"drone_id": droneID}, update, opts).Decode(&updatedDrone)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update drone status: %w", err)
	}

	return &updatedDrone, nil
}

func (m *MongoDatabase) FindAvailableDrone(ctx context.Context, minBattery float64) (*models.DroneDoc, error) {
	if minBattery <= 0 {
		minBattery = 20.0
	}

	filter := bson.M{
		"status":        "AVAILABLE",
		"battery_level": bson.M{"$gt": minBattery},
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "battery_level", Value: -1}})

	var drone models.DroneDoc
	err := m.DronesCol.FindOne(ctx, filter, opts).Decode(&drone)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find available drone: %w", err)
	}

	return &drone, nil
}

func (m *MongoDatabase) ListDrones(ctx context.Context, status string, minBattery float64, limit int32) ([]models.DroneDoc, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	if minBattery > 0 {
		filter["battery_level"] = bson.M{"$gte": minBattery}
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "last_updated_at", Value: -1}})
	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}

	cursor, err := m.DronesCol.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list drones: %w", err)
	}
	defer cursor.Close(ctx)

	var drones []models.DroneDoc
	if err := cursor.All(ctx, &drones); err != nil {
		return nil, fmt.Errorf("failed to decode drones: %w", err)
	}

	return drones, nil
}
