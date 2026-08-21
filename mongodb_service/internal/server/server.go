package server

import (
	"github.com/RevenueIQ/mongo_service/internal/db"
	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
)

// Server implements the UserService, OrderMongoService, and DroneMongoService gRPC services.
type Server struct {
	pb.UnimplementedUserServiceServer
	pb.UnimplementedOrderMongoServiceServer
	pb.UnimplementedDroneMongoServiceServer
	pb.UnimplementedFlightMongoServiceServer
	pb.UnimplementedOperationsMongoServiceServer
	pb.UnimplementedDeliveryMongoServiceServer
	pb.UnimplementedPaymentMongoServiceServer
	pb.UnimplementedWarehouseMongoServiceServer
	pb.UnimplementedAnalyticsMongoServiceServer
	pb.UnimplementedWorkflowMongoServiceServer
	db          *db.MongoDatabase
	redisClient redis_pb.RedisCacheServiceClient
}

// NewServer creates a new instance of the Server.
func NewServer(database *db.MongoDatabase, rClient redis_pb.RedisCacheServiceClient) *Server {
	return &Server{
		db:          database,
		redisClient: rClient,
	}
}
