package server

import (
	"context"

	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/models"
	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) RegisterDrone(ctx context.Context, req *pb.RegisterMongoDroneRequest) (*pb.MongoDroneResponse, error) {
	if req.GetDrone() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "drone is required")
	}

	droneDoc := convertDroneProtoToDoc(req.GetDrone())
	res, err := s.db.RegisterDrone(ctx, droneDoc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register drone in DB: %v", err)
	}

	return &pb.MongoDroneResponse{
		Success: true,
		Message: "Drone registered in MongoDB",
		Drone:   convertDroneDocToProto(res),
	}, nil
}

func (s *Server) GetDrone(ctx context.Context, req *pb.GetMongoDroneRequest) (*pb.MongoDroneResponse, error) {
	if req.GetDroneId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "drone_id is required")
	}

	doc, err := s.db.GetDroneByID(ctx, req.GetDroneId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get drone: %v", err)
	}
	if doc == nil {
		return nil, status.Errorf(codes.NotFound, "drone not found: %s", req.GetDroneId())
	}

	return &pb.MongoDroneResponse{
		Success: true,
		Message: "Drone retrieved from MongoDB",
		Drone:   convertDroneDocToProto(doc),
	}, nil
}

func (s *Server) UpdateDroneStatus(ctx context.Context, req *pb.UpdateMongoDroneStatusRequest) (*pb.MongoDroneResponse, error) {
	if req.GetDroneId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "drone_id is required")
	}

	var loc *models.LocationDoc
	if req.GetLocation() != nil {
		loc = &models.LocationDoc{
			Latitude:  req.GetLocation().GetLatitude(),
			Longitude: req.GetLocation().GetLongitude(),
		}
	}

	doc, err := s.db.UpdateDroneStatus(ctx, req.GetDroneId(), req.GetStatus(), req.GetBatteryLevel(), loc, req.GetCurrentOrderId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update drone status in DB: %v", err)
	}
	if doc == nil {
		return nil, status.Errorf(codes.NotFound, "drone not found: %s", req.GetDroneId())
	}

	return &pb.MongoDroneResponse{
		Success: true,
		Message: "Drone status updated in MongoDB",
		Drone:   convertDroneDocToProto(doc),
	}, nil
}

func (s *Server) FindAvailableDrone(ctx context.Context, req *pb.FindAvailableDroneRequest) (*pb.MongoDroneResponse, error) {
	doc, err := s.db.FindAvailableDrone(ctx, req.GetMinBattery())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find available drone: %v", err)
	}
	if doc == nil {
		return &pb.MongoDroneResponse{
			Success: false,
			Message: "No available drones found",
			Drone:   nil,
		}, nil
	}

	return &pb.MongoDroneResponse{
		Success: true,
		Message: "Available drone found",
		Drone:   convertDroneDocToProto(doc),
	}, nil
}

func (s *Server) ReserveDrone(ctx context.Context, req *pb.ReserveMongoDroneRequest) (*pb.MongoDroneResponse, error) {
	if req.GetDroneId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "drone_id is required")
	}

	doc, err := s.db.UpdateDroneStatus(ctx, req.GetDroneId(), "RESERVED", 0, nil, req.GetOrderId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve drone in DB: %v", err)
	}
	if doc == nil {
		return nil, status.Errorf(codes.NotFound, "drone not found: %s", req.GetDroneId())
	}

	return &pb.MongoDroneResponse{
		Success: true,
		Message: "Drone reserved in MongoDB",
		Drone:   convertDroneDocToProto(doc),
	}, nil
}

func (s *Server) AssignDrone(ctx context.Context, req *pb.AssignMongoDroneRequest) (*pb.MongoDroneResponse, error) {
	if req.GetDroneId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "drone_id is required")
	}

	doc, err := s.db.UpdateDroneStatus(ctx, req.GetDroneId(), "DELIVERING", 0, nil, req.GetOrderId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign drone in DB: %v", err)
	}
	if doc == nil {
		return nil, status.Errorf(codes.NotFound, "drone not found: %s", req.GetDroneId())
	}

	return &pb.MongoDroneResponse{
		Success: true,
		Message: "Drone assigned in MongoDB",
		Drone:   convertDroneDocToProto(doc),
	}, nil
}

func (s *Server) ListDrones(ctx context.Context, req *pb.ListMongoDronesRequest) (*pb.ListMongoDronesResponse, error) {
	docs, err := s.db.ListDrones(ctx, req.GetStatus(), req.GetMinBattery(), req.GetLimit())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list drones from DB: %v", err)
	}

	var protoDrones []*pb.DroneData
	for _, doc := range docs {
		protoDrones = append(protoDrones, convertDroneDocToProto(&doc))
	}

	return &pb.ListMongoDronesResponse{
		Drones:     protoDrones,
		TotalCount: int32(len(protoDrones)),
	}, nil
}

func convertDroneProtoToDoc(p *pb.DroneData) *models.DroneDoc {
	if p == nil {
		return &models.DroneDoc{}
	}
	doc := &models.DroneDoc{
		DroneID:           p.GetDroneId(),
		Model:             p.GetModel(),
		Status:            p.GetStatus(),
		BatteryLevel:      p.GetBatteryLevel(),
		PayloadCapacityKg: p.GetPayloadCapacityKg(),
		CurrentOrderID:    p.GetCurrentOrderId(),
		RegisteredAt:      p.GetRegisteredAt(),
		LastUpdatedAt:     p.GetLastUpdatedAt(),
	}
	if p.GetCurrentLocation() != nil {
		doc.CurrentLocation = models.LocationDoc{
			Latitude:  p.GetCurrentLocation().GetLatitude(),
			Longitude: p.GetCurrentLocation().GetLongitude(),
		}
	}
	return doc
}

func convertDroneDocToProto(doc *models.DroneDoc) *pb.DroneData {
	if doc == nil {
		return nil
	}
	return &pb.DroneData{
		DroneId:           doc.DroneID,
		Model:             doc.Model,
		Status:            doc.Status,
		BatteryLevel:      doc.BatteryLevel,
		PayloadCapacityKg: doc.PayloadCapacityKg,
		CurrentLocation: &pb.LocationData{
			Latitude:  doc.CurrentLocation.Latitude,
			Longitude: doc.CurrentLocation.Longitude,
		},
		CurrentOrderId: doc.CurrentOrderID,
		RegisteredAt:   doc.RegisteredAt,
		LastUpdatedAt:  doc.LastUpdatedAt,
	}
}
