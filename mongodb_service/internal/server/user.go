package server

import (
	"context"
	"errors"
	"time"

	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/models"
	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/namespaces"
	protohelpers "github.com/RevenueIQ/revenueiq_dev_kit/pkg/proto_helpers"
	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// CreateUser handles insertion of a new user.
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	doc := &models.UserDoc{
		FirstName:     req.GetFirstName(),
		LastName:      req.GetLastName(),
		Email:         req.GetEmail(),
		Age:           req.GetAge(),
		Password:      req.GetPassword(),
		Role:          req.GetRole(),
		PhoneNumber:   req.GetPhoneNumber(),
		EmailVerified: req.GetEmailVerified(),
		PhoneVerified: req.GetPhoneVerified(),
		Active:        req.GetActive(),
	}

	result, err := s.db.CreateUser(ctx, doc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	protoUser := protohelpers.ToProtoUser(result)

	if s.redisClient != nil {
		go func(userProto *pb.User) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			anyVal, err := anypb.New(userProto)
			if err == nil {
				_, _ = s.redisClient.Set(bgCtx, &redis_pb.SetRequest{
					Key:        userProto.GetId(),
					Value:      anyVal,
					TtlSeconds: 120,
					Namespace:  string(namespaces.Users),
				})
			}
		}(protoUser)
	}

	return &pb.CreateUserResponse{
		User: protoUser,
	}, nil
}

// GetUser handles retrieving a user by ID.
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	if s.redisClient != nil {
		cacheResp, err := s.redisClient.Get(ctx, &redis_pb.GetRequest{
			Key:       req.GetId(),
			Namespace: string(namespaces.Users),
		})
		if err == nil && cacheResp.GetFound() {
			var cachedUser pb.User
			err = cacheResp.GetValue().UnmarshalTo(&cachedUser)
			if err == nil {
				return &pb.GetUserResponse{
					User: &cachedUser,
				}, nil
			}
		}
	}

	result, err := s.db.GetUser(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	protoUser := protohelpers.ToProtoUser(result)

	if s.redisClient != nil {
		go func(userProto *pb.User) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			anyVal, err := anypb.New(userProto)
			if err == nil {
				_, _ = s.redisClient.Set(bgCtx, &redis_pb.SetRequest{
					Key:        userProto.GetId(),
					Value:      anyVal,
					TtlSeconds: 120,
					Namespace:  string(namespaces.Users),
				})
			}
		}(protoUser)
	}

	return &pb.GetUserResponse{
		User: protoUser,
	}, nil
}

// GetUserByEmail handles retrieving a user by email.
func (s *Server) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserByEmailResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}

	result, err := s.db.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user by email: %v", err)
	}

	return &pb.GetUserByEmailResponse{
		User: protohelpers.ToProtoUser(result),
	}, nil
}

// UpdateUser handles updating user fields.
func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	updateFields := make(map[string]interface{})
	if req.FirstName != nil {
		updateFields["first_name"] = req.GetFirstName()
	}
	if req.LastName != nil {
		updateFields["last_name"] = req.GetLastName()
	}
	if req.Email != nil {
		updateFields["email"] = req.GetEmail()
	}
	if req.Age != nil {
		updateFields["age"] = req.GetAge()
	}
	if req.Password != nil {
		updateFields["password"] = req.GetPassword()
	}
	if req.Role != nil {
		updateFields["role"] = req.GetRole()
	}
	if req.PhoneNumber != nil {
		updateFields["phone_number"] = req.GetPhoneNumber()
	}
	if req.EmailVerified != nil {
		updateFields["email_verified"] = req.GetEmailVerified()
	}
	if req.PhoneVerified != nil {
		updateFields["phone_verified"] = req.GetPhoneVerified()
	}
	if req.Active != nil {
		updateFields["active"] = req.GetActive()
	}

	result, err := s.db.UpdateUser(ctx, req.GetId(), updateFields)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	protoUser := protohelpers.ToProtoUser(result)

	if s.redisClient != nil {
		_, _ = s.redisClient.Delete(ctx, &redis_pb.DeleteRequest{
			Key:       protoUser.GetId(),
			Namespace: string(namespaces.Users),
		})
	}

	return &pb.UpdateUserResponse{
		User: protoUser,
	}, nil
}

// DeleteUser handles user deletion by ID.
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user ID is required")
	}

	success, err := s.db.DeleteUser(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	if s.redisClient != nil {
		_, _ = s.redisClient.Delete(ctx, &redis_pb.DeleteRequest{
			Key:       req.GetId(),
			Namespace: string(namespaces.Users),
		})
	}

	return &pb.DeleteUserResponse{
		Success: success,
	}, nil
}

// QueryUsers queries users supporting pagination, regex search, active flags, and dynamic map filters.
func (s *Server) QueryUsers(ctx context.Context, req *pb.QueryUsersRequest) (*pb.QueryUsersResponse, error) {
	docs, totalCount, totalPages, err := s.db.QueryUsers(ctx, req.GetPage(), req.GetLimit(), req.GetSearch(), req.GetActive(), req.GetFilters())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query users: %v", err)
	}

	var protoUsers []*pb.User
	for _, doc := range docs {
		protoUsers = append(protoUsers, protohelpers.ToProtoUser(doc))
	}

	return &pb.QueryUsersResponse{
		Users:       protoUsers,
		TotalCount:  int32(totalCount),
		TotalPages:  totalPages,
		CurrentPage: req.GetPage(),
	}, nil
}
