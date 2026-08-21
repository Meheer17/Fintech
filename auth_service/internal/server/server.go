package server

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/auth"
	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/namespaces"
	auth_pb "github.com/RevenueIQ/auth_service/auth_service"
	mongo_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"
	redis_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/redis_service"
	queue_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/queue_service"
	notification_pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/notification_service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Server implements the AuthService gRPC service.
type Server struct {
	auth_pb.UnimplementedAuthServiceServer
	mongoClient mongo_pb.UserServiceClient
	redisClient redis_pb.RedisCacheServiceClient
	queueClient queue_pb.QueueServiceClient
	jwtSecret   []byte
}

// NewServer creates a new instance of the Auth Server.
func NewServer(mClient mongo_pb.UserServiceClient, rClient redis_pb.RedisCacheServiceClient, qClient queue_pb.QueueServiceClient) *Server {
	secret := []byte(getEnv("JWT_SECRET", "super-secret-jwt-key"))
	return &Server{
		mongoClient: mClient,
		redisClient: rClient,
		queueClient: qClient,
		jwtSecret:   secret,
	}
}

// RegisterUser registers a user with the "User" role by caching their details and generating an OTP.
func (s *Server) RegisterUser(ctx context.Context, req *auth_pb.RegisterUserRequest) (*auth_pb.RegisterUserResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}
	if req.GetFirstName() == "" || req.GetLastName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "first_name and last_name are required")
	}

	existingUserResp, err := s.mongoClient.GetUserByEmail(ctx, &mongo_pb.GetUserByEmailRequest{Email: req.GetEmail()})
	if err == nil && existingUserResp.GetUser() != nil {
		return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", req.GetEmail())
	}

	// Construct pending user details
	pendingUser := &mongo_pb.User{
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		Email:       req.GetEmail(),
		Password:    req.GetPassword(),
		Role:        "User",
		PhoneNumber: req.GetPhoneNumber(),
		Age:         req.GetAge(),
		Active:      true,
	}

	anyVal, err := anypb.New(pendingUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal pending user: %v", err)
	}

	// Cache pending user details in Redis
	_, err = s.redisClient.Set(ctx, &redis_pb.SetRequest{
		Key:        "pending_user:" + req.GetEmail(),
		Value:      anyVal,
		TtlSeconds: 900, // 15 minutes TTL
		Namespace:  string(namespaces.Sessions),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cache pending user: %v", err)
	}

	// Generate and cache OTP
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	otpVal := &auth_pb.OTPCode{
		Code:        code,
		GeneratedAt: time.Now().Unix(),
	}

	anyOtp, err := anypb.New(otpVal)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal OTP: %v", err)
	}

	_, err = s.redisClient.Set(ctx, &redis_pb.SetRequest{
		Key:        "otp:" + req.GetEmail(),
		Value:      anyOtp,
		TtlSeconds: 300, // 5 minutes TTL
		Namespace:  string(namespaces.Sessions),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store OTP: %v", err)
	}

	// Send OTP email via Queue Service
	s.sendOTPEmail(ctx, req.GetEmail(), code, "Registration Signup")

	return &auth_pb.RegisterUserResponse{
		User: pendingUser,
	}, nil
}

// SendOTP generates a 6-digit OTP, stores it in Redis under the "sessions_ns" namespace, and mocks delivery.
func (s *Server) SendOTP(ctx context.Context, req *auth_pb.SendOTPRequest) (*auth_pb.SendOTPResponse, error) {
	target := req.GetEmail()
	if target == "" {
		target = req.GetPhoneNumber()
	}
	if target == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email or phone_number is required")
	}

	// Generate a 6-digit random OTP
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Create OTPCode payload
	otpVal := &auth_pb.OTPCode{
		Code:        code,
		GeneratedAt: time.Now().Unix(),
	}

	anyVal, err := anypb.New(otpVal)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal OTP: %v", err)
	}

	// Cache in Redis under namespaces.Sessions with a 5-minute TTL (300 seconds)
	redisKey := "otp:" + target
	_, err = s.redisClient.Set(ctx, &redis_pb.SetRequest{
		Key:        redisKey,
		Value:      anyVal,
		TtlSeconds: 300,
		Namespace:  string(namespaces.Sessions),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store OTP in cache: %v", err)
	}

	// Send OTP email via Queue Service if target is an email address
	if strings.Contains(target, "@") {
		s.sendOTPEmail(ctx, target, code, "Login Verification")
	}

	// Mock OTP delivery by printing to standard logger
	log.Printf("==========================================")
	log.Printf("[OTP SERVICE] Sent OTP Code: %s to %s", code, target)
	log.Printf("==========================================")

	return &auth_pb.SendOTPResponse{
		Success: true,
		Message: "OTP sent successfully",
	}, nil
}

