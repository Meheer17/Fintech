package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) SaveWorkflow(ctx context.Context, req *pb.SaveWorkflowRequest) (*pb.MongoWorkflowResponse, error) {
	if req.Workflow == nil {
		return &pb.MongoWorkflowResponse{Success: false, Message: "Workflow data required"}, nil
	}
	err := s.db.SaveWorkflow(ctx, req.Workflow)
	if err != nil {
		return &pb.MongoWorkflowResponse{Success: false, Message: fmt.Sprintf("Failed to save workflow: %v", err)}, nil
	}
	return &pb.MongoWorkflowResponse{Success: true, Message: "Workflow saved", Workflow: req.Workflow}, nil
}

func (s *Server) ListWorkflows(ctx context.Context, req *pb.ListWorkflowsRequest) (*pb.ListWorkflowsResponse, error) {
	workflows, err := s.db.ListWorkflows(ctx, req.Status, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	return &pb.ListWorkflowsResponse{Workflows: workflows, TotalCount: int32(len(workflows))}, nil
}
