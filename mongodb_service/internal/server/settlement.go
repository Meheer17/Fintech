package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) GetSettlementAggregations(ctx context.Context, req *pb.GetSettlementAggregationsRequest) (*pb.GetSettlementAggregationsResponse, error) {
	days := int32(7)
	if req != nil && req.Days > 0 {
		days = req.Days
	}
	res, err := s.db.GetSettlementAggregations(ctx, days)
	if err != nil {
		return &pb.GetSettlementAggregationsResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to aggregate settlements: %v", err),
		}, nil
	}
	return res, nil
}
