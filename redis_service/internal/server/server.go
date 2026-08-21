package server

import (
	"context"
	"time"

	"github.com/RevenueIQ/redis_service/internal/db"
	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/namespaces"
	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Server implements the RedisCacheService gRPC service.
type Server struct {
	pb.UnimplementedRedisCacheServiceServer
	db *db.RedisDatabase
}

// NewServer creates a new instance of the Server.
func NewServer(database *db.RedisDatabase) *Server {
	return &Server{db: database}
}

// validateNamespace checks if the provided namespace matches the revenueiq_dev_kit registry.
func (s *Server) validateNamespace(nsStr string) error {
	ns := namespaces.CacheNamespace(nsStr)
	if !ns.IsValid() {
		return status.Errorf(codes.InvalidArgument, "invalid cache namespace: %q. Must be a registered namespace from revenueiq_dev_kit", nsStr)
	}
	return nil
}

// Set handles storing data wrapped in a CacheItem envelope in Redis under a namespace.
func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if req.GetKey() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key is required")
	}
	if req.GetValue() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "value (Any) is required")
	}
	if err := s.validateNamespace(req.GetNamespace()); err != nil {
		return nil, err
	}

	namespace := req.GetNamespace()
	now := time.Now().Unix()
	var expiresAt int64

	// Setup TTL
	var ttl time.Duration
	if req.GetTtlSeconds() > 0 {
		ttl = time.Duration(req.GetTtlSeconds()) * time.Second
		expiresAt = now + req.GetTtlSeconds()
	}

	// Wrap the cached data in a CacheItem envelope with metadata
	envelope := &pb.CacheItem{
		Value:     req.GetValue(),
		Namespace: namespace,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	// Serialize the envelope
	data, err := proto.Marshal(envelope)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal cache envelope: %v", err)
	}

	namespacedKey := db.GetNamespacedKey(namespace, req.GetKey())
	err = s.db.Set(ctx, namespacedKey, data, ttl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write to Redis: %v", err)
	}

	return &pb.SetResponse{
		Success: true,
	}, nil
}

// Get handles retrieving the CacheItem from Redis and returning the value and metadata.
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.GetKey() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key is required")
	}
	if err := s.validateNamespace(req.GetNamespace()); err != nil {
		return nil, err
	}

	namespacedKey := db.GetNamespacedKey(req.GetNamespace(), req.GetKey())
	data, found, err := s.db.Get(ctx, namespacedKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read from Redis: %v", err)
	}
	if !found {
		return &pb.GetResponse{
			Found: false,
		}, nil
	}

	// Deserialize the CacheItem envelope
	var envelope pb.CacheItem
	err = proto.Unmarshal(data, &envelope)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal cache envelope: %v", err)
	}

	return &pb.GetResponse{
		Value: envelope.GetValue(),
		Found: true,
		Item:  &envelope,
	}, nil
}

// Delete handles removing a namespaced key from the Redis cache.
func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if req.GetKey() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key is required")
	}
	if err := s.validateNamespace(req.GetNamespace()); err != nil {
		return nil, err
	}

	namespacedKey := db.GetNamespacedKey(req.GetNamespace(), req.GetKey())
	success, err := s.db.Delete(ctx, namespacedKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete from Redis: %v", err)
	}

	return &pb.DeleteResponse{
		Success: success,
	}, nil
}

// InvalidateNamespace clears all keys starting with the namespace pattern.
func (s *Server) InvalidateNamespace(ctx context.Context, req *pb.InvalidateNamespaceRequest) (*pb.InvalidateNamespaceResponse, error) {
	if err := s.validateNamespace(req.GetNamespace()); err != nil {
		return nil, err
	}
	
	namespace := req.GetNamespace()
	
	count, err := s.db.InvalidateNamespace(ctx, namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to invalidate namespace %q: %v", namespace, err)
	}

	return &pb.InvalidateNamespaceResponse{
		InvalidatedCount: count,
	}, nil
}
