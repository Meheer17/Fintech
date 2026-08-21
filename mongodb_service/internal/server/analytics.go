package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) SaveAnalyticsSnapshot(ctx context.Context, req *pb.SaveAnalyticsSnapshotRequest) (*pb.MongoAnalyticsResponse, error) {
	if req.Snapshot == nil {
		return &pb.MongoAnalyticsResponse{Success: false, Message: "Snapshot data required"}, nil
	}
	err := s.db.SaveAnalyticsSnapshot(ctx, req.Snapshot)
	if err != nil {
		return &pb.MongoAnalyticsResponse{Success: false, Message: fmt.Sprintf("Failed to save snapshot: %v", err)}, nil
	}
	return &pb.MongoAnalyticsResponse{Success: true, Message: "Analytics snapshot saved"}, nil
}

func (s *Server) GetLatestAnalyticsSnapshot(ctx context.Context, req *pb.GetAnalyticsSnapshotRequest) (*pb.MongoAnalyticsSnapshotData, error) {
	snap, err := s.db.GetLatestAnalyticsSnapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch snapshot: %w", err)
	}
	return snap, nil
}

func (s *Server) SaveFailureEvent(ctx context.Context, req *pb.SaveFailureEventRequest) (*pb.MongoAnalyticsResponse, error) {
	if req.Event == nil {
		return &pb.MongoAnalyticsResponse{Success: false, Message: "Event data required"}, nil
	}
	err := s.db.SaveFailureEvent(ctx, req.Event)
	if err != nil {
		return &pb.MongoAnalyticsResponse{Success: false, Message: fmt.Sprintf("Failed to save failure event: %v", err)}, nil
	}
	return &pb.MongoAnalyticsResponse{Success: true, Message: "Failure event logged"}, nil
}

func (s *Server) ListFailureEvents(ctx context.Context, req *pb.ListFailureEventsRequest) (*pb.ListFailureEventsResponse, error) {
	events, err := s.db.ListFailureEvents(ctx, req.Category, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list failure events: %w", err)
	}
	return &pb.ListFailureEventsResponse{Events: events, TotalCount: int32(len(events))}, nil
}
