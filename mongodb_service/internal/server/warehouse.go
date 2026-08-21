package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) SaveBusiness(ctx context.Context, req *pb.SaveMongoBusinessRequest) (*pb.MongoBusinessResponse, error) {
	if req.Business == nil {
		return &pb.MongoBusinessResponse{Success: false, Message: "Business data required"}, nil
	}
	biz, err := s.db.SaveBusiness(ctx, req.Business)
	if err != nil {
		return &pb.MongoBusinessResponse{Success: false, Message: fmt.Sprintf("Failed to save business: %v", err)}, nil
	}
	return &pb.MongoBusinessResponse{Success: true, Message: "Business registered", Business: biz}, nil
}

func (s *Server) SaveWarehouse(ctx context.Context, req *pb.SaveMongoWarehouseRequest) (*pb.MongoWarehouseResponse, error) {
	if req.Warehouse == nil {
		return &pb.MongoWarehouseResponse{Success: false, Message: "Warehouse data required"}, nil
	}
	wh, err := s.db.SaveWarehouse(ctx, req.Warehouse)
	if err != nil {
		return &pb.MongoWarehouseResponse{Success: false, Message: fmt.Sprintf("Failed to save warehouse: %v", err)}, nil
	}
	return &pb.MongoWarehouseResponse{Success: true, Message: "Warehouse registered", Warehouse: wh}, nil
}

func (s *Server) SaveInventory(ctx context.Context, req *pb.SaveMongoInventoryRequest) (*pb.MongoInventoryResponse, error) {
	if req.Item == nil {
		return &pb.MongoInventoryResponse{Success: false, Message: "Inventory data required"}, nil
	}
	item, err := s.db.SaveInventory(ctx, req.Item)
	if err != nil {
		return &pb.MongoInventoryResponse{Success: false, Message: fmt.Sprintf("Failed to save inventory: %v", err)}, nil
	}
	return &pb.MongoInventoryResponse{Success: true, Message: "Inventory updated", Item: item}, nil
}

func (s *Server) GetInventory(ctx context.Context, req *pb.GetMongoInventoryRequest) (*pb.ListMongoInventoryResponse, error) {
	items, err := s.db.GetInventory(ctx, req.BusinessId, req.WarehouseId, req.Sku)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory: %w", err)
	}
	return &pb.ListMongoInventoryResponse{Items: items, TotalCount: int32(len(items))}, nil
}
