package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	// UserClaimsKey is the context key for storing authenticated UserClaims.
	UserClaimsKey contextKey = "user_claims"
)

// UserClaims defines the JWT claims payload structure.
type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates an access or refresh token signed with HS256.
func GenerateToken(userID, email, role string, duration time.Duration, secret []byte) (string, error) {
	claims := &UserClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateToken parses and validates a HS256-signed JWT token string.
func ValidateToken(tokenStr string, secret []byte) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// AppendJWTToContext appends a JWT to gRPC outgoing metadata.
func AppendJWTToContext(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

// ExtractJWTFromContext extracts a Bearer token from incoming gRPC metadata.
func ExtractJWTFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata not found in context")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.New("authorization header not found")
	}

	parts := strings.Split(authHeaders[0], " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header format must be 'Bearer <token>'")
	}

	return parts[1], nil
}

// InjectClaimsToContext puts the parsed claims in the context.
func InjectClaimsToContext(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, UserClaimsKey, claims)
}

// GetClaimsFromContext extracts claims from the context.
func GetClaimsFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	return claims, ok
}

// UnaryServerInterceptor validates the JWT in metadata and injects UserClaims into the context.
func UnaryServerInterceptor(secret []byte) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// You can whitelist health checks or non-protected methods here if needed
		if strings.Contains(info.FullMethod, "/health") {
			return handler(ctx, req)
		}

		token, err := ExtractJWTFromContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing or malformed auth token: %v", err)
		}

		claims, err := ValidateToken(token, secret)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Inject claims into handler context
		newCtx := InjectClaimsToContext(ctx, claims)
		return handler(newCtx, req)
	}
}
