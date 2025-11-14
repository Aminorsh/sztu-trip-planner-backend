package controller

import (
	"context"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(s *service.UserService) *UserController {
	return &UserController{
		userService: s,
	}
}

// POST /api/auth/send-code
func (uc *UserController) SendVerificationCode(c *gin.Context) {
	var req dto.SendVerificationEmail
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Rate limiting: allow one request per minute per email
	ctx := context.Background()
	rateKey := "verify:limit:" + req.Email

	exists, _ := database.RedisClient.Exists(ctx, rateKey).Result()
	if exists > 0 {
		c.JSON(429, gin.H{"error": "Too many requests. Please try again later."})
		return
	}

	if err := uc.userService.SendVerificationCode(ctx, req.Email); err != nil {
		c.JSON(500, gin.H{"error": "Failed to send verification code"})
		return
	}

	// Set rate limit key with 1 minute expiration
	_ = database.RedisClient.Set(ctx, rateKey, "1", 1*time.Minute).Err()

	c.JSON(200, gin.H{"message": "Verification code sent"})
}

// POST /api/auth/register
func (uc *UserController) RegisterUser(c *gin.Context) {
	var req dto.RegisterUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	ctx := context.Background()
	user, err := uc.userService.RegisterUser(ctx, req)
	if err != nil {
		if err == redis.Nil {
			c.JSON(400, gin.H{"error": "Verification code expired or not found"})
			return
		}

		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateToken(uint64(user.ID))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(201, dto.RegisterResponse{
		Message: "User registered successfully",
		User: dto.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		},
		Token: token,
	})
}

// POST /api/auth/login
func (uc *UserController) LoginUser(c *gin.Context) {
	var req dto.UsernameLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	ctx := context.Background()
	user, err := uc.userService.LoginUser(ctx, req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, dto.UserLoginResponse{
		Token: user.Token,
	})
}

// POST /api/auth/login-email
func (uc *UserController) LoginUserByEmail(c *gin.Context) {
	var req dto.EmailLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	ctx := context.Background()
	user, err := uc.userService.LoginUserByEmail(ctx, req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, dto.UserLoginResponse{
		Token: user.Token,
	})
}

// POST /api/auth/send-forget-code
func (uc *UserController) SendForgetPasswordCode(c *gin.Context) {
	var req dto.ForgetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	ctx := context.Background()
	if err := uc.userService.SendForgetPasswordCode(ctx, req.Email); err != nil {
		c.JSON(500, gin.H{"error": "Failed to send forget password code"})
		return
	}

	c.JSON(200, dto.ForgetPasswordResponse{
		Message: "Forget password code sent",
	})
}

// POST /api/auth/send-forget-code-by-username
func (uc *UserController) SendForgetPasswordCodeByUsername(c *gin.Context) {
	var req dto.ForgetPasswordByUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	ctx := context.Background()
	if err := uc.userService.SendForgetPasswordCodeByUsername(ctx, req.Username); err != nil {
		c.JSON(500, gin.H{"error": "Failed to send forget password code"})
		return
	}

	c.JSON(200, dto.ForgetPasswordResponse{
		Message: "Forget password code sent",
	})
}

// POST /api/auth/verify-forget-password
func (uc *UserController) VerifyForgetPassword(c *gin.Context) {
	var req dto.ForgetPasswordVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	ctx := context.Background()
	if err := uc.userService.VerifyForgetPassword(ctx, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, dto.ForgetPasswordResponse{
		Message: "Password reset successfully",
	})
}
