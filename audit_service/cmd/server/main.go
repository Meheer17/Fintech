package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"

	common_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/common"
	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/audit"
)

type server struct {
	pb.UnimplementedAuditServiceServer
	mu      sync.Mutex
	entries []*pb.AuditEntry
}

func (s *server) LogAction(ctx context.Context, req *pb.LogActionRequest) (*pb.LogActionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := req.GetEntry()
	if entry.GetId() == "" {
		entry.Id = fmt.Sprintf("audit_%d", len(s.entries)+1)
	}
	s.entries = append(s.entries, entry)
	log.Printf("[AUDIT] Action: %s | Service: %s | Status: %s | Reasoning: %s",
		entry.GetAction(), entry.GetServiceName(), entry.GetStatus(), entry.GetReasoning())

	return &pb.LogActionResponse{
		Status:   &common_pb.StatusResponse{Success: true, RequestId: req.GetRequestId()},
		AuditId: entry.GetId(),
	}, nil
}

func (s *server) GetAuditTrail(ctx context.Context, req *pb.GetAuditTrailRequest) (*pb.GetAuditTrailResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []*pb.AuditEntry
	serviceFilter := req.GetServiceName()

	for _, e := range s.entries {
		if serviceFilter == "" || e.GetServiceName() == serviceFilter {
			filtered = append(filtered, e)
		}
	}

	return &pb.GetAuditTrailResponse{
		Status:  &common_pb.StatusResponse{Success: true, RequestId: req.GetRequestId()},
		Entries: filtered,
	}, nil
}

func (s *server) GetActionsByEntity(ctx context.Context, req *pb.GetEntityActionsRequest) (*pb.GetAuditTrailResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []*pb.AuditEntry
	entityType := req.GetEntityType()
	entityID := req.GetEntityId()

	for _, e := range s.entries {
		if (entityType == "" || e.GetEntityType() == entityType) && (entityID == "" || e.GetEntityId() == entityID) {
			filtered = append(filtered, e)
		}
	}

	return &pb.GetAuditTrailResponse{
		Status:  &common_pb.StatusResponse{Success: true, RequestId: req.GetRequestId()},
		Entries: filtered,
	}, nil
}

func main() {
	grpcPort := os.Getenv("PORT")
	if grpcPort == "" {
		grpcPort = "50007"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	s := &server{entries: make([]*pb.AuditEntry, 0)}
	pb.RegisterAuditServiceServer(grpcServer, s)

	log.Printf("Starting RevenueIQ Audit Service gRPC server on :%s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
