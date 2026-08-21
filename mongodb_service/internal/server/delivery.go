package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) SaveDelivery(ctx context.Context, req *pb.SaveMongoDeliveryRequest) (*pb.MongoDeliveryResponse, error) {
	if req.Delivery == nil {
		return &pb.MongoDeliveryResponse{Success: false, Message: "Delivery data required"}, nil
	}
	del, err := s.db.SaveDelivery(ctx, req.Delivery)
	if err != nil {
		return &pb.MongoDeliveryResponse{Success: false, Message: fmt.Sprintf("Failed to save delivery: %v", err)}, nil
	}
	return &pb.MongoDeliveryResponse{Success: true, Message: "Delivery record saved", Delivery: del}, nil
}

func (s *Server) GetDelivery(ctx context.Context, req *pb.GetMongoDeliveryRequest) (*pb.MongoDeliveryResponse, error) {
	del, err := s.db.GetDelivery(ctx, req.DeliveryId, req.OrderId)
	if err != nil {
		return &pb.MongoDeliveryResponse{Success: false, Message: fmt.Sprintf("Failed to fetch delivery: %v", err)}, nil
	}
	if del == nil {
		return &pb.MongoDeliveryResponse{Success: false, Message: "Delivery record not found"}, nil
	}
	return &pb.MongoDeliveryResponse{Success: true, Message: "Delivery record found", Delivery: del}, nil
}
