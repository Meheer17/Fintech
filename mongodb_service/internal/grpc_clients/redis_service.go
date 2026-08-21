package grpcclients

import (
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewRedisCacheClient dials the Redis gRPC Cache Service and returns the client instance and client connection.
func NewRedisCacheClient(addr string) (redis_pb.RedisCacheServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	client := redis_pb.NewRedisCacheServiceClient(conn)
	return client, conn, nil
}
