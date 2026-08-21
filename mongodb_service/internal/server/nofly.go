package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) AddNoFlyZone(ctx context.Context, req *pb.AddMongoNoFlyZoneRequest) (*pb.MongoNoFlyZoneResponse, error) {
	if req.Zone == nil {
		return &pb.MongoNoFlyZoneResponse{
			Success: false,
			Message: "Zone details required",
		}, nil
	}

	savedZone, err := s.db.AddNoFlyZone(ctx, req.Zone)
	if err != nil {
		return &pb.MongoNoFlyZoneResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to save No-Fly Zone: %v", err),
		}, nil
	}

	return &pb.MongoNoFlyZoneResponse{
		Success: true,
		Message: "No-Fly Zone saved successfully",
		Zone:    savedZone,
	}, nil
}

func (s *Server) GetNoFlyZones(ctx context.Context, req *pb.GetMongoNoFlyZonesRequest) (*pb.ListMongoNoFlyZonesResponse, error) {
	zones, err := s.db.GetNoFlyZones(ctx, req.ActiveOnly)
	if err != nil {
		return &pb.ListMongoNoFlyZonesResponse{
			Zones:      nil,
			TotalCount: 0,
		}, fmt.Errorf("failed to fetch No-Fly Zones: %w", err)
	}

	return &pb.ListMongoNoFlyZonesResponse{
		Zones:      zones,
		TotalCount: int32(len(zones)),
	}, nil
}
