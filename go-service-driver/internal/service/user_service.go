package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	apperrors "github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) SendVerificationCode(ctx context.Context, email string) error {
	// Generate verification code
	code := utils.GenerateVerificationCode()

	// Store code in Redis with expiration
	key := fmt.Sprintf("verify:%s", email)
	if err := database.RedisClient.Set(ctx, key, code, 15*time.Minute).Err(); err != nil {
		return err
	}

	// Send verification email
	if err := utils.NewEmailService().SendEmail(email, code, 0); err != nil {
		_ = database.RedisClient.Del(ctx, key).Err()
		return err
	}

	return nil
}

func (s *UserService) RegisterUser(ctx context.Context, req dto.RegisterUser) (*model.User, error) {
	if !config.IsTestMode() {
		// Verify code from Redis
		key := fmt.Sprintf("verify:%s", req.Email)
		storedCode, err := database.RedisClient.Get(ctx, key).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, apperrors.NewInternalServerError(err)
			}
			return nil, apperrors.NewVerifyCodeExpiredError()
		}
		if storedCode != req.Code {
			return nil, apperrors.NewInvalidVerifyCodeError()
		}

	}

	// Check if user already exists (repository handles soft-delete check)
	existingUser, err := s.userRepo.FindByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, apperrors.NewUserAlreadyExistsError("email or username")
	}

	// Create new user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}

	defaultAvatar := fmt.Sprintf("https://api.dicebear.com/6.x/initials/svg?seed=%s", req.Username)

	newUser := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  req.DisplayName,
		AvatarURL:    defaultAvatar,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Use repository to create (handles soft-delete restoration automatically)
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	if !config.IsTestMode() {
		// Delete verification code from Redis
		_ = database.RedisClient.Del(ctx, fmt.Sprintf("verify:%s", req.Email)).Err()
	}

	// Zero out password hash before returning
	newUser.PasswordHash = ""

	return newUser, nil
}

func (s *UserService) LoginUser(ctx context.Context, req dto.UsernameLogin) (*dto.UserLoginResponse, error) {
	// Find user by username using repository
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		// Repository returns UserNotFoundError, convert to InvalidCredentials for security
		if errors.Is(err, apperrors.NewUserNotFoundError()) {
			return nil, apperrors.NewInvalidCredentialsError()
		}
		return nil, err
	}

	if user.Status == "suspended" {
		return nil, apperrors.NewAccountSuspendedError()
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.NewInvalidCredentialsError()
	}

	// Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, uint(user.ID))

	// Generate JWT token
	token, err := utils.GenerateToken(uint64(user.ID))
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}

	return &dto.UserLoginResponse{
		Token: token,
	}, nil
}

func (s *UserService) LoginUserByEmail(ctx context.Context, req dto.EmailLogin) (*dto.UserLoginResponse, error) {
	// Find user by email using repository
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Repository returns UserNotFoundError, convert to InvalidCredentials for security
		if errors.Is(err, apperrors.NewUserNotFoundError()) {
			return nil, apperrors.NewInvalidCredentialsError()
		}
		return nil, err
	}

	if user.Status == "suspended" {
		return nil, apperrors.NewAccountSuspendedError()
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.NewInvalidCredentialsError()
	}

	// Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, uint(user.ID))

	// Generate JWT token
	token, err := utils.GenerateToken(uint64(user.ID))
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}

	return &dto.UserLoginResponse{
		Token: token,
	}, nil
}

func (s *UserService) SendForgetPasswordCode(ctx context.Context, email string) error {
	// Check if user exists using repository
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.Status == "suspended" {
		return apperrors.NewAccountSuspendedError()
	}

	// Generate verification code
	code := utils.GenerateVerificationCode()

	// Store code in Redis with expiration
	key := fmt.Sprintf("forget_password:%s", email)
	if err := database.RedisClient.Set(ctx, key, code, 15*time.Minute).Err(); err != nil {
		return apperrors.NewRedisError(err)
	}

	// Send verification email
	if err := utils.NewEmailService().SendEmail(email, code, 1); err != nil {
		_ = database.RedisClient.Del(ctx, key).Err()
		return apperrors.NewInternalServerError(err)
	}

	return nil
}

