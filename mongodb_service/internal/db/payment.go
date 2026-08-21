package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDatabase) SaveTransaction(ctx context.Context, tx *pb.MongoTxData) (*pb.MongoTxData, error) {
	now := time.Now().Unix()
	if tx.Timestamp == 0 {
		tx.Timestamp = now
	}
	if tx.TransactionId == "" {
		tx.TransactionId = fmt.Sprintf("tx-%d", now)
	}

	doc := bson.M{
		"_id":                  tx.TransactionId,
		"order_id":             tx.OrderId,
		"user_id":              tx.UserId,
		"amount":               tx.Amount,
		"currency":             tx.Currency,
		"type":                 tx.Type,
		"status":               tx.Status,
		"payment_method":       tx.PaymentMethod,
		"razorpay_payment_id": tx.RazorpayPaymentId,
		"razorpay_order_id":   tx.RazorpayOrderId,
		"razorpay_refund_id":  tx.RazorpayRefundId,
		"timestamp":            tx.Timestamp,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := m.TxCol.ReplaceOne(ctx, bson.M{"_id": tx.TransactionId}, doc, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	return tx, nil
}

func (m *MongoDatabase) GetUserWallet(ctx context.Context, userID string) (*pb.MongoWalletResponse, error) {
	var doc struct {
		UserID    string  `bson:"_id"`
		Balance   float64 `bson:"balance"`
		Currency  string  `bson:"currency"`
		UpdatedAt int64   `bson:"updated_at"`
	}

	err := m.WalletCol.FindOne(ctx, bson.M{"_id": userID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Return default wallet with 0 balance
			return &pb.MongoWalletResponse{
				Success:   true,
				Message:   "Wallet initialized",
				UserId:    userID,
				Balance:   0.0,
				Currency:  "INR",
				UpdatedAt: time.Now().Unix(),
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch user wallet: %w", err)
	}

	return &pb.MongoWalletResponse{
		Success:   true,
		Message:   "Wallet retrieved successfully",
		UserId:    doc.UserID,
		Balance:   doc.Balance,
		Currency:  doc.Currency,
		UpdatedAt: doc.UpdatedAt,
	}, nil
}

func (m *MongoDatabase) UpdateUserWallet(ctx context.Context, userID string, delta float64) (*pb.MongoWalletResponse, error) {
	now := time.Now().Unix()

	filter := bson.M{"_id": userID}
	update := bson.M{
		"$inc": bson.M{"balance": delta},
		"$setOnInsert": bson.M{
			"currency": "INR",
		},
		"$set": bson.M{
			"updated_at": now,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var doc struct {
		UserID    string  `bson:"_id"`
		Balance   float64 `bson:"balance"`
		Currency  string  `bson:"currency"`
		UpdatedAt int64   `bson:"updated_at"`
	}

	err := m.WalletCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to update user wallet: %w", err)
	}

	return &pb.MongoWalletResponse{
		Success:   true,
		Message:   "Wallet balance updated successfully",
		UserId:    doc.UserID,
		Balance:   doc.Balance,
		Currency:  doc.Currency,
		UpdatedAt: doc.UpdatedAt,
	}, nil
}

func (m *MongoDatabase) ListUserTransactions(ctx context.Context, userID string, limit int32) ([]*pb.MongoTxData, error) {
	filter := bson.M{}
	if userID != "" {
		filter["user_id"] = userID
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}

	cursor, err := m.TxCol.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var txs []*pb.MongoTxData
	for cursor.Next(ctx) {
		var doc struct {
			ID                string  `bson:"_id"`
			OrderID           string  `bson:"order_id"`
			UserID            string  `bson:"user_id"`
			Amount            float64 `bson:"amount"`
			Currency          string  `bson:"currency"`
			Type              string  `bson:"type"`
			Status            string  `bson:"status"`
			PaymentMethod     string  `bson:"payment_method"`
			RazorpayPaymentID string  `bson:"razorpay_payment_id"`
			RazorpayOrderID   string  `bson:"razorpay_order_id"`
			RazorpayRefundID  string  `bson:"razorpay_refund_id"`
			Timestamp         int64   `bson:"timestamp"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		txs = append(txs, &pb.MongoTxData{
			TransactionId:     doc.ID,
			OrderId:           doc.OrderID,
			UserId:            doc.UserID,
			Amount:            doc.Amount,
			Currency:          doc.Currency,
			Type:              doc.Type,
			Status:            doc.Status,
			PaymentMethod:     doc.PaymentMethod,
			RazorpayPaymentId: doc.RazorpayPaymentID,
			RazorpayOrderId:   doc.RazorpayOrderID,
			RazorpayRefundId:  doc.RazorpayRefundID,
			Timestamp:         doc.Timestamp,
		})
	}
	return txs, nil
}
