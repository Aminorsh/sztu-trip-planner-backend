package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/database"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		DB: db,
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
	// Verify code from Redis
	key := fmt.Sprintf("verify:%s", req.Email)
	storedCode, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, errors.New("verification code expired or not found")
	}
	if storedCode != req.Code {
		return nil, errors.New("invalid verification code")
	}

	// Check if user already exists
	var existingUser model.User
	if err := s.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else if existingUser.Status != "inactive" {
		return nil, errors.New("user with this email or username already exists")
	}

	// Create new user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// If existing user is found but inactive, update that record instead of creating a new one
	if existingUser.Status == "inactive" {
		existingUser.Username = req.Username
		existingUser.PasswordHash = string(hashedPassword)
		existingUser.DisplayName = req.DisplayName
		existingUser.Status = "active"
		existingUser.UpdatedAt = time.Now()

		if err := s.DB.Save(&existingUser).Error; err != nil {
			return nil, err
		}

		// Delete verification code from Redis
		_ = database.RedisClient.Del(ctx, key).Err()

		// Zero out password hash before returning
		existingUser.PasswordHash = ""

		return &existingUser, nil
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

	if err := s.DB.Create(newUser).Error; err != nil {
		return nil, err
	}

	// Delete verification code from Redis
	_ = database.RedisClient.Del(ctx, key).Err()

	// Zero out password hash before returning
	newUser.PasswordHash = ""

	return newUser, nil
}

func (s *UserService) LoginUser(ctx context.Context, req dto.UsernameLogin) (*dto.UserLoginResponse, error) {
	// Find user by username
	var user model.User
	if err := s.DB.Where("username = ? AND status != ?", req.Username, "inactive").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	if user.Status == "suspended" {
		return nil, errors.New("account is suspended")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(uint64(user.ID))
	if err != nil {
		return nil, err
	}

	return &dto.UserLoginResponse{
		Token: token,
	}, nil
}

func (s *UserService) LoginUserByEmail(ctx context.Context, req dto.EmailLogin) (*dto.UserLoginResponse, error) {
	// Find user by email
	var user model.User
	if err := s.DB.Where("email = ? AND status != ?", req.Email, "inactive").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if user.Status == "suspended" {
		return nil, errors.New("account is suspended")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(uint64(user.ID))
	if err != nil {
		return nil, err
	}

	return &dto.UserLoginResponse{
		Token: token,
	}, nil
}

func (s *UserService) SendForgetPasswordCode(ctx context.Context, email string) error {
	// Check if user exists
	var user model.User
	if err := s.DB.Where("email = ? AND status != ?", email, "inactive").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user with this email does not exist")
		}
		return err
	}

	if user.Status == "suspended" {
		return errors.New("account is suspended")
	}

	// Generate verification code
	code := utils.GenerateVerificationCode()

	// Store code in Redis with expiration
	key := fmt.Sprintf("forget_password:%s", email)
	if err := database.RedisClient.Set(ctx, key, code, 15*time.Minute).Err(); err != nil {
		return err
	}

	// Send verification email
	if err := utils.NewEmailService().SendEmail(email, code, 1); err != nil {
		_ = database.RedisClient.Del(ctx, key).Err()
		return err
	}

	return nil
}

func (s *UserService) SendForgetPasswordCodeByUsername(ctx context.Context, username string) error {
	// Find user by username
	var user model.User
	if err := s.DB.Where("username = ? AND status != ?", username, "inactive").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user with this username does not exist")
		}
		return err
	}

	if user.Status == "suspended" {
		return errors.New("account is suspended")
	}

	// Generate verification code
	code := utils.GenerateVerificationCode()

	// Store code in Redis with expiration
	key := fmt.Sprintf("forget_password:%s", user.Email)
	if err := database.RedisClient.Set(ctx, key, code, 15*time.Minute).Err(); err != nil {
		return err
	}

	// Send verification email
	if err := utils.NewEmailService().SendEmail(user.Email, code, 1); err != nil {
		_ = database.RedisClient.Del(ctx, key).Err()
		return err
	}

	return nil
}

func (s *UserService) VerifyForgetPassword(ctx context.Context, req dto.ForgetPasswordVerifyRequest) error {
	// Verify code from Redis
	key := fmt.Sprintf("forget_password:%s", req.Email)
	storedCode, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return errors.New("verification code expired or not found")
	}
	if storedCode != req.Code {
		return errors.New("invalid verification code")
	}

	// Find user by email
	var user model.User
	if err := s.DB.Where("email = ? AND status != ?", req.Email, "inactive").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user with this email does not exist")
		}
		return err
	}

	if user.Status == "suspended" {
		return errors.New("account is suspended")
	}

	// Update user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	if err := s.DB.Save(&user).Error; err != nil {
		return err
	}

	// Delete verification code from Redis
	_ = database.RedisClient.Del(ctx, key).Err()

	return nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if user.Status == "inactive" {
		return nil, errors.New("user not found")
	}

	// Zero out password hash before returning
	user.PasswordHash = ""

	return &user, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID uint, req dto.UpdateUserProfile) (*model.User, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if user.Status == "inactive" {
		return nil, errors.New("user not found")
	}
	if user.Status == "suspended" {
		return nil, errors.New("account is suspended")
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}

	user.UpdatedAt = time.Now()

	if err := s.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	// Zero out password hash before returning
	user.PasswordHash = ""

	return &user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	if user.Status == "inactive" {
		return errors.New("user not found")
	}
	if user.Status == "suspended" {
		return errors.New("account is suspended")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Update to new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.DB.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteAccount(ctx context.Context, userID uint, req dto.DeleteAccountRequest) error {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	if user.Status == "inactive" {
		return errors.New("user not found")
	}
	if user.Status == "suspended" {
		return errors.New("account is suspended")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return errors.New("password is incorrect")
	}

	// Soft delete: set status to inactive
	user.Status = "inactive"
	user.UpdatedAt = time.Now()

	if err := s.DB.Save(&user).Error; err != nil {
		return err
	}

	return nil
}
