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

func (m *MongoDatabase) SaveAnalyticsSnapshot(ctx context.Context, snap *pb.MongoAnalyticsSnapshotData) error {
	now := time.Now().Unix()
	if snap.Timestamp == 0 {
		snap.Timestamp = now
	}

	doc := bson.M{
		"timestamp":             snap.Timestamp,
		"total_drones":          snap.TotalDrones,
		"active_drones":         snap.ActiveDrones,
		"maintenance_drones":    snap.MaintenanceDrones,
		"avg_battery_drain":     snap.AvgBatteryDrain,
		"total_orders":          snap.TotalOrders,
		"completed_orders":      snap.CompletedOrders,
		"failed_orders":         snap.FailedOrders,
		"delivery_success_rate": snap.DeliverySuccessRate,
		"total_revenue":         snap.TotalRevenue,
	}

	_, err := m.AnalyticsSnapshotsCol.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to save analytics snapshot: %w", err)
	}
	return nil
}

func (m *MongoDatabase) GetLatestAnalyticsSnapshot(ctx context.Context) (*pb.MongoAnalyticsSnapshotData, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	var doc struct {
		Timestamp           int64   `bson:"timestamp"`
		TotalDrones         int32   `bson:"total_drones"`
		ActiveDrones        int32   `bson:"active_drones"`
		MaintenanceDrones   int32   `bson:"maintenance_drones"`
		AvgBatteryDrain     float64 `bson:"avg_battery_drain"`
		TotalOrders         int64   `bson:"total_orders"`
		CompletedOrders     int64   `bson:"completed_orders"`
		FailedOrders        int64   `bson:"failed_orders"`
		DeliverySuccessRate float64 `bson:"delivery_success_rate"`
		TotalRevenue        float64 `bson:"total_revenue"`
	}

	err := m.AnalyticsSnapshotsCol.FindOne(ctx, bson.M{}, opts).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &pb.MongoAnalyticsSnapshotData{
				Timestamp:           time.Now().Unix(),
				TotalDrones:         10,
				ActiveDrones:        4,
				MaintenanceDrones:   1,
				AvgBatteryDrain:     1.2,
				TotalOrders:         25,
				CompletedOrders:     22,
				FailedOrders:        3,
				DeliverySuccessRate: 88.0,
				TotalRevenue:        1249.75,
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch analytics snapshot: %w", err)
	}

	return &pb.MongoAnalyticsSnapshotData{
		Timestamp:           doc.Timestamp,
		TotalDrones:         doc.TotalDrones,
		ActiveDrones:        doc.ActiveDrones,
		MaintenanceDrones:   doc.MaintenanceDrones,
		AvgBatteryDrain:     doc.AvgBatteryDrain,
		TotalOrders:         doc.TotalOrders,
		CompletedOrders:     doc.CompletedOrders,
		FailedOrders:        doc.FailedOrders,
		DeliverySuccessRate: doc.DeliverySuccessRate,
		TotalRevenue:        doc.TotalRevenue,
	}, nil
}

func (m *MongoDatabase) SaveFailureEvent(ctx context.Context, ev *pb.MongoFailureEventData) error {
	now := time.Now().Unix()
	if ev.Timestamp == 0 {
		ev.Timestamp = now
	}
	if ev.EventId == "" {
		ev.EventId = fmt.Sprintf("fail-%d", now)
	}

	doc := bson.M{
		"_id":          ev.EventId,
		"category":     ev.Category,
		"service_name": ev.ServiceName,
		"description":  ev.Description,
		"timestamp":    ev.Timestamp,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.FailureEventsCol.ReplaceOne(ctx, bson.M{"_id": ev.EventId}, doc, opts)
	if err != nil {
		return fmt.Errorf("failed to save failure event: %w", err)
	}
	return nil
}

func (m *MongoDatabase) ListFailureEvents(ctx context.Context, category string, limit int32) ([]*pb.MongoFailureEventData, error) {
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := m.FailureEventsCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list failure events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []*pb.MongoFailureEventData
	for cursor.Next(ctx) {
		var doc struct {
			ID          string `bson:"_id"`
			Category    string `bson:"category"`
			ServiceName string `bson:"service_name"`
			Description string `bson:"description"`
			Timestamp   int64  `bson:"timestamp"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		events = append(events, &pb.MongoFailureEventData{
			EventId:     doc.ID,
			Category:    doc.Category,
			ServiceName: doc.ServiceName,
			Description: doc.Description,
			Timestamp:   doc.Timestamp,
		})
	}
	return events, nil
}
