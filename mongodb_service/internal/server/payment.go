package server

import (
	"context"
	"fmt"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
)

func (s *Server) SaveTransaction(ctx context.Context, req *pb.SaveMongoTxRequest) (*pb.MongoTxResponse, error) {
	if req.Transaction == nil {
		return &pb.MongoTxResponse{Success: false, Message: "Transaction data required"}, nil
	}
	tx, err := s.db.SaveTransaction(ctx, req.Transaction)
	if err != nil {
		return &pb.MongoTxResponse{Success: false, Message: fmt.Sprintf("Failed to save transaction: %v", err)}, nil
	}
	return &pb.MongoTxResponse{Success: true, Message: "Transaction saved", Transaction: tx}, nil
}

func (s *Server) GetUserWallet(ctx context.Context, req *pb.GetUserWalletRequest) (*pb.MongoWalletResponse, error) {
	res, err := s.db.GetUserWallet(ctx, req.UserId)
	if err != nil {
		return &pb.MongoWalletResponse{Success: false, Message: fmt.Sprintf("Failed to fetch wallet: %v", err)}, nil
	}
	return res, nil
}

func (s *Server) UpdateUserWallet(ctx context.Context, req *pb.UpdateUserWalletRequest) (*pb.MongoWalletResponse, error) {
	res, err := s.db.UpdateUserWallet(ctx, req.UserId, req.DeltaAmount)
	if err != nil {
		return &pb.MongoWalletResponse{Success: false, Message: fmt.Sprintf("Failed to update wallet: %v", err)}, nil
	}
	return res, nil
}

func (s *Server) ListUserTransactions(ctx context.Context, req *pb.ListUserTxRequest) (*pb.ListUserTxResponse, error) {
	txs, err := s.db.ListUserTransactions(ctx, req.UserId, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	return &pb.ListUserTxResponse{Transactions: txs, TotalCount: int32(len(txs))}, nil
}
