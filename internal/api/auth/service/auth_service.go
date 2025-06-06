package service

import (
	"context"
	"errors"
	"fmt"
	"linked-clone/internal/api/auth/dto"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"
	"linked-clone/pkg/auth"
	"linked-clone/pkg/logger"
	"linked-clone/pkg/redis"
	email "linked-clone/pkg/smtp"
	"linked-clone/pkg/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authService struct {
	userRepo     repositories.UserRepository
	jwtService   auth.JWTService
	emailService email.EmailService
	redisClient  redis.RedisClient
	logger       logger.Logger
}

func NewAuthService(
	userRepo repositories.UserRepository,
	jwtService auth.JWTService,
	emailService email.EmailService,
	redisClient redis.RedisClient,
	logger logger.Logger,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		jwtService:   jwtService,
		emailService: emailService,
		redisClient:  redisClient,
		logger:       logger,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("username already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, errors.New("failed to process password")
	}

	user := &entities.User{
		Email:    req.Email,
		Username: req.Username,
		FullName: req.FullName,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, errors.New("failed to create user")
	}

	verificationCode := utils.GenerateRandomCode(6)
	cacheKey := fmt.Sprintf("email_verification:%d", user.ID)

	if err := s.redisClient.Set(ctx, cacheKey, verificationCode, 15*time.Minute); err != nil {
		s.logger.Error("Failed to cache verification code", "error", err)
	} else {

		go func() {
			if err := s.emailService.SendVerificationEmail(user.Email, user.FullName, verificationCode); err != nil {
				s.logger.Error("Failed to send verification email", "error", err)
			}
		}()
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err)
		return nil, errors.New("failed to generate token")
	}

	return &dto.AuthResponse{
		User: &dto.UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			FullName:       user.FullName,
			ProfilePicture: user.ProfilePicture,
			Bio:            user.Bio,
			IsVerified:     user.IsVerified,
			IsPremium:      user.IsPremium,
		},
		Token:     token.AccessToken,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to authenticate")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err)
		return nil, errors.New("failed to generate token")
	}

	return &dto.AuthResponse{
		User: &dto.UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			FullName:       user.FullName,
			ProfilePicture: user.ProfilePicture,
			Bio:            user.Bio,
			IsVerified:     user.IsVerified,
			IsPremium:      user.IsPremium,
		},
		Token:     token.AccessToken,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) error {
	cacheKey := fmt.Sprintf("email_verification:%d", req.UserID)

	cachedCode, err := s.redisClient.Get(ctx, cacheKey)
	if err != nil {
		return errors.New("verification code expired or invalid")
	}

	if cachedCode != req.Code {
		return errors.New("invalid verification code")
	}

	if err := s.userRepo.VerifyEmail(ctx, req.UserID); err != nil {
		s.logger.Error("Failed to verify email", "error", err)
		return errors.New("failed to verify email")
	}

	s.redisClient.Delete(ctx, cacheKey)

	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil
		}
		s.logger.Error("Failed to get user", "error", err)
		return errors.New("failed to process request")
	}

	resetCode := utils.GenerateRandomCode(6)
	cacheKey := fmt.Sprintf("password_reset:%d", user.ID)

	if err := s.redisClient.Set(ctx, cacheKey, resetCode, 15*time.Minute); err != nil {
		s.logger.Error("Failed to cache reset code", "error", err)
		return errors.New("failed to process request")
	}

	go func() {
		if err := s.emailService.SendPasswordResetEmail(user.Email, user.FullName, resetCode); err != nil {
			s.logger.Error("Failed to send reset email", "error", err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid email")
		}
		s.logger.Error("Failed to get user", "error", err)
		return errors.New("failed to process request")
	}

	cacheKey := fmt.Sprintf("password_reset:%d", user.ID)

	cachedCode, err := s.redisClient.Get(ctx, cacheKey)
	if err != nil {
		return errors.New("reset code expired or invalid")
	}

	if cachedCode != req.Code {
		return errors.New("invalid reset code")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return errors.New("failed to process password")
	}

	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update password", "error", err)
		return errors.New("failed to update password")
	}

	s.redisClient.Delete(ctx, cacheKey)

	return nil
}

func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {

	claims, err := s.jwtService.ValidateToken(req.Token)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to get user", "error", err)
		return nil, errors.New("failed to refresh token")
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err)
		return nil, errors.New("failed to generate token")
	}

	return &dto.AuthResponse{
		User: &dto.UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			FullName:       user.FullName,
			ProfilePicture: user.ProfilePicture,
			Bio:            user.Bio,
			IsVerified:     user.IsVerified,
			IsPremium:      user.IsPremium,
		},
		Token:     token.AccessToken,
		ExpiresAt: token.ExpiresAt,
	}, nil
}