// VerifyOTP validates the OTP from Redis and deletes it upon successful match.
func (s *Server) VerifyOTP(ctx context.Context, req *auth_pb.VerifyOTPRequest) (*auth_pb.VerifyOTPResponse, error) {
	target := req.GetEmail()
	if target == "" {
		target = req.GetPhoneNumber()
	}
	if target == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email or phone_number is required")
	}
	if req.GetCode() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "code is required")
	}

	redisKey := "otp:" + target

	// Get OTP envelope from Redis
	resp, err := s.redisClient.Get(ctx, &redis_pb.GetRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get OTP from cache: %v", err)
	}
	if !resp.GetFound() {
		return &auth_pb.VerifyOTPResponse{
			Success: false,
			Message: "OTP has expired or does not exist",
		}, nil
	}

	// Decode OTPCode
	var otpVal auth_pb.OTPCode
	err = resp.GetValue().UnmarshalTo(&otpVal)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal cached OTP: %v", err)
	}

	// Validate OTP
	if otpVal.GetCode() != req.GetCode() {
		return &auth_pb.VerifyOTPResponse{
			Success: false,
			Message: "Invalid OTP code",
		}, nil
	}

	// Delete key on success (single use OTP)
	_, _ = s.redisClient.Delete(ctx, &redis_pb.DeleteRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})

	// Check if there is a pending user for this target
	pendingUserKey := "pending_user:" + target
	pendingResp, err := s.redisClient.Get(ctx, &redis_pb.GetRequest{
		Key:       pendingUserKey,
		Namespace: string(namespaces.Sessions),
	})
	if err == nil && pendingResp.GetFound() {
		var pendingUser mongo_pb.User
		err = pendingResp.GetValue().UnmarshalTo(&pendingUser)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to unmarshal pending user: %v", err)
		}

		// Create user in MongoDB with verification flags set to true
		_, err = s.mongoClient.CreateUser(ctx, &mongo_pb.CreateUserRequest{
			FirstName:     pendingUser.GetFirstName(),
			LastName:      pendingUser.GetLastName(),
			Email:         pendingUser.GetEmail(),
			Password:      pendingUser.GetPassword(),
			Role:          pendingUser.GetRole(),
			PhoneNumber:   pendingUser.GetPhoneNumber(),
			Age:           pendingUser.GetAge(),
			EmailVerified: true,
			PhoneVerified: true,
			Active:        true,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create user in DB: %v", err)
		}

		// Generate signup email payload and publish to Queue
		dataID := fmt.Sprintf("signup_email_%s_%d", pendingUser.GetEmail(), time.Now().UnixNano())
		emailPayload := &notification_pb.EmailPayload{
			ToAddress: pendingUser.GetEmail(),
			Subject:   "Welcome to RevenueIQ!",
			BodyHtml:  fmt.Sprintf("<h1>Welcome %s %s!</h1><p>Your registration is complete.</p>", pendingUser.GetFirstName(), pendingUser.GetLastName()),
			BodyText:  fmt.Sprintf("Welcome %s %s! Your registration is complete.", pendingUser.GetFirstName(), pendingUser.GetLastName()),
		}

		anyEmail, err := anypb.New(emailPayload)
		if err != nil {
			log.Printf("Failed to marshal signup email payload: %v", err)
		} else {
			// Store email data in Redis under QueueData namespace
			_, err = s.redisClient.Set(ctx, &redis_pb.SetRequest{
				Key:        dataID,
				Value:      anyEmail,
				TtlSeconds: 3600, // 1 hour TTL
				Namespace:  string(namespaces.QueueData),
			})
			if err != nil {
				log.Printf("Failed to store signup email payload in Redis: %v", err)
			} else {
				// Publish to Queue Service via gRPC
				_, err = s.queueClient.PublishEvent(ctx, &queue_pb.PublishEventRequest{
					Topic:  "EMAIL",
					Type:   "signup",
					DataId: dataID,
				})
				if err != nil {
					log.Printf("Failed to publish signup email event to Queue Service: %v", err)
				} else {
					log.Printf("Successfully published signup email event to Queue Service for user %s", pendingUser.GetEmail())
				}
			}
		}

		// Clean up pending user from cache
		_, _ = s.redisClient.Delete(ctx, &redis_pb.DeleteRequest{
			Key:       pendingUserKey,
			Namespace: string(namespaces.Sessions),
		})

		return &auth_pb.VerifyOTPResponse{
			Success: true,
			Message: "User registered and verified successfully",
		}, nil
	}

	return &auth_pb.VerifyOTPResponse{
		Success: true,
		Message: "OTP verified successfully",
	}, nil
}

