package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	auth_pb "github.com/RevenueIQ/auth_service/auth_service"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authClient auth_pb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing Authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: malformed token (must be 'Bearer <token>')"})
			c.Abort()
			return
		}

		token := parts[1]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := authClient.VerifyToken(ctx, &auth_pb.VerifyTokenRequest{
			AccessToken: token,
		})
		if err != nil || !resp.GetSuccess() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid or expired token"})
			c.Abort()
			return
		}

		// Inject user details context into the Gin context
		c.Set("user_id", resp.GetUserId())
		c.Set("email", resp.GetEmail())
		c.Set("role", resp.GetRole())

		c.Next()
	}
}
