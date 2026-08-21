package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	auth_pb "github.com/RevenueIQ/auth_service/auth_service"
	"github.com/RevenueIQ/revenueiq_dev_kit/pkg/errors"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authClient auth_pb.AuthServiceClient
}

func NewAuthHandler(authClient auth_pb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

// Helper to validate email format
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// Health Check
func (h *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// User Sign Up
func (h *AuthHandler) Signup(c *gin.Context) {
	var req auth_pb.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	if strings.TrimSpace(req.GetEmail()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if !isValidEmail(req.GetEmail()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	if len(req.GetPassword()) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long"})
		return
	}
	if strings.TrimSpace(req.GetFirstName()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "First name is required"})
		return
	}
	if strings.TrimSpace(req.GetLastName()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Last name is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.RegisterUser(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Send OTP
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req auth_pb.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	email := strings.TrimSpace(req.GetEmail())
	phone := strings.TrimSpace(req.GetPhoneNumber())
	if email == "" && phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either email or phone number is required"})
		return
	}
	if email != "" && !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.SendOTP(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Verify OTP
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req auth_pb.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	email := strings.TrimSpace(req.GetEmail())
	phone := strings.TrimSpace(req.GetPhoneNumber())
	code := strings.TrimSpace(req.GetCode())
	if email == "" && phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either email or phone number is required"})
		return
	}
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP code is required"})
		return
	}
	if len(code) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP code must be 6 digits"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.VerifyOTP(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth_pb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	if strings.TrimSpace(req.GetEmail()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if !isValidEmail(req.GetEmail()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	if req.GetPassword() == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Login(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Refresh Tokens
func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	var req auth_pb.RefreshTokensRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	if strings.TrimSpace(req.GetRefreshToken()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.RefreshTokens(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Forgot Password Send
func (h *AuthHandler) ForgotPasswordSend(c *gin.Context) {
	var req auth_pb.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	if strings.TrimSpace(req.GetEmail()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if !isValidEmail(req.GetEmail()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.ForgotPassword(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Forgot Password Reset
func (h *AuthHandler) ForgotPasswordReset(c *gin.Context) {
	var req auth_pb.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	if strings.TrimSpace(req.GetEmail()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if !isValidEmail(req.GetEmail()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	if strings.TrimSpace(req.GetCode()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset code is required"})
		return
	}
	if strings.TrimSpace(req.GetNewPassword()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password is required"})
		return
	}
	if len(req.GetNewPassword()) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 6 characters long"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.ResetPassword(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout (Protected)
func (h *AuthHandler) Logout(c *gin.Context) {
	var req auth_pb.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Prevalidation
	if strings.TrimSpace(req.GetRefreshToken()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Logout(ctx, &req)
	if err != nil {
		status, msg := errors.MapGRPCErrToHTTPStatus(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Verify Token (Protected)
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	c.Header("X-User-Id", fmt.Sprintf("%v", userID))
	c.Header("X-Email", fmt.Sprintf("%v", email))
	c.Header("X-Role", fmt.Sprintf("%v", role))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"user_id": userID,
			"email":   email,
			"role":    role,
		},
		"message": "Token is valid and active",
	})
}

// Me (Protected)
func (h *AuthHandler) Me(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing Authorization header"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: malformed token"})
		return
	}

	token := parts[1]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.VerifyToken(ctx, &auth_pb.VerifyTokenRequest{
		AccessToken: token,
	})
	if err != nil || !resp.GetSuccess() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    resp.GetUser(),
	})
}