// Login verifies credentials, issues JWT tokens, and stores the refresh token in Redis.
func (s *Server) Login(ctx context.Context, req *auth_pb.LoginRequest) (*auth_pb.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}

	// Query user from mongodb_service
	userResp, err := s.mongoClient.GetUserByEmail(ctx, &mongo_pb.GetUserByEmailRequest{
		Email: req.GetEmail(),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return &auth_pb.LoginResponse{
				Success: false,
				Message: "Invalid credentials: user not found",
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to look up user in DB: %v", err)
	}

	user := userResp.GetUser()

	// Check password (plain text match for demonstration)
	if user.GetPassword() != req.GetPassword() {
		return &auth_pb.LoginResponse{
			Success: false,
			Message: "Invalid credentials: password incorrect",
		}, nil
	}

	// Check if user is active
	if !user.GetActive() {
		return &auth_pb.LoginResponse{
			Success: false,
			Message: "Invalid credentials: account is inactive",
		}, nil
	}

	// Update user's last login timestamp in MongoDB
	go func(uid string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = s.mongoClient.UpdateUser(bgCtx, &mongo_pb.UpdateUserRequest{
			Id:            uid,
			FirstName:     proto.String(user.GetFirstName()),
			LastName:      proto.String(user.GetLastName()),
			Email:         proto.String(user.GetEmail()),
			Age:           proto.Int32(user.GetAge()),
			Password:      proto.String(user.GetPassword()),
			Role:          proto.String(user.GetRole()),
			PhoneNumber:   proto.String(user.GetPhoneNumber()),
			EmailVerified: proto.Bool(user.GetEmailVerified()),
			PhoneVerified: proto.Bool(user.GetPhoneVerified()),
			Active:        proto.Bool(user.GetActive()),
		})
	}(user.GetId())

	// Generate Access Token (expires in 15 minutes)
	accessToken, err := auth.GenerateToken(user.GetId(), user.GetEmail(), user.GetRole(), 15*time.Minute, s.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	// Generate Refresh Token (expires in 7 days)
	refreshToken, err := auth.GenerateToken(user.GetId(), user.GetEmail(), user.GetRole(), 7*24*time.Hour, s.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}

	// Store Refresh Token in Redis under sessions_ns namespace with a 7-day TTL
	anyVal := wrapperspb.String(refreshToken)
	anyEnv, err := anypb.New(anyVal)
	if err == nil {
		_, _ = s.redisClient.Set(ctx, &redis_pb.SetRequest{
			Key:        "refresh:" + user.GetId(),
			Value:      anyEnv,
			TtlSeconds: int64(7 * 24 * 3600),
			Namespace:  string(namespaces.Sessions),
		})
	}

	// Store Access Token in Redis under sessions_ns namespace with a 15-minute TTL
	anyAccessVal := wrapperspb.String(accessToken)
	anyAccessEnv, err := anypb.New(anyAccessVal)
	if err == nil {
		_, _ = s.redisClient.Set(ctx, &redis_pb.SetRequest{
			Key:        "access:" + user.GetId(),
			Value:      anyAccessEnv,
			TtlSeconds: int64(15 * 60),
			Namespace:  string(namespaces.Sessions),
		})
	}

	return &auth_pb.LoginResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "Logged in successfully",
		User:         user,
	}, nil
}

// RefreshTokens validates the refresh token and issues a new access/refresh token pair.
func (s *Server) RefreshTokens(ctx context.Context, req *auth_pb.RefreshTokensRequest) (*auth_pb.RefreshTokensResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh_token is required")
	}

	// 1. Validate JWT structure and signature
	claims, err := auth.ValidateToken(req.GetRefreshToken(), s.jwtSecret)
	if err != nil {
		return &auth_pb.RefreshTokensResponse{
			Success: false,
		}, nil
	}

	// 2. Verify against Redis session
	redisKey := "refresh:" + claims.UserID
	cacheResp, err := s.redisClient.Get(ctx, &redis_pb.GetRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})
	if err != nil || !cacheResp.GetFound() {
		return &auth_pb.RefreshTokensResponse{
			Success: false,
		}, nil
	}

	var cachedToken wrapperspb.StringValue
	err = cacheResp.GetValue().UnmarshalTo(&cachedToken)
	if err != nil || cachedToken.GetValue() != req.GetRefreshToken() {
		return &auth_pb.RefreshTokensResponse{
			Success: false,
		}, nil
	}

	// 3. Query User details
	userResp, err := s.mongoClient.GetUser(ctx, &mongo_pb.GetUserRequest{Id: claims.UserID})
	if err != nil {
		return &auth_pb.RefreshTokensResponse{
			Success: false,
		}, nil
	}
	user := userResp.GetUser()

	// 4. Generate new tokens
	newAccessToken, err := auth.GenerateToken(user.GetId(), user.GetEmail(), user.GetRole(), 15*time.Minute, s.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	newRefreshToken, err := auth.GenerateToken(user.GetId(), user.GetEmail(), user.GetRole(), 7*24*time.Hour, s.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}

	// 5. Update cached refresh token in Redis
	anyVal := wrapperspb.String(newRefreshToken)
	anyEnv, err := anypb.New(anyVal)
	if err == nil {
		_, _ = s.redisClient.Set(ctx, &redis_pb.SetRequest{
			Key:        redisKey,
			Value:      anyEnv,
			TtlSeconds: int64(7 * 24 * 3600),
			Namespace:  string(namespaces.Sessions),
		})
	}

	// 6. Update cached access token in Redis with 15-minute TTL
	anyAccessVal := wrapperspb.String(newAccessToken)
	anyAccessEnv, err := anypb.New(anyAccessVal)
	if err == nil {
		_, _ = s.redisClient.Set(ctx, &redis_pb.SetRequest{
			Key:        "access:" + user.GetId(),
			Value:      anyAccessEnv,
			TtlSeconds: int64(15 * 60),
			Namespace:  string(namespaces.Sessions),
		})
	}

	return &auth_pb.RefreshTokensResponse{
		Success:      true,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// ForgotPassword generates a reset OTP code and caches it in Redis.
func (s *Server) ForgotPassword(ctx context.Context, req *auth_pb.ForgotPasswordRequest) (*auth_pb.ForgotPasswordResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}

	// Check if user exists in DB first
	_, err := s.mongoClient.GetUserByEmail(ctx, &mongo_pb.GetUserByEmailRequest{Email: req.GetEmail()})
	if err != nil {
		return &auth_pb.ForgotPasswordResponse{
			Success: false,
			Message: "Email address not found",
		}, nil
	}

	// Generate a 6-digit random Reset OTP
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Store OTP in Redis under namespaces.Sessions with a 5-minute TTL
	otpVal := &auth_pb.OTPCode{
		Code:        code,
		GeneratedAt: time.Now().Unix(),
	}

	anyVal, err := anypb.New(otpVal)
	if err == nil {
		redisKey := "reset_otp:" + req.GetEmail()
		_, _ = s.redisClient.Set(ctx, &redis_pb.SetRequest{
			Key:        redisKey,
			Value:      anyVal,
			TtlSeconds: 300,
			Namespace:  string(namespaces.Sessions),
		})
	}

	// Send OTP email via Queue Service
	s.sendOTPEmail(ctx, req.GetEmail(), code, "Password Reset")

	// Mock reset email delivery
	log.Printf("==========================================")
	log.Printf("[OTP SERVICE] Sent Reset Password OTP: %s to %s", code, req.GetEmail())
	log.Printf("==========================================")

	return &auth_pb.ForgotPasswordResponse{
		Success: true,
		Message: "Password reset OTP sent successfully",
	}, nil
}

// ResetPassword verifies the reset OTP and updates the user's password in MongoDB.
func (s *Server) ResetPassword(ctx context.Context, req *auth_pb.ResetPasswordRequest) (*auth_pb.ResetPasswordResponse, error) {
	if req.GetEmail() == "" || req.GetCode() == "" || req.GetNewPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email, code, and new_password are required")
	}

	redisKey := "reset_otp:" + req.GetEmail()

	// Get OTP from Redis
	cacheResp, err := s.redisClient.Get(ctx, &redis_pb.GetRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})
	if err != nil || !cacheResp.GetFound() {
		return &auth_pb.ResetPasswordResponse{
			Success: false,
			Message: "Reset OTP has expired or does not exist",
		}, nil
	}

	var otpVal auth_pb.OTPCode
	err = cacheResp.GetValue().UnmarshalTo(&otpVal)
	if err != nil || otpVal.GetCode() != req.GetCode() {
		return &auth_pb.ResetPasswordResponse{
			Success: false,
			Message: "Invalid or incorrect OTP code",
		}, nil
	}

	// OTP is correct! Fetch user
	userResp, err := s.mongoClient.GetUserByEmail(ctx, &mongo_pb.GetUserByEmailRequest{Email: req.GetEmail()})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	user := userResp.GetUser()

	// Update user password in DB
	_, err = s.mongoClient.UpdateUser(ctx, &mongo_pb.UpdateUserRequest{
		Id:            user.GetId(),
		FirstName:     proto.String(user.GetFirstName()),
		LastName:      proto.String(user.GetLastName()),
		Email:         proto.String(user.GetEmail()),
		Age:           proto.Int32(user.GetAge()),
		Password:      proto.String(req.GetNewPassword()), // Setting new password
		Role:          proto.String(user.GetRole()),
		PhoneNumber:   proto.String(user.GetPhoneNumber()),
		EmailVerified: proto.Bool(user.GetEmailVerified()),
		PhoneVerified: proto.Bool(user.GetPhoneVerified()),
		Active:        proto.Bool(user.GetActive()),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user password: %v", err)
	}

	// Delete reset OTP
	_, _ = s.redisClient.Delete(ctx, &redis_pb.DeleteRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})

	return &auth_pb.ResetPasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