func (s *UserService) SendForgetPasswordCodeByUsername(ctx context.Context, username string) error {
	// Find user by username using repository
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}

	if user.Status == "suspended" {
		return apperrors.NewAccountSuspendedError()
	}

	// Generate verification code
	code := utils.GenerateVerificationCode()

	// Store code in Redis with expiration
	key := fmt.Sprintf("forget_password:%s", user.Email)
	if err := database.RedisClient.Set(ctx, key, code, 15*time.Minute).Err(); err != nil {
		return apperrors.NewRedisError(err)
	}

	// Send verification email
	if err := utils.NewEmailService().SendEmail(user.Email, code, 1); err != nil {
		_ = database.RedisClient.Del(ctx, key).Err()
		return apperrors.NewInternalServerError(err)
	}

	return nil
}

func (s *UserService) VerifyForgetPassword(ctx context.Context, req dto.ForgetPasswordVerifyRequest) error {
	// Verify code from Redis
	key := fmt.Sprintf("forget_password:%s", req.Email)
	storedCode, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return apperrors.NewInternalServerError(err)
		}
		return apperrors.NewVerifyCodeExpiredError()
	}
	if storedCode != req.Code {
		return apperrors.NewInvalidVerifyCodeError()
	}

	// Find user by email using repository
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	if user.Status == "suspended" {
		return apperrors.NewAccountSuspendedError()
	}

	// Update user's password using repository
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.NewInternalServerError(err)
	}

	if err := s.userRepo.UpdatePassword(ctx, uint(user.ID), string(hashedPassword)); err != nil {
		return err
	}

	// Delete verification code from Redis
	_ = database.RedisClient.Del(ctx, key).Err()

	return nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID uint64) (dto.UserProfileResponse, error) {
	// Find user by ID using repository
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return dto.UserProfileResponse{}, err
	}

	// Zero out password hash before returning
	user.PasswordHash = ""

	return dto.UserProfileResponse{
		ID:          int(user.ID),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		LastLogin:   "",
	}, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID uint64, req dto.UpdateUserProfile) (*model.User, error) {
	// Find user by ID using repository
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}

	if user.Status == "suspended" {
		return nil, apperrors.NewAccountSuspendedError()
	}

	// Update fields
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	// if req.AvatarURL != nil {
	// 	user.AvatarURL = *req.AvatarURL
	// }
	if req.Bio != nil {
		user.Bio = *req.Bio
	}

	user.UpdatedAt = time.Now()

	// Use repository to update
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Zero out password hash before returning
	user.PasswordHash = ""

	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint64, req dto.ChangePasswordRequest) error {
	// Find user by ID using repository
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return err
	}

	if user.Status == "suspended" {
		return apperrors.NewAccountSuspendedError()
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return apperrors.NewInvalidCredentialsError()
	}

	// Update to new password using repository
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.NewInternalServerError(err)
	}

	if err := s.userRepo.UpdatePassword(ctx, uint(userID), string(hashedPassword)); err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteAccount(ctx context.Context, userID uint64, req dto.DeleteAccountRequest) error {
	// Find user by ID using repository
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return err
	}

	if user.Status == "suspended" {
		return apperrors.NewAccountSuspendedError()
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return apperrors.NewInvalidCredentialsError()
	}

	// Use repository to soft delete
	if err := s.userRepo.SoftDelete(ctx, uint(userID)); err != nil {
		return err
	}

	return nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userID uint64, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", apperrors.NewInvalidFileTypeError("avatar")
	}

	avatarPath := fmt.Sprintf("uploads/avatars/user_%d_%d%s", userID, time.Now().Unix(), ext)

	// Save file to disk
	if err := utils.SaveUploadedFile(file, avatarPath); err != nil {
		return "", apperrors.NewInternalServerError(err)
	}

	// Update user's avatar URL using repository
	avatarURL := fmt.Sprintf("/%s", avatarPath) // Assuming static file server serves from root
	if err := s.userRepo.UpdateAvatarURL(ctx, uint(userID), avatarURL); err != nil {
		return "", err
	}

	return avatarURL, nil
}
