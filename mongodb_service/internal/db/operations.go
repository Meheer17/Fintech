package db

import (
	"context"
	"fmt"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDatabase) RegisterStation(ctx context.Context, st *pb.StationData) (*pb.StationData, error) {
	now := time.Now().Unix()
	if st.UpdatedAt == 0 {
		st.UpdatedAt = now
	}
	if st.StationId == "" {
		st.StationId = fmt.Sprintf("st-%d", now)
	}

	doc := bson.M{
		"_id":              st.StationId,
		"name":             st.Name,
		"latitude":         st.Latitude,
		"longitude":        st.Longitude,
		"total_pads":       st.TotalPads,
		"occupied_pads":    st.OccupiedPads,
		"status":           st.Status,
		"parked_drone_ids": st.ParkedDroneIds,
		"updated_at":       st.UpdatedAt,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.StationsCol.ReplaceOne(ctx, bson.M{"_id": st.StationId}, doc, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register station: %w", err)
	}

	return st, nil
}

func (m *MongoDatabase) GetStations(ctx context.Context, status string) ([]*pb.StationData, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	cursor, err := m.StationsCol.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query stations: %w", err)
	}
	defer cursor.Close(ctx)

	var stations []*pb.StationData
	for cursor.Next(ctx) {
		var doc struct {
			ID             string   `bson:"_id"`
			Name           string   `bson:"name"`
			Latitude       float64  `bson:"latitude"`
			Longitude      float64  `bson:"longitude"`
			TotalPads      int32    `bson:"total_pads"`
			OccupiedPads   int32    `bson:"occupied_pads"`
			Status         string   `bson:"status"`
			ParkedDroneIDs []string `bson:"parked_drone_ids"`
			UpdatedAt      int64    `bson:"updated_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		stations = append(stations, &pb.StationData{
			StationId:      doc.ID,
			Name:           doc.Name,
			Latitude:       doc.Latitude,
			Longitude:      doc.Longitude,
			TotalPads:      doc.TotalPads,
			OccupiedPads:   doc.OccupiedPads,
			Status:         doc.Status,
			ParkedDroneIds: doc.ParkedDroneIDs,
			UpdatedAt:      doc.UpdatedAt,
		})
	}
	return stations, nil
}

func (m *MongoDatabase) SaveMaintenance(ctx context.Context, rec *pb.MaintenanceData) (*pb.MaintenanceData, error) {
	now := time.Now().Unix()
	if rec.ScheduledAt == 0 {
		rec.ScheduledAt = now
	}
	if rec.RecordId == "" {
		rec.RecordId = fmt.Sprintf("maint-%d", now)
	}

	doc := bson.M{
		"_id":                 rec.RecordId,
		"drone_id":            rec.DroneId,
		"issue_description":  rec.IssueDescription,
		"maintenance_type":   rec.MaintenanceType,
		"status":             rec.Status,
		"assigned_technician": rec.AssignedTechnician,
		"scheduled_at":       rec.ScheduledAt,
		"completed_at":       rec.CompletedAt,
		"notes":              rec.Notes,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.MaintCol.ReplaceOne(ctx, bson.M{"_id": rec.RecordId}, doc, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save maintenance record: %w", err)
	}

	return rec, nil
}
