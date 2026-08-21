package db

import (
	"context"
	"fmt"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NoFlyZoneDoc struct {
	ID           string  `bson:"_id,omitempty" json:"id"`
	Name         string  `bson:"name" json:"name"`
	CenterLat    float64 `bson:"center_lat" json:"center_lat"`
	CenterLng    float64 `bson:"center_lng" json:"center_lng"`
	RadiusMeters float64 `bson:"radius_meters" json:"radius_meters"`
	MinAltitude  float64 `bson:"min_altitude" json:"min_altitude"`
	MaxAltitude  float64 `bson:"max_altitude" json:"max_altitude"`
	Active       bool    `bson:"active" json:"active"`
	Reason       string  `bson:"reason" json:"reason"`
	CreatedAt    int64   `bson:"created_at" json:"created_at"`
}

func (m *MongoDatabase) AddNoFlyZone(ctx context.Context, zone *pb.NoFlyZoneData) (*pb.NoFlyZoneData, error) {
	now := time.Now().Unix()
	if zone.CreatedAt == 0 {
		zone.CreatedAt = now
	}
	if zone.Id == "" {
		zone.Id = fmt.Sprintf("nfz-%d", now)
	}

	doc := bson.M{
		"_id":           zone.Id,
		"name":          zone.Name,
		"center_lat":    zone.CenterLat,
		"center_lng":    zone.CenterLng,
		"radius_meters": zone.RadiusMeters,
		"min_altitude":  zone.MinAltitude,
		"max_altitude":  zone.MaxAltitude,
		"active":        zone.Active,
		"reason":        zone.Reason,
		"created_at":    zone.CreatedAt,
	}

	_, err := m.NoFlyCol.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert no-fly zone: %w", err)
	}

	return zone, nil
}

func (m *MongoDatabase) GetNoFlyZones(ctx context.Context, activeOnly bool) ([]*pb.NoFlyZoneData, error) {
	filter := bson.M{}
	if activeOnly {
		filter["active"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := m.NoFlyCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query no-fly zones: %w", err)
	}
	defer cursor.Close(ctx)

	var zones []*pb.NoFlyZoneData
	for cursor.Next(ctx) {
		var doc struct {
			ID           string  `bson:"_id"`
			Name         string  `bson:"name"`
			CenterLat    float64 `bson:"center_lat"`
			CenterLng    float64 `bson:"center_lng"`
			RadiusMeters float64 `bson:"radius_meters"`
			MinAltitude  float64 `bson:"min_altitude"`
			MaxAltitude  float64 `bson:"max_altitude"`
			Active       bool    `bson:"active"`
			Reason       string  `bson:"reason"`
			CreatedAt    int64   `bson:"created_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode no-fly zone doc: %w", err)
		}

		zones = append(zones, &pb.NoFlyZoneData{
			Id:           doc.ID,
			Name:         doc.Name,
			CenterLat:    doc.CenterLat,
			CenterLng:    doc.CenterLng,
			RadiusMeters: doc.RadiusMeters,
			MinAltitude:  doc.MinAltitude,
			MaxAltitude:  doc.MaxAltitude,
			Active:       doc.Active,
			Reason:       doc.Reason,
			CreatedAt:    doc.CreatedAt,
		})
	}

	return zones, nil
}
