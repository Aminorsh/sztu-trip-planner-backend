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
	} else {
		return nil, errors.New("user with this email or username already exists")
	}

	// Create new user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
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
	if err := s.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
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
	if err := s.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
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
