package server

import (
	"context"
	"strings"

	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/auth"
	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/namespaces"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// Check implements the Envoy ext_authz gRPC interface
func (s *Server) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
	headers := req.GetAttributes().GetRequest().GetHttp().GetHeaders()

	// Extract Authorization header (case-insensitive)
	var authHeader string
	for k, v := range headers {
		if strings.ToLower(k) == "authorization" {
			authHeader = v
			break
		}
	}

	if authHeader == "" {
		return s.deniedResponse(int32(typev3.StatusCode_Unauthorized), "Authorization header is required"), nil
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return s.deniedResponse(int32(typev3.StatusCode_Unauthorized), "Authorization header must be Bearer token"), nil
	}
	tokenStr := parts[1]

	// Validate token
	claims, err := auth.ValidateToken(tokenStr, s.jwtSecret)
	if err != nil {
		return s.deniedResponse(int32(typev3.StatusCode_Unauthorized), "Invalid or expired token"), nil
	}

	// Verify active session against Redis
	redisKey := "access:" + claims.UserID
	cacheResp, err := s.redisClient.Get(ctx, &redis_pb.GetRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})
	if err != nil || !cacheResp.GetFound() {
		return s.deniedResponse(int32(typev3.StatusCode_Unauthorized), "Session expired or logged out"), nil
	}

	var cachedAccess wrapperspb.StringValue
	err = cacheResp.GetValue().UnmarshalTo(&cachedAccess)
	if err != nil || cachedAccess.GetValue() != tokenStr {
		return s.deniedResponse(int32(typev3.StatusCode_Unauthorized), "Token blacklisted"), nil
	}

	// OK response - inject user identity headers
	return &pb.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.OK),
		},
		HttpResponse: &pb.CheckResponse_OkResponse{
			OkResponse: &pb.OkHttpResponse{
				Headers: []*corev3.HeaderValueOption{
					{
						Header: &corev3.HeaderValue{
							Key:   "x-user-id",
							Value: claims.UserID,
						},
					},
					{
						Header: &corev3.HeaderValue{
							Key:   "x-email",
							Value: claims.Email,
						},
					},
					{
						Header: &corev3.HeaderValue{
							Key:   "x-role",
							Value: claims.Role,
						},
					},
				},
			},
		},
	}, nil
}

func (s *Server) deniedResponse(statusCode int32, message string) *pb.CheckResponse {
	return &pb.CheckResponse{
		Status: &status.Status{
			Code:    int32(codes.PermissionDenied),
			Message: message,
		},
		HttpResponse: &pb.CheckResponse_DeniedResponse{
			DeniedResponse: &pb.DeniedHttpResponse{
				Status: &typev3.HttpStatus{
					Code: typev3.StatusCode(statusCode),
				},
				Body: message,
			},
		},
	}
}
