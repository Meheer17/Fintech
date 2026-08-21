package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) RegisterStation(ctx context.Context, req *pb.RegisterMongoStationRequest) (*pb.MongoStationResponse, error) {
	if req.Station == nil {
		return &pb.MongoStationResponse{Success: false, Message: "Station data required"}, nil
	}
	st, err := s.db.RegisterStation(ctx, req.Station)
	if err != nil {
		return &pb.MongoStationResponse{Success: false, Message: fmt.Sprintf("Failed to register station: %v", err)}, nil
	}
	return &pb.MongoStationResponse{Success: true, Message: "Station registered", Station: st}, nil
}

func (s *Server) GetStations(ctx context.Context, req *pb.GetMongoStationsRequest) (*pb.ListMongoStationsResponse, error) {
	stations, err := s.db.GetStations(ctx, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stations: %w", err)
	}
	return &pb.ListMongoStationsResponse{Stations: stations, TotalCount: int32(len(stations))}, nil
}

func (s *Server) SaveMaintenance(ctx context.Context, req *pb.SaveMongoMaintenanceRequest) (*pb.MongoMaintenanceResponse, error) {
	if req.Record == nil {
		return &pb.MongoMaintenanceResponse{Success: false, Message: "Maintenance record required"}, nil
	}
	rec, err := s.db.SaveMaintenance(ctx, req.Record)
	if err != nil {
		return &pb.MongoMaintenanceResponse{Success: false, Message: fmt.Sprintf("Failed to save maintenance: %v", err)}, nil
	}
	return &pb.MongoMaintenanceResponse{Success: true, Message: "Maintenance record saved", Record: rec}, nil
}
