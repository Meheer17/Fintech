package grpcservers

import (
	"net"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
	"github.com/RevenueIQ/mongo_service/internal/db"
	"github.com/RevenueIQ/mongo_service/internal/server"

	"google.golang.org/grpc"
)

// StartGRPCServer sets up a TCP listener on the configured port, creates a new gRPC server,
// registers the User service, and returns the server instance and the listener.
func StartGRPCServer(port string, mongoDB *db.MongoDatabase, redisClient redis_pb.RedisCacheServiceClient) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, nil, err
	}

	grpcServer := grpc.NewServer()
	serviceServer := server.NewServer(mongoDB, redisClient)

	// Register Services
	pb.RegisterUserServiceServer(grpcServer, serviceServer)
	pb.RegisterOrderMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterDroneMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterFlightMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterOperationsMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterDeliveryMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterPaymentMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterWarehouseMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterAnalyticsMongoServiceServer(grpcServer, serviceServer)
	pb.RegisterWorkflowMongoServiceServer(grpcServer, serviceServer)

	return grpcServer, lis, nil
}
