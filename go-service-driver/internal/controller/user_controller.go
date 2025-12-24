package controller

import (
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
	rateKey := "verify:limit:" + req.Email

	exists, _ := database.RedisClient.Exists(c, rateKey).Result()
	if exists > 0 {
		middleware.HandleError(c, errors.NewRateLimitExceededError())
		return
	}

	if err := uc.userService.SendVerificationCode(c, req.Email); err != nil {
		middleware.HandleError(c, errors.NewInternalServerError(err))
		return
	}

	// Set rate limit key with 1 minute expiration
	_ = database.RedisClient.Set(c, rateKey, "1", 1*time.Minute).Err()

	response.SuccessWithMessage(c, nil, "Verification code sent")
}

// POST /api/auth/register
func (uc *UserController) RegisterUser(c *gin.Context) {
	var req dto.RegisterUser
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError(err.Error()))
		return
	}

	user, err := uc.userService.RegisterUser(c, req)
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

	user, err := uc.userService.LoginUser(c, req)
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

	user, err := uc.userService.LoginUserByEmail(c, req)
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

	if err := uc.userService.SendForgetPasswordCode(c, req.Email); err != nil {
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

	if err := uc.userService.SendForgetPasswordCodeByUsername(c, req.Username); err != nil {
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

	if err := uc.userService.VerifyForgetPassword(c, req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Password reset successfully")
}

// GET /api/v2/users/profile
func (uc *UserController) GetUserProfile(c *gin.Context) {
	userID := c.GetUint64("userID")

	profile, err := uc.userService.GetUserProfile(c, userID)
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

	profile, err := uc.userService.UpdateUserProfile(c, userID, req)
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

	if err := uc.userService.ChangePassword(c, userID, req); err != nil {
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

	if err := uc.userService.DeleteAccount(c, userID, req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "Account deleted successfully")
}

// POST /api/v2/users/upload-avatar
func (uc *UserController) UploadAvatar(c *gin.Context) {
	userID := c.GetUint64("userID")

	file, err := c.FormFile("avatar")
	if err != nil {
		middleware.HandleError(c, errors.NewInvalidRequestError("Avatar file is required"))
		return
	}

	avatarURL, err := uc.userService.UploadAvatar(c, userID, file)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"avatar_url": avatarURL})
}
