package service

import (
	"context"
	"linked-clone/internal/api/auth/dto"
	"linked-clone/internal/domain/entities"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) error
	ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)

	Logout(ctx context.Context, refreshToken string) error
	GetUserActiveSessions(ctx context.Context, userID uint, limit, offset int) ([]*entities.Session, error)
	RevokeSession(ctx context.Context, userID, sessionID uint) error
	RevokeAllUserSessions(ctx context.Context, userID uint) error
}
