package controller

import (
	"context"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/middleware"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/response"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/service"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
	"github.com/gin-gonic/gin"
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
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	// Rate limiting: allow one request per minute per email
	ctx := context.Background()
	rateKey := "verify:limit:" + req.Email

	exists, _ := database.RedisClient.Exists(ctx, rateKey).Result()
	if exists > 0 {
		middleware.HandleError(c, errors.NewRateLimitExceededError())
		return
	}

	if err := uc.userService.SendVerificationCode(ctx, req.Email); err != nil {
		middleware.HandleError(c, errors.NewInternalServerError(err))
		return
	}

	// Set rate limit key with 1 minute expiration
	_ = database.RedisClient.Set(ctx, rateKey, "1", 1*time.Minute).Err()

	response.SuccessWithMessage(c, nil, "Verification code sent")
}

// POST /api/auth/register
func (uc *UserController) RegisterUser(c *gin.Context) {
	var req dto.RegisterUser
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	user, err := uc.userService.RegisterUser(ctx, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		middleware.HandleError(c, errors.NewInternalServerError(err))
		return
	}

	response.Created(c, dto.RegisterResponse{
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
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	user, err := uc.userService.LoginUser(ctx, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.Success(c, user)
}

// POST /api/auth/login-email
func (uc *UserController) LoginUserByEmail(c *gin.Context) {
	var req dto.EmailLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	user, err := uc.userService.LoginUserByEmail(ctx, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.Success(c, user)
}

// POST /api/auth/send-forget-code
func (uc *UserController) SendForgetPasswordCode(c *gin.Context) {
	var req dto.ForgetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	if err := uc.userService.SendForgetPasswordCode(ctx, req.Email); err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Forget password code sent")
}

// POST /api/auth/send-forget-code-by-username
func (uc *UserController) SendForgetPasswordCodeByUsername(c *gin.Context) {
	var req dto.ForgetPasswordByUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	if err := uc.userService.SendForgetPasswordCodeByUsername(ctx, req.Username); err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Forget password code sent")
}

// POST /api/auth/verify-forget-password
func (uc *UserController) VerifyForgetPassword(c *gin.Context) {
	var req dto.ForgetPasswordVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	if err := uc.userService.VerifyForgetPassword(ctx, req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Password reset successfully")
}

// GET /api/v2/users/profile
func (uc *UserController) GetUserProfile(c *gin.Context) {
	userID := c.GetUint64("userID")

	ctx := context.Background()
	profile, err := uc.userService.GetUserProfile(ctx, userID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.Success(c, profile)
}

// PUT /api/v2/users/profile
func (uc *UserController) UpdateUserProfile(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.UpdateUserProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	profile, err := uc.userService.UpdateUserProfile(ctx, userID, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.Success(c, profile)
}

// PUT /api/v2/users/change-password
func (uc *UserController) ChangePassword(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	if err := uc.userService.ChangePassword(ctx, userID, req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Password changed successfully")
}

// DELETE /api/v2/users/delete-account
func (uc *UserController) DeleteAccount(c *gin.Context) {
	userID := c.GetUint64("userID")

	var req dto.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	ctx := context.Background()
	if err := uc.userService.DeleteAccount(ctx, userID, req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Account deleted successfully")
}