// Logout deletes the cached refresh token from Redis.
func (s *Server) Logout(ctx context.Context, req *auth_pb.LogoutRequest) (*auth_pb.LogoutResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh_token is required")
	}

	// Validate token
	claims, err := auth.ValidateToken(req.GetRefreshToken(), s.jwtSecret)
	if err != nil {
		return &auth_pb.LogoutResponse{Success: false}, nil
	}

	// Delete from Redis
	redisKey := "refresh:" + claims.UserID
	_, _ = s.redisClient.Delete(ctx, &redis_pb.DeleteRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})

	redisAccessKey := "access:" + claims.UserID
	_, _ = s.redisClient.Delete(ctx, &redis_pb.DeleteRequest{
		Key:       redisAccessKey,
		Namespace: string(namespaces.Sessions),
	})

	return &auth_pb.LogoutResponse{
		Success: true,
	}, nil
}

// VerifyToken validates the access token, parses claims, and returns user details.
func (s *Server) VerifyToken(ctx context.Context, req *auth_pb.VerifyTokenRequest) (*auth_pb.VerifyTokenResponse, error) {
	if req.GetAccessToken() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "access_token is required")
	}

	claims, err := auth.ValidateToken(req.GetAccessToken(), s.jwtSecret)
	if err != nil {
		return &auth_pb.VerifyTokenResponse{
			Success: false,
		}, nil
	}

	// Verify active session against Redis
	redisKey := "access:" + claims.UserID
	cacheResp, err := s.redisClient.Get(ctx, &redis_pb.GetRequest{
		Key:       redisKey,
		Namespace: string(namespaces.Sessions),
	})
	if err != nil || !cacheResp.GetFound() {
		return &auth_pb.VerifyTokenResponse{
			Success: false,
		}, nil
	}

	var cachedAccess wrapperspb.StringValue
	err = cacheResp.GetValue().UnmarshalTo(&cachedAccess)
	if err != nil || cachedAccess.GetValue() != req.GetAccessToken() {
		return &auth_pb.VerifyTokenResponse{
			Success: false,
		}, nil
	}

	// Retrieve latest user details from mongodb_service
	userResp, err := s.mongoClient.GetUser(ctx, &mongo_pb.GetUserRequest{Id: claims.UserID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user details: %v", err)
	}

	return &auth_pb.VerifyTokenResponse{
		Success: true,
		UserId:  claims.UserID,
		Email:   claims.Email,
		Role:    claims.Role,
		User:    userResp.GetUser(),
	}, nil
}

// sendOTPEmail constructs an OTP email payload, caches it in Redis, and publishes the event to the queue.
func (s *Server) sendOTPEmail(ctx context.Context, toEmail, code, reason string) {
	if toEmail == "" {
		return
	}

	dataID := fmt.Sprintf("otp_email_%s_%d", toEmail, time.Now().UnixNano())
	subject := fmt.Sprintf("Your OTP Code for %s", reason)
	bodyHtml := fmt.Sprintf(
		"<h1>Security Verification</h1><p>You requested a verification code for <strong>%s</strong>.</p><p>Your OTP code is: <strong style='font-size: 24px; letter-spacing: 2px; color: #4F46E5;'>%s</strong></p><p>This code will expire in 5 minutes.</p>",
		reason, code,
	)
	bodyText := fmt.Sprintf(
		"Security Verification\n\nYou requested a verification code for %s.\n\nYour OTP code is: %s\n\nThis code will expire in 5 minutes.",
		reason, code,
	)

	emailPayload := &notification_pb.EmailPayload{
		ToAddress: toEmail,
		Subject:   subject,
		BodyHtml:  bodyHtml,
		BodyText:  bodyText,
	}

	anyEmail, err := anypb.New(emailPayload)
	if err != nil {
		log.Printf("Failed to marshal OTP email payload: %v", err)
		return
	}

	// Store in Redis under namespaces.QueueData
	_, err = s.redisClient.Set(ctx, &redis_pb.SetRequest{
		Key:        dataID,
		Value:      anyEmail,
		TtlSeconds: 3600, // 1 hour TTL
		Namespace:  string(namespaces.QueueData),
	})
	if err != nil {
		log.Printf("Failed to store OTP email payload in Redis: %v", err)
		return
	}

	// Publish to Queue Service
	_, err = s.queueClient.PublishEvent(ctx, &queue_pb.PublishEventRequest{
		Topic:  "EMAIL",
		Type:   "otp",
		DataId: dataID,
	})
	if err != nil {
		log.Printf("Failed to publish OTP email event to Queue Service: %v", err)
		return
	}

	log.Printf("Successfully published OTP email event (%s) to Queue Service for %s", reason, toEmail)
}

